package ficologic

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type VendorBalanceCalc struct {
	db *datahub.Hub
	//ev        kaos.EventHub
	dimNames  []string
	companyID string
}

func NewVendorBalanceCalc(h *datahub.Hub, companyID string, dimNames ...string) *VendorBalanceCalc {
	lb := &VendorBalanceCalc{}
	lb.db = h
	//lb.ev = ev
	lb.dimNames = dimNames
	lb.companyID = companyID
	return lb
}

type VendorBalanceGetOpts struct {
	VendorIDs []string
	DimNames  []string
}

func (o *VendorBalanceCalc) Get(balanceDate *time.Time, parm *VendorBalanceGetOpts) []*ficomodel.VendorBalance {
	if parm == nil {
		parm = new(VendorBalanceGetOpts)
	}
	filters := []*dbflex.Filter{dbflex.Eq("CompanyID", o.companyID)}
	if balanceDate == nil {
		filters = append(filters, dbflex.Eq("BalanceDate", balanceDate))
	} else {
		dateFilter := dbflex.And(dbflex.Eq("CompanyID", o.companyID), dbflex.Gte("BalanceDate", balanceDate))
		dateRecord, err := datahub.GetByParm(o.db, new(ficomodel.VendorBalance), dbflex.NewQueryParam().SetWhere(dateFilter).SetSort("-BalanceDate"))
		if err != nil {
			filters = append(filters, dbflex.Eq("BalanceDate", nil))
		} else {
			filters = append(filters, dbflex.Eq("BalanceDate", dateRecord.BalanceDate))
		}
	}
	if len(parm.VendorIDs) > 0 {
		filters = append(filters, dbflex.In("VendorID", lo.Map(parm.VendorIDs, func(s string, i int) interface{} {
			return s
		})...))
	}
	filter := dbflex.And(filters...)

	balances, _ := datahub.FindByFilter(o.db, &ficomodel.VendorBalance{}, filter)
	groupBalances := lo.GroupBy(balances, func(b *ficomodel.VendorBalance) string {
		groupers := make([]string, len(parm.DimNames)+1)
		groupers[0] = b.VendorID
		dimMap := b.Dimension.ToMap()
		for index, name := range parm.DimNames {
			groupers[index+1] = dimMap[name]
		}
		return strings.Join(groupers, "|")
	})
	return lo.MapToSlice(groupBalances, func(_ string, bals []*ficomodel.VendorBalance) *ficomodel.VendorBalance {
		res := new(ficomodel.VendorBalance)
		res.CompanyID = bals[0].CompanyID
		res.VendorID = bals[0].VendorID
		res.Dimension = tenantcoremodel.Dimension{}
		balDim := bals[0].Dimension.ToMap()
		for _, dimName := range parm.DimNames {
			res.Dimension = res.Dimension.Set(dimName, balDim[dimName])
		}
		lo.ForEach(bals, func(item *ficomodel.VendorBalance, index int) {
			res.Balance += item.Balance
		})
		return res
	})
}

func (o *VendorBalanceCalc) Update(b *ficomodel.VendorBalance) (*ficomodel.VendorBalance, error) {
	filter := dbflex.And(
		dbflex.Eqs("CompanyID", o.companyID, "BalanceDate", nil, "VendorID", b.VendorID),
		b.Dimension.Where())
	res, err := datahub.GetByFilter(o.db, new(ficomodel.VendorBalance), filter)
	if err != nil {
		res = new(ficomodel.VendorBalance)
		res.CompanyID = o.companyID
		res.VendorID = b.VendorID
		res.BalanceDate = nil
		res.Dimension = b.Dimension
	}
	res.Balance += b.Balance
	if err := o.db.Save(res); err != nil {
		return res, fmt.Errorf("ledger balance: %s", err.Error())
	}

	return res, nil
}

func GetVendorBalanceByTrxSum(db *datahub.Hub, companyID string, id string, dim tenantcoremodel.Dimension,
	dateFrom time.Time, dateTo *time.Time) *ficomodel.VendorBalance {
	bal := new(ficomodel.VendorBalance)
	filters := []*dbflex.Filter{
		dbflex.Eqs("CompanyID", companyID, "VendorID", id),
		dbflex.Gt("TrxDate", dateFrom)}
	if len(dim) > 0 {
		filters = append(filters, dim.Where())
	}
	if dateTo != nil {
		filters = append(filters, dbflex.Lte("TrxDate", dateTo))
	}
	trxs, _ := datahub.FindAnyByParm(db, new(ficomodel.LedgerTransaction), new(ficomodel.VendorTransaction).TableName(),
		dbflex.NewQueryParam().
			SetWhere(dbflex.And(filters...)).
			SetAggr(dbflex.NewAggrItem("Amount", dbflex.AggrSum, "Amount")).
			SetGroupBy("Status"))
	for _, trx := range trxs {
		switch trx.Status {
		case ficomodel.AmountConfirmed:
			bal.Balance -= trx.Amount
		}
	}
	bal.Calc()
	return bal
}
