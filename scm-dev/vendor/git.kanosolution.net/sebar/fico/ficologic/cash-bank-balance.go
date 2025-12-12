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

type CashBankBalanceHandler struct {
}

type CashBankBalanceRequest struct {
	ID       []string
	DateFrom *time.Time
	DateTo   *time.Time
	TrxType  string
}

type CashBankBalanceGetsRequest struct {
	Where struct {
		CashBankBalanceRequest
		Dimension []string
		Items     []*dbflex.Filter
		Op        dbflex.FilterOp
	}
	Sort []string
	Skip int
	Take int
}

func (m *CashBankBalanceHandler) Gets(ctx *kaos.Context, payload *CashBankBalanceGetsRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{}
	if len(payload.Where.ID) > 0 {
		filters = append(filters, dbflex.In("_id", payload.Where.ID...))
	}

	// filter dimension site
	jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})

	if len(dimIface) > 0 {
		filters = append(filters, GetFilterDimension(dimIface)...)
	}

	if len(payload.Where.Items) > 0 {
		filters2 := make([]*dbflex.Filter, len(payload.Where.Items))
		for _, item := range payload.Where.Items {
			if len(item.Items) > 0 && item.Op == dbflex.OpOr {
				filters3 := GetFilterItems(item.Items)
				filters2 = append(filters2, dbflex.Or(filters3...))
			} else if item.Op == dbflex.OpIn {
				aInterface := item.Value.([]interface{})
				aString := make([]string, len(aInterface))
				for i, v := range aInterface {
					aString[i] = v.(string)
				}
				filters2 = append(filters2, dbflex.In(item.Field, aString...))
			}
		}
		filters = append(filters, dbflex.And(filters2...))
	}

	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	// get count
	count, err := h.Count(new(tenantcoremodel.CashBank), query)
	if err != nil {
		return nil, fmt.Errorf("error when get count cash bank: %s", err.Error())
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

	// get cash bank
	cashBanks := []tenantcoremodel.CashBank{}
	err = h.Gets(new(tenantcoremodel.CashBank), query, &cashBanks)
	if err != nil {
		return nil, fmt.Errorf("error when get cash bank: %s", err.Error())
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

	balance := NewCashBalanceHub(h)
	balances, err := balance.Get(payload.Where.DateFrom, CashBalanceOpt{
		AccountIDs:       payload.Where.ID,
		GroupByDimension: payload.Where.Dimension,
		CompanyID:        coID,
	})
	if err != nil {
		return nil, fmt.Errorf("error when get cash bank balance: %s", err.Error())
	}

	mapBalance := map[string][]*ficomodel.CashBalance{}
	for _, d := range balances {
		mapBalance[d.CashBookID] = append(mapBalance[d.CashBookID], d)
	}

	result := []codekit.M{}
	for _, cashBank := range cashBanks {
		if values, ok := mapBalance[cashBank.ID]; ok {
			for _, b := range values {
				cb, _ := codekit.ToM(cashBank)
				for _, d := range payload.Where.Dimension {
					cb[d] = mapDimension[b.Dimension.Get(d)]
				}
				cb["Balance"] = b.Balance
				cb["Reserved"] = b.Reserved
				cb["Planned"] = b.Planned
				cb["Available"] = b.Available
				result = append(result, cb)
			}
		} else {
			cb, _ := codekit.ToM(cashBank)
			cb["Balance"] = 0
			cb["Reserved"] = 0
			cb["Planned"] = 0
			cb["Available"] = 0
			result = append(result, cb)
		}
	}

	return codekit.M{"count": count, "data": result}, nil
}

type CashBankBalanceTransactionRequest struct {
	CashBankBalanceRequest
	DimFinance map[string][]string
}

func (m *CashBankBalanceHandler) GetTransaction(ctx *kaos.Context, payload *CashBankBalanceTransactionRequest) (codekit.M, error) {
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
		dbflex.Eq("Status", ficomodel.AmountConfirmed),
	}

	if len(payload.ID) > 0 {
		query = append(query, dbflex.In("CashBank._id", payload.ID...))
	}

	if payload.DateTo != nil {
		date := payload.DateTo.AddDate(0, 0, 1)
		payload.DateTo = &date
	}

	balance := NewCashBalanceHub(h)
	opt := CashBalanceOpt{
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

	openingBalance := new(ficomodel.CashBalance)
	closingBalance := new(ficomodel.CashBalance)
	if len(openingBalances) > 0 {
		openingBalance = openingBalances[0]
	}
	if len(closingBalances) > 0 {
		closingBalance = closingBalances[0]
	}

	var startTransaction *time.Time
	// if date filter empty, get latest closing date in cash balance
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
	if payload.TrxType != "" {
		query = append(query, dbflex.Eq("TrxType", payload.TrxType))
	}
	if len(query) > 0 {
		param = param.SetWhere(dbflex.And(query...))
	}

	trxs := []codekit.M{}

	err = h.Gets(new(ficomodel.CashTransaction), param.SetSort("TrxDate"), &trxs)
	if err != nil {
		return nil, fmt.Errorf("error when get cash transaction: %s", err.Error())
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
