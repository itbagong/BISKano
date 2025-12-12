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

type AssetBalanceOpt struct {
	CompanyID        string
	GroupByDimension []string
	AccountIDs       []string
	Dimension        tenantcoremodel.Dimension
}

func NewAssetBalanceHub(db *datahub.Hub) *BalanceHub[*ficomodel.AssetBalance, *ficomodel.AssetTransaction, AssetBalanceOpt] {
	calc := new(BalanceHub[*ficomodel.AssetBalance, *ficomodel.AssetTransaction, AssetBalanceOpt])
	calc.db = db
	calc.provider = new(AssetBalanceProvider)
	return calc
}

type AssetBalanceProvider struct {
}

func (*AssetBalanceProvider) Update(bals []*ficomodel.AssetBalance, trxs []*ficomodel.AssetTransaction,
	balDate *time.Time, deduct bool) ([]*ficomodel.AssetBalance, error) {

	// convert bals ke map untuk mempermudah index
	mapBals := lo.SliceToMap(bals, func(bal *ficomodel.AssetBalance) (string, *ficomodel.AssetBalance) {
		return fmt.Sprintf("%s|%s", bal.AssetID, bal.Dimension.Hash()), bal
	})

	for _, trx := range trxs {
		// ignore non confirmed for now
		if trx.Status != ficomodel.AmountConfirmed {
			continue
		}

		// get map of trx
		trxMapID := fmt.Sprintf("%s|%s", trx.Asset.ID, trx.Dimension.Hash())
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(ficomodel.AssetBalance)
			bal.CompanyID = trx.CompanyID
			bal.AssetID = trx.Asset.ID
			bal.Dimension = trx.Dimension
			mapBals[trxMapID] = bal
		}

		if !deduct {
			bal.Balance += trx.Amount
		} else {
			bal.Balance -= trx.Amount
		}
	}
	res := lo.MapToSlice(mapBals, func(_ string, v *ficomodel.AssetBalance) *ficomodel.AssetBalance {
		return v
	})

	return res, nil
}

func (*AssetBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom, dateTo *time.Time,
	initialWhere *dbflex.Filter,
	includeDateTo bool, opt AssetBalanceOpt) ([]*ficomodel.AssetTransaction, error) {

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

	res, err := datahub.FindByFilter(db, new(ficomodel.AssetTransaction), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (*AssetBalanceProvider) BalanceFilter(obj *ficomodel.AssetBalance, opt AssetBalanceOpt) *dbflex.Filter {
	if opt.CompanyID == "" {
		return nil
	}

	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("AssetID", opt.AccountIDs...))
	}
	return dbflex.And(wheres...)
}

func (*AssetBalanceProvider) TransactionFilter(obj *ficomodel.AssetTransaction, opt AssetBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("Customer._id", opt.AccountIDs...))
	}
	if len(opt.Dimension) > 0 {
		wheres = append(wheres, opt.Dimension.Where())
	}
	if obj.CompanyID != "" {
		wheres = append(wheres, dbflex.Eq("CompanyID", obj.CompanyID))
	}
	if obj.Asset.ID != "" {
		wheres = append(wheres, dbflex.Eq("Customer._id", obj.Asset.ID))
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

func (*AssetBalanceProvider) AggregateBalance(origModel *ficomodel.AssetBalance, aggrModels []*ficomodel.AssetBalance, add bool, opt AssetBalanceOpt) error {
	if len(aggrModels) > 0 {
		if origModel.CompanyID == "" {
			origModel.CompanyID = aggrModels[0].CompanyID
		}
		origModel.AssetID = aggrModels[0].AssetID
		origModel.Dimension = tenantcorelogic.TernaryDimension(origModel.Dimension, aggrModels[0].Dimension)
	}

	for _, item := range aggrModels {
		mul := float64(1)
		if !add {
			mul = -1
		}
		origModel.Balance += mul * item.Balance
	}
	return nil
}

func (*AssetBalanceProvider) BalanceGrouping(obj *ficomodel.AssetBalance, opt AssetBalanceOpt) string {
	dimGroup := getDimensionHash(obj.Dimension, opt.GroupByDimension...)
	return fmt.Sprintf("%s|%s", obj.AssetID, dimGroup)
}
