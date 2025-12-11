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

type CashBalanceCalc struct {
	db *datahub.Hub
	//ev        kaos.EventHub
	dimNames  []string
	companyID string
}

func NewCashBalanceCalc(h *datahub.Hub, companyID string, dimNames ...string) *CashBalanceCalc {
	lb := &CashBalanceCalc{}
	lb.db = h
	//lb.ev = ev
	lb.dimNames = dimNames
	lb.companyID = companyID
	return lb
}

type CashBalanceGetOpts struct {
	CashBookIDs []string
	DimNames    []string
}

func (o *CashBalanceCalc) Get(balanceDate *time.Time, parm *CashBalanceGetOpts) []*ficomodel.CashBalance {
	if parm == nil {
		parm = new(CashBalanceGetOpts)
	}
	filters := []*dbflex.Filter{dbflex.Eq("CompanyID", o.companyID)}
	if balanceDate == nil {
		filters = append(filters, dbflex.Eq("BalanceDate", balanceDate))
	} else {
		dateFilter := dbflex.And(dbflex.Eq("CompanyID", o.companyID), dbflex.Gte("BalanceDate", balanceDate))
		dateRecord, err := datahub.GetByParm(o.db, new(ficomodel.CashBalance), dbflex.NewQueryParam().SetWhere(dateFilter).SetSort("-BalanceDate"))
		if err != nil {
			filters = append(filters, dbflex.Eq("BalanceDate", nil))
		} else {
			filters = append(filters, dbflex.Eq("BalanceDate", dateRecord.BalanceDate))
		}
	}
	if len(parm.CashBookIDs) > 0 {
		filters = append(filters, dbflex.In("CashBookID", lo.Map(parm.CashBookIDs, func(s string, i int) interface{} {
			return s
		})...))
	}
	filter := dbflex.And(filters...)

	balances, _ := datahub.FindByFilter(o.db, &ficomodel.CashBalance{}, filter)
	groupBalances := lo.GroupBy(balances, func(b *ficomodel.CashBalance) string {
		groupers := make([]string, len(parm.DimNames)+1)
		groupers[0] = b.CashBookID
		dimMap := b.Dimension.ToMap()
		for index, name := range parm.DimNames {
			groupers[index+1] = dimMap[name]
		}
		return strings.Join(groupers, "|")
	})
	return lo.MapToSlice(groupBalances, func(_ string, bals []*ficomodel.CashBalance) *ficomodel.CashBalance {
		res := new(ficomodel.CashBalance)
		res.CompanyID = bals[0].CompanyID
		res.CashBookID = bals[0].CashBookID
		res.Dimension = tenantcoremodel.Dimension{}
		balDim := bals[0].Dimension.ToMap()
		for _, dimName := range parm.DimNames {
			res.Dimension = res.Dimension.Set(dimName, balDim[dimName])
		}
		lo.ForEach(bals, func(item *ficomodel.CashBalance, index int) {
			res.Balance += item.Balance
			res.Reserved += item.Reserved
			res.Planned += item.Planned
		})
		res.Available = res.Balance + res.Planned - res.Reserved
		return res
	})
}

func (o *CashBalanceCalc) Update(b *ficomodel.CashBalance) (*ficomodel.CashBalance, error) {
	filter := dbflex.And(
		dbflex.Eqs("CompanyID", o.companyID, "BalanceDate", nil, "CashBookID", b.CashBookID),
		b.Dimension.Where())
	res, err := datahub.GetByFilter(o.db, new(ficomodel.CashBalance), filter)
	if err != nil {
		res = new(ficomodel.CashBalance)
		res.CompanyID = o.companyID
		res.CashBookID = b.CashBookID
		res.BalanceDate = nil
		res.Dimension = b.Dimension
	}
	res.Balance += b.Balance
	res.Reserved += b.Reserved
	res.Planned += b.Planned
	res.Available = res.Balance + res.Planned - res.Reserved
	if err := o.db.Save(res); err != nil {
		return res, fmt.Errorf("ledger balance: %s", err.Error())
	}

	return res, nil
}

func GetCashBalanceByTrxSum(db *datahub.Hub, companyID string, id string, dim tenantcoremodel.Dimension,
	dateFrom time.Time, dateTo *time.Time) *ficomodel.CashBalance {
	bal := new(ficomodel.CashBalance)
	filters := []*dbflex.Filter{
		dbflex.Eqs("CompanyID", companyID, "CashBookID", id),
		dbflex.Gt("TrxDate", dateFrom)}
	if len(dim) > 0 {
		filters = append(filters, dim.Where())
	}
	if dateTo != nil {
		filters = append(filters, dbflex.Lte("TrxDate", dateTo))
	}
	trxs, _ := datahub.FindAnyByParm(db, new(ficomodel.LedgerTransaction), new(ficomodel.CashTransaction).TableName(),
		dbflex.NewQueryParam().
			SetWhere(dbflex.And(filters...)).
			SetAggr(dbflex.NewAggrItem("Amount", dbflex.AggrSum, "Amount")).
			SetGroupBy("Status"))
	for _, trx := range trxs {
		switch trx.Status {
		case ficomodel.AmountConfirmed:
			bal.Balance -= trx.Amount

		case ficomodel.AmountReserved:
			bal.Reserved -= trx.Amount

		case ficomodel.AmountPlanned:
			bal.Planned -= trx.Amount
		}
	}
	bal.Calc()
	return bal
}
