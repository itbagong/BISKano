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

type LedgerBalanceCalc struct {
	db *datahub.Hub
	//ev        kaos.EventHub
	dimNames  []string
	companyID string
}

func NewLedgerBalanceCalc(h *datahub.Hub, companyID string, dimNames ...string) *LedgerBalanceCalc {
	lb := &LedgerBalanceCalc{}
	lb.db = h
	//lb.ev = ev
	lb.dimNames = dimNames
	lb.companyID = companyID
	return lb
}

type LedgerBalanceGetOpts struct {
	LedgerAccounts []string
	DimNames       []string
}

func (o *LedgerBalanceCalc) Get(balanceDate *time.Time, parm LedgerBalanceGetOpts) []*ficomodel.LedgerBalance {
	// build filter to create balance to nearest snapshot
	var upperDate *time.Time
	filters := []*dbflex.Filter{dbflex.Eq("CompanyID", o.companyID)}
	if balanceDate == nil {
		filters = append(filters, dbflex.Eq("BalanceDate", balanceDate))
	} else {
		dateFilter := dbflex.And(dbflex.Eq("CompanyID", o.companyID), dbflex.Gte("BalanceDate", balanceDate))
		dateRecord, err := datahub.GetByParm(o.db, new(ficomodel.LedgerBalance), dbflex.NewQueryParam().SetWhere(dateFilter).SetSort("-BalanceDate"))
		if err != nil {
			filters = append(filters, dbflex.Eq("BalanceDate", nil))
		} else {
			filters = append(filters, dbflex.Eq("BalanceDate", dateRecord.BalanceDate))
		}
		upperDate = dateRecord.BalanceDate
	}
	if len(parm.LedgerAccounts) > 0 {
		filters = append(filters, dbflex.In("LedgerAccountID", lo.Map(parm.LedgerAccounts, func(s string, i int) interface{} {
			return s
		})...))
	}
	filter := dbflex.And(filters...)

	// get the balance
	balances, _ := datahub.FindByFilter(o.db, &ficomodel.LedgerBalance{}, filter)
	// find trx gt expected date and deduct the value
	if balanceDate != nil && balanceDate.Before(*upperDate) {
		for _, balance := range balances {
			balSub := GetLedgerBalanceByTrxSum(o.db, balance.CompanyID, balance.LedgerAccountID, balance.Dimension, *balanceDate, upperDate)
			balance.Balance -= balSub.Balance
		}
	}
	// grouped by dim
	groupBalances := lo.GroupBy(balances, func(b *ficomodel.LedgerBalance) string {
		groupers := make([]string, len(parm.DimNames)+1)
		groupers[0] = b.LedgerAccountID
		dimMap := b.Dimension.ToMap()
		for index, name := range parm.DimNames {
			groupers[index+1] = dimMap[name]
		}
		return strings.Join(groupers, "|")
	})

	return lo.MapToSlice(groupBalances, func(_ string, bals []*ficomodel.LedgerBalance) *ficomodel.LedgerBalance {
		res := new(ficomodel.LedgerBalance)
		res.CompanyID = bals[0].CompanyID
		res.LedgerAccountID = bals[0].LedgerAccountID
		res.Dimension = tenantcoremodel.Dimension{}
		balDim := bals[0].Dimension.ToMap()
		for _, dimName := range parm.DimNames {
			res.Dimension = res.Dimension.Set(dimName, balDim[dimName])
		}
		res.Balance += lo.SumBy(bals, func(b *ficomodel.LedgerBalance) float64 {
			return b.Balance
		})
		return res
	})
}

func (o *LedgerBalanceCalc) Update(b *ficomodel.LedgerBalance) (*ficomodel.LedgerBalance, error) {
	filter := dbflex.And(
		dbflex.Eqs("CompanyID", o.companyID, "BalanceDate", nil, "LedgerAccountID", b.LedgerAccountID),
		b.Dimension.Where())
	res, err := datahub.GetByFilter(o.db, new(ficomodel.LedgerBalance), filter)
	if err != nil {
		res = new(ficomodel.LedgerBalance)
		res.CompanyID = o.companyID
		res.LedgerAccountID = b.LedgerAccountID
		res.BalanceDate = nil
		res.Dimension = b.Dimension
	}
	res.Balance += b.Balance
	if err := o.db.Save(res); err != nil {
		return res, fmt.Errorf("ledger balance: %s", err.Error())
	}

	return res, nil
}

func GetLedgerBalanceByTrxSum(db *datahub.Hub, companyID string, id string, dim tenantcoremodel.Dimension, dateFrom time.Time, dateTo *time.Time) *ficomodel.LedgerBalance {
	bal := new(ficomodel.LedgerBalance)
	filters := []*dbflex.Filter{
		dbflex.Eqs("CompanyID", companyID, "LedgerAccountID", id),
		dbflex.Gt("TrxDate", dateFrom)}
	if len(dim) > 0 {
		filters = append(filters, dim.Where())
	}
	if dateTo != nil {
		filters = append(filters, dbflex.Lte("TrxDate", dateTo))
	}
	trxs, _ := datahub.FindAnyByParm(db, new(ficomodel.LedgerTransaction), new(ficomodel.LedgerTransaction).TableName(),
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
	return bal
}
