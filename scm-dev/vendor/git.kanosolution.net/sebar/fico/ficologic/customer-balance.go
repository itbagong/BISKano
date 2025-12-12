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

type CustomerBalanceOpt struct {
	CompanyID        string
	GroupByDimension []string
	AccountIDs       []string
	Dimension        tenantcoremodel.Dimension
}

func NewCustomerBalanceHub(db *datahub.Hub) *BalanceHub[*ficomodel.CustomerBalance, *ficomodel.CustomerTransaction, CustomerBalanceOpt] {
	calc := new(BalanceHub[*ficomodel.CustomerBalance, *ficomodel.CustomerTransaction, CustomerBalanceOpt])
	calc.db = db
	calc.provider = new(CustomerBalanceProvider)
	return calc
}

type CustomerBalanceProvider struct {
}

func (*CustomerBalanceProvider) Update(bals []*ficomodel.CustomerBalance, trxs []*ficomodel.CustomerTransaction,
	balDate *time.Time, deduct bool) ([]*ficomodel.CustomerBalance, error) {

	// convert bals ke map untuk mempermudah index
	mapBals := lo.SliceToMap(bals, func(bal *ficomodel.CustomerBalance) (string, *ficomodel.CustomerBalance) {
		return fmt.Sprintf("%s|%s", bal.CustomerID, bal.Dimension.Hash()), bal
	})

	for _, trx := range trxs {
		// ignore non confirmed for now
		if trx.Status != ficomodel.AmountConfirmed {
			continue
		}

		// get map of trx
		trxMapID := fmt.Sprintf("%s|%s", trx.Customer.ID, trx.Dimension.Hash())
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(ficomodel.CustomerBalance)
			bal.CompanyID = trx.CompanyID
			bal.CustomerID = trx.Customer.ID
			bal.Dimension = trx.Dimension
			mapBals[trxMapID] = bal
		}

		if !deduct {
			bal.Balance += trx.Amount
		} else {
			bal.Balance -= trx.Amount
		}
	}
	res := lo.MapToSlice(mapBals, func(_ string, v *ficomodel.CustomerBalance) *ficomodel.CustomerBalance {
		return v
	})

	return res, nil
}

func (*CustomerBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom, dateTo *time.Time,
	initialWhere *dbflex.Filter,
	includeDateTo bool, opt CustomerBalanceOpt) ([]*ficomodel.CustomerTransaction, error) {

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

	res, err := datahub.FindByFilter(db, new(ficomodel.CustomerTransaction), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (*CustomerBalanceProvider) BalanceFilter(obj *ficomodel.CustomerBalance, opt CustomerBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("CustomerID", opt.AccountIDs...))
	}
	if obj != nil {
		if obj.CustomerID != "" {
			wheres = append(wheres, dbflex.Eq("CustomerID", obj.CustomerID))
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

func (*CustomerBalanceProvider) TransactionFilter(obj *ficomodel.CustomerTransaction, opt CustomerBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("Customer._id", opt.AccountIDs...))
	}
	if len(opt.Dimension) > 0 {
		wheres = append(wheres, opt.Dimension.Where())
	}
	if obj != nil {
		if obj.CompanyID != "" {
			wheres = append(wheres, dbflex.Eq("CompanyID", obj.CompanyID))
		}
		if obj.Customer.ID != "" {
			wheres = append(wheres, dbflex.Eq("Customer._id", obj.Customer.ID))
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

func (*CustomerBalanceProvider) AggregateBalance(origModel *ficomodel.CustomerBalance, balModels []*ficomodel.CustomerBalance, add bool, opt CustomerBalanceOpt) error {
	if len(balModels) > 0 {
		if origModel.CompanyID == "" {
			origModel.CompanyID = balModels[0].CompanyID
		}
		origModel.CustomerID = balModels[0].CustomerID
		if len(opt.GroupByDimension) == 0 {
			origModel.Dimension = tenantcorelogic.TernaryDimension(origModel.Dimension, balModels[0].Dimension)
		} else {
			origModel.Dimension = balModels[0].Dimension.Sub(opt.GroupByDimension...)
		}

		// must be zerorised
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

func (*CustomerBalanceProvider) BalanceGrouping(obj *ficomodel.CustomerBalance, opt CustomerBalanceOpt) string {
	dimGroup := getDimensionHash(obj.Dimension, opt.GroupByDimension...)
	return fmt.Sprintf("%s|%s", obj.CustomerID, dimGroup)
}
