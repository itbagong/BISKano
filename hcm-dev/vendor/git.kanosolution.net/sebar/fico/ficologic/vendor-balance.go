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

type VendorBalanceOpt struct {
	CompanyID        string
	GroupByDimension []string
	AccountIDs       []string
	Dimension        tenantcoremodel.Dimension
}

func NewVendorBalanceHub(db *datahub.Hub) *BalanceHub[*ficomodel.VendorBalance, *ficomodel.VendorTransaction, VendorBalanceOpt] {
	calc := new(BalanceHub[*ficomodel.VendorBalance, *ficomodel.VendorTransaction, VendorBalanceOpt])
	calc.db = db
	calc.provider = new(VendorBalanceProvider)
	return calc
}

type VendorBalanceProvider struct {
}

func (*VendorBalanceProvider) Update(bals []*ficomodel.VendorBalance, trxs []*ficomodel.VendorTransaction,
	balDate *time.Time, deduct bool) ([]*ficomodel.VendorBalance, error) {

	// convert bals ke map untuk mempermudah index
	mapBals := lo.SliceToMap(bals, func(bal *ficomodel.VendorBalance) (string, *ficomodel.VendorBalance) {
		return fmt.Sprintf("%s|%s", bal.VendorID, bal.Dimension.Hash()), bal
	})

	for _, trx := range trxs {
		// ignore non confirmed for now
		if trx.Status != ficomodel.AmountConfirmed {
			continue
		}

		// get map of trx
		trxMapID := fmt.Sprintf("%s|%s", trx.Vendor.ID, trx.Dimension.Hash())
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(ficomodel.VendorBalance)
			bal.CompanyID = trx.CompanyID
			bal.VendorID = trx.Vendor.ID
			bal.Dimension = trx.Dimension
			mapBals[trxMapID] = bal
		}

		if !deduct {
			bal.Balance += trx.Amount
		} else {
			bal.Balance -= trx.Amount
		}
	}
	res := lo.MapToSlice(mapBals, func(_ string, v *ficomodel.VendorBalance) *ficomodel.VendorBalance {
		return v
	})

	return res, nil
}

func (*VendorBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom, dateTo *time.Time,
	initialWhere *dbflex.Filter,
	includeDateTo bool, opt VendorBalanceOpt) ([]*ficomodel.VendorTransaction, error) {

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

	res, err := datahub.FindByFilter(db, new(ficomodel.VendorTransaction), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (*VendorBalanceProvider) BalanceFilter(obj *ficomodel.VendorBalance, opt VendorBalanceOpt) *dbflex.Filter {
	if opt.CompanyID == "" {
		return nil
	}

	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("VendorID", opt.AccountIDs...))
	}
	return dbflex.And(wheres...)
}

func (*VendorBalanceProvider) TransactionFilter(obj *ficomodel.VendorTransaction, opt VendorBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		opt.Dimension.Where(),
	}
	if len(opt.AccountIDs) > 0 {
		wheres = append(wheres, dbflex.In("Vendor._id", opt.AccountIDs...))
	}
	if len(opt.Dimension) > 0 {
		wheres = append(wheres, opt.Dimension.Where())
	}
	if obj.CompanyID != "" {
		wheres = append(wheres, dbflex.Eq("CompanyID", obj.CompanyID))
	}
	if obj.Vendor.ID != "" {
		wheres = append(wheres, dbflex.Eq("Vendor._id", obj.Vendor.ID))
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

func (*VendorBalanceProvider) AggregateBalance(origModel *ficomodel.VendorBalance, aggrModels []*ficomodel.VendorBalance, add bool, opt VendorBalanceOpt) error {
	if len(aggrModels) > 0 {
		if origModel.CompanyID == "" {
			origModel.CompanyID = aggrModels[0].CompanyID
		}
		origModel.VendorID = aggrModels[0].VendorID
		origModel.BalanceDate = aggrModels[0].BalanceDate
		if len(opt.GroupByDimension) == 0 {
			origModel.Dimension = tenantcorelogic.TernaryDimension(origModel.Dimension, aggrModels[0].Dimension)
		} else {
			origModel.Dimension = aggrModels[0].Dimension.Sub(opt.GroupByDimension...)
		}
		origModel.Balance = 0
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

func (*VendorBalanceProvider) BalanceGrouping(obj *ficomodel.VendorBalance, opt VendorBalanceOpt) string {
	dimGroup := getDimensionHash(obj.Dimension, opt.GroupByDimension...)
	return fmt.Sprintf("%s|%s", obj.VendorID, dimGroup)
}
