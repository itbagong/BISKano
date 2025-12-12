package ficologic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type CashBalanceOpt struct {
	CompanyID        string
	GroupByDimension []string
	AccountIDs       []string
	Dimension        tenantcoremodel.Dimension
}

func NewCashBalanceHub(db *datahub.Hub) *BalanceHub[*ficomodel.CashBalance, *ficomodel.CashTransaction, CashBalanceOpt] {
	calc := new(BalanceHub[*ficomodel.CashBalance, *ficomodel.CashTransaction, CashBalanceOpt])
	calc.db = db
	calc.provider = new(CashBalanceProvider)
	return calc
}

type CashBalanceProvider struct {
}

func (*CashBalanceProvider) Update(bals []*ficomodel.CashBalance, trxs []*ficomodel.CashTransaction,
	balDate *time.Time, deduct bool) ([]*ficomodel.CashBalance, error) {

	// convert bals ke map untuk mempermudah index
	mapBals := lo.SliceToMap(bals, func(bal *ficomodel.CashBalance) (string, *ficomodel.CashBalance) {
		return fmt.Sprintf("%s|%s", bal.CashBookID, bal.Dimension.Hash()), bal
	})

	for _, trx := range trxs {
		// get map of trx
		trxMapID := fmt.Sprintf("%s|%s", trx.CashBank.ID, trx.Dimension.Hash())
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(ficomodel.CashBalance)
			bal.CompanyID = trx.CompanyID
			bal.CashBookID = trx.CashBank.ID
			bal.Dimension = trx.Dimension
			mapBals[trxMapID] = bal
		}

		mul := float64(1)
		if deduct {
			mul = -1
		}
		switch trx.Status {
		case ficomodel.AmountConfirmed:
			bal.Balance += mul * trx.Amount
		case ficomodel.AmountReserved:
			bal.Reserved += mul * trx.Amount
		case ficomodel.AmountPlanned:
			bal.Planned += mul * trx.Amount
		}
	}
	res := lo.MapToSlice(mapBals, func(_ string, v *ficomodel.CashBalance) *ficomodel.CashBalance {
		v.Calc()
		return v
	})

	return res, nil
}

func (*CashBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom, dateTo *time.Time,
	initialWhere *dbflex.Filter,
	includeDateTo bool, opt CashBalanceOpt) ([]*ficomodel.CashTransaction, error) {

	trxWheres := []*dbflex.Filter{initialWhere}
	if dateFrom != nil {
		trxWheres = append(trxWheres, dbflex.Gt("TrxDate", dateFrom))
	}
	if dateTo != nil {
		if includeDateTo {
			trxWheres = append(trxWheres, dbflex.Lte("TrxDate", dateTo))
		} else {
			trxWheres = append(trxWheres, dbflex.Lt("TrxDate", dateTo))
		}
	}

	res, err := datahub.FindByFilter(db, new(ficomodel.CashTransaction), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (*CashBalanceProvider) BalanceFilter(obj *ficomodel.CashBalance, opt CashBalanceOpt) *dbflex.Filter {
	if opt.CompanyID == "" {
		return nil
	}

	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("CashBookID", opt.AccountIDs...))
	}
	return dbflex.And(wheres...)
}

func (*CashBalanceProvider) TransactionFilter(obj *ficomodel.CashTransaction, opt CashBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("CashBank._id", opt.AccountIDs...))
	}
	if len(opt.Dimension) > 0 {
		wheres = append(wheres, opt.Dimension.Where())
	}
	if obj.CompanyID != "" {
		wheres = append(wheres, dbflex.Eq("CompanyID", obj.CompanyID))
	}
	if obj.CashBank.ID != "" {
		wheres = append(wheres, dbflex.Eq("CashBank._id", obj.CashBank.ID))
	}
	if len(opt.GroupByDimension) > 0 {
		gd := tenantcoremodel.Dimension{}
		for _, d := range opt.GroupByDimension {
			gd.Set(d, obj.Dimension.Get(d))
		}
		wheres = append(wheres, gd.Where())
	}
	return dbflex.And(wheres...)
}

func (*CashBalanceProvider) AggregateBalance(origModel *ficomodel.CashBalance, aggrModels []*ficomodel.CashBalance, add bool, opt CashBalanceOpt) error {
	if len(aggrModels) > 0 {
		if origModel.CompanyID == "" {
			origModel.CompanyID = aggrModels[0].CompanyID
		}
		origModel.CashBookID = aggrModels[0].CashBookID
		if len(opt.GroupByDimension) == 0 {
			origModel.Dimension = tenantcorelogic.TernaryDimension(origModel.Dimension, aggrModels[0].Dimension)
		} else {
			origModel.Dimension = aggrModels[0].Dimension.Sub(opt.GroupByDimension...)
		}
		origModel.Balance = 0
		origModel.Planned = 0
		origModel.Reserved = 0
		origModel.Available = 0
	}

	for _, item := range aggrModels {
		mul := float64(1)
		if !add {
			mul = -1
		}
		origModel.Balance += mul * item.Balance
		origModel.Planned += mul * item.Planned
		origModel.Reserved += mul * item.Reserved
	}
	origModel.Calc()
	return nil
}

func (*CashBalanceProvider) BalanceGrouping(obj *ficomodel.CashBalance, opt CashBalanceOpt) string {
	dimGroup := getDimensionHash(obj.Dimension, opt.GroupByDimension...)
	return fmt.Sprintf("%s|%s", obj.CashBookID, dimGroup)
}

type CashBalanceHandler struct {
}

type GetCurrentRequest struct {
	CashBankID string
}

func (m *CashBalanceHandler) GetCurrent(ctx *kaos.Context, payload *GetCurrentRequest) (*ficomodel.CashBalance, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	cashBalance, err := datahub.GetByFilter(h, new(ficomodel.CashBalance), dbflex.And(
		dbflex.Eq("CashBookID", payload.CashBankID),
		dbflex.Eq("BalanceDate", nil),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get cash balance %s: %s", payload.CashBankID, err.Error())
	}

	return cashBalance, nil
}
