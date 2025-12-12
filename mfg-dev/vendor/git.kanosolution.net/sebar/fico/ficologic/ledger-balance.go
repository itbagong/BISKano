package ficologic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type LedgerBalanceOpt struct {
	CompanyID        string
	GroupByDimension []string
	Dimension        tenantcoremodel.Dimension
	AccountIDs       []string
}

type LedgerBalanceProvider struct {
}

func NewLedgerBalanceHub(db *datahub.Hub) *BalanceHub[*ficomodel.LedgerBalance, *ficomodel.LedgerTransaction, LedgerBalanceOpt] {
	calc := new(BalanceHub[*ficomodel.LedgerBalance, *ficomodel.LedgerTransaction, LedgerBalanceOpt])
	calc.db = db
	calc.provider = new(LedgerBalanceProvider)
	return calc
}

func (o *LedgerBalanceProvider) BalanceFilter(obj *ficomodel.LedgerBalance, opt LedgerBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("LedgerAccountID", opt.AccountIDs...))
	}
	if obj != nil {
		if obj.LedgerAccountID != "" {
			wheres = append(wheres, dbflex.Eq("LedgerAccountID", obj.LedgerAccountID))
		}
		if len(opt.GroupByDimension) > 0 {
			dim := tenantcoremodel.Dimension{}
			for _, d := range opt.GroupByDimension {
				dv := obj.Dimension.Get(d)
				if dv != "" {
					dim.Set(d, dv)
				}
			}
			if len(dim) > 0 {
				wheres = append(wheres, dim.Where())
			}
		}
	}
	return dbflex.And(wheres...)
}

func (o *LedgerBalanceProvider) TransactionFilter(obj *ficomodel.LedgerTransaction, opt LedgerBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("Account._id", opt.AccountIDs...))
	}
	if len(opt.Dimension) > 0 {
		wheres = append(wheres, opt.Dimension.Where())
	}
	if obj != nil {
		if obj.CompanyID != "" {
			wheres = append(wheres, dbflex.Eq("CompanyID", obj.CompanyID))
		}
		if obj.Account.ID != "" {
			wheres = append(wheres, dbflex.Eq("Account._id", obj.Account.ID))
		}
		if len(opt.GroupByDimension) > 0 {
			dim := tenantcoremodel.Dimension{}
			for _, d := range opt.GroupByDimension {
				dv := obj.Dimension.Get(d)
				if dv != "" {
					dim.Set(d, dv)
				}
			}
			if len(dim) > 0 {
				wheres = append(wheres, dim.Where())
			}
		}
	}
	return dbflex.And(wheres...)
}

func (o *LedgerBalanceProvider) BalanceGrouping(obj *ficomodel.LedgerBalance, opt LedgerBalanceOpt) string {
	dimGroup := getDimensionHash(obj.Dimension, opt.GroupByDimension...)
	return fmt.Sprintf("%s|%s", obj.LedgerAccountID, dimGroup)
}

func (o *LedgerBalanceProvider) AggregateBalance(origModel *ficomodel.LedgerBalance, balModels []*ficomodel.LedgerBalance, add bool, opt LedgerBalanceOpt) error {
	if len(balModels) > 0 {
		if origModel.CompanyID == "" {
			origModel.CompanyID = balModels[0].CompanyID
		}
		origModel.LedgerAccountID = balModels[0].LedgerAccountID
		if len(opt.GroupByDimension) == 0 {
			origModel.Dimension = tenantcorelogic.TernaryDimension(origModel.Dimension, balModels[0].Dimension)
		} else {
			origModel.Dimension = balModels[0].Dimension.Sub(opt.GroupByDimension...)
		}
		origModel.BalanceDate = balModels[0].BalanceDate
		origModel.Balance = 0
	}

	for _, item := range balModels {
		mul := float64(1)
		if !add {
			mul = -1
		}
		origModel.Balance += mul * item.Balance
	}
	return nil
}

func (o *LedgerBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom *time.Time, dateTo *time.Time, initialWhere *dbflex.Filter, includeDateTo bool, opt LedgerBalanceOpt) ([]*ficomodel.LedgerTransaction, error) {
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

	res, err := datahub.FindByFilter(db, new(ficomodel.LedgerTransaction), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *LedgerBalanceProvider) Update(bals []*ficomodel.LedgerBalance, trxs []*ficomodel.LedgerTransaction, balDate *time.Time, deduct bool) ([]*ficomodel.LedgerBalance, error) {
	// convert bals ke map untuk mempermudah index
	mapBals := lo.SliceToMap(bals, func(bal *ficomodel.LedgerBalance) (string, *ficomodel.LedgerBalance) {
		return fmt.Sprintf("%s|%s", bal.LedgerAccountID, bal.Dimension.Hash()), bal
	})

	for _, trx := range trxs {
		// get map of trx
		trxMapID := fmt.Sprintf("%s|%s", trx.Account.ID, trx.Dimension.Hash())
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(ficomodel.LedgerBalance)
			bal.CompanyID = trx.CompanyID
			bal.LedgerAccountID = trx.Account.ID
			bal.Dimension = trx.Dimension
			mapBals[trxMapID] = bal
		}

		if !deduct {
			bal.Balance += trx.Amount
		} else {
			bal.Balance -= trx.Amount
		}
	}

	res := lo.MapToSlice(mapBals, func(_ string, v *ficomodel.LedgerBalance) *ficomodel.LedgerBalance {
		return v
	})

	return res, nil
}
