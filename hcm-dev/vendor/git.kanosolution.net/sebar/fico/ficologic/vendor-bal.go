package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type VendorBalanceHandler struct {
}

type VendorBalanceRequest struct {
	ID       []string
	DateFrom *time.Time
	DateTo   *time.Time
}

type VendorBalanceGetsRequest struct {
	Where struct {
		VendorBalanceRequest
		Dimension []string
		Items     []*dbflex.Filter
		Op        dbflex.FilterOp
	}
	Skip int
	Take int
	Sort []string
}

func (m *VendorBalanceHandler) Gets(ctx *kaos.Context, payload *VendorBalanceGetsRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{}
	if len(payload.Where.ID) > 0 {
		filters = append(filters, dbflex.In("_id", payload.Where.ID...))
	}

	filters2 := GetFilterItems(payload.Where.Items)
	if len(filters2) > 0 {
		filters = append(filters, dbflex.Or(filters2...))
	}

	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	// get count
	count, err := h.Count(new(tenantcoremodel.Vendor), query)
	if err != nil {
		return nil, fmt.Errorf("error when get count vendor: %s", err.Error())
	}

	if payload.Skip != 0 {
		query = query.SetSkip(payload.Skip)
	}

	if payload.Take != 0 {
		query = query.SetTake(payload.Take)
	}

	if len(payload.Sort) > 0 {
		query = query.SetSort(payload.Sort...)
	}

	// get vendor
	vendors := []tenantcoremodel.Vendor{}
	err = h.Gets(new(tenantcoremodel.Vendor), query, &vendors)
	if err != nil {
		return nil, fmt.Errorf("error when get vendor: %s", err.Error())
	}

	mapDimension := map[string]string{}
	if len(payload.Where.Dimension) > 0 {
		// get dimension
		dimensions := []tenantcoremodel.DimensionMaster{}
		err = h.Gets(new(tenantcoremodel.DimensionMaster), dbflex.NewQueryParam().SetWhere(
			dbflex.In("DimensionType", payload.Where.Dimension...),
		), &dimensions)
		if err != nil {
			return nil, fmt.Errorf("error when dimension: %s", err.Error())
		}

		mapDimension = lo.Associate(dimensions, func(dim tenantcoremodel.DimensionMaster) (string, string) {
			return dim.ID, dim.Label
		})
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	balance := NewVendorBalanceHub(h)
	balances, err := balance.Get(payload.Where.DateFrom, VendorBalanceOpt{
		AccountIDs:       payload.Where.ID,
		GroupByDimension: payload.Where.Dimension,
		CompanyID:        coID,
	})
	if err != nil {
		return nil, fmt.Errorf("error when get vendor balance: %s", err.Error())
	}

	mapBalance := map[string][]*ficomodel.VendorBalance{}
	for _, d := range balances {
		mapBalance[d.VendorID] = append(mapBalance[d.VendorID], d)
	}

	result := []codekit.M{}
	for _, vendor := range vendors {
		if values, ok := mapBalance[vendor.ID]; ok {
			for _, b := range values {
				ven, _ := codekit.ToM(vendor)
				for _, d := range payload.Where.Dimension {
					ven[d] = mapDimension[b.Dimension.Get(d)]
				}
				ven["Balance"] = b.Balance
				result = append(result, ven)
			}
		} else {
			ven, _ := codekit.ToM(vendor)
			ven["Balance"] = 0
			result = append(result, ven)
		}
	}

	return codekit.M{"count": count, "data": result}, nil
}

type VendorBalanceTransactionRequest struct {
	VendorBalanceRequest
	DimFinance map[string][]string
}

func (m *VendorBalanceHandler) GetTransaction(ctx *kaos.Context, payload *VendorBalanceTransactionRequest) (codekit.M, error) {
	var err error
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	param := dbflex.NewQueryParam()
	query := []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
	}

	if len(payload.ID) > 0 {
		query = append(query, dbflex.In("Vendor._id", payload.ID...))
	}

	if payload.DateTo != nil {
		date := payload.DateTo.AddDate(0, 0, 1)
		payload.DateTo = &date
	}

	balance := NewVendorBalanceHub(h)
	opt := VendorBalanceOpt{
		AccountIDs: payload.ID,
		CompanyID:  coID,
	}

	if len(payload.DimFinance) > 0 {
		for key, values := range payload.DimFinance {
			for _, v := range values {
				opt.Dimension = opt.Dimension.Set(key, v)
			}
		}

		query = append(query, opt.Dimension.Where())
	}

	openingBalances, err := balance.Get(payload.DateFrom, opt)
	if err != nil {
		return nil, fmt.Errorf("error when get opening balance: %s", err.Error())
	}
	closingBalances, err := balance.Get(payload.DateTo, opt)
	if err != nil {
		return nil, fmt.Errorf("error when get closing balance: %s", err.Error())
	}

	openingBalance := new(ficomodel.VendorBalance)
	closingBalance := new(ficomodel.VendorBalance)
	if len(openingBalances) > 0 {
		openingBalance = openingBalances[0]
	}
	if len(closingBalances) > 0 {
		closingBalance = closingBalances[0]
	}

	var startTransaction *time.Time
	// if date filter empty, get latest closing date in customer balance
	if payload.DateFrom == nil {
		startTransaction = openingBalance.BalanceDate
	} else {
		startTransaction = payload.DateFrom
	}

	if startTransaction != nil {
		query = append(query, dbflex.Gte("TrxDate", startTransaction))
	}

	if payload.DateTo != nil {
		query = append(query, dbflex.Lt("TrxDate", payload.DateTo))
	}

	if len(query) > 0 {
		param = param.SetWhere(dbflex.And(query...))
	}

	trxs := []codekit.M{}
	err = h.Gets(new(ficomodel.VendorTransaction), param.SetSort("TrxDate"), &trxs)
	if err != nil {
		return nil, fmt.Errorf("error when get vendor transaction: %s", err.Error())
	}

	transactionTotal := 0.0
	for _, tr := range trxs {
		transactionTotal += tr["Amount"].(float64)
		tr["Balance"] = openingBalance.Balance + transactionTotal
	}

	res := codekit.M{
		"Transaction": trxs,
		"Opening": codekit.M{
			"Date":    openingBalance.BalanceDate,
			"Balance": openingBalance.Balance,
		},
		"Closing": codekit.M{
			"Date":    closingBalance.BalanceDate,
			"Balance": closingBalance.Balance,
		},
	}

	return res, nil
}
