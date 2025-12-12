package scmlogic

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func GetCostPerUnit(db *datahub.Hub, item tenantcoremodel.Item, dim scmmodel.InventDimension, trxDate *time.Time) float64 {
	dim.Calc()
	if trxDate != nil {
		*trxDate = codekit.String2Date(codekit.Date2String(*trxDate, "yyyy-MM-dd"), "yyyy-MM-dd")
	}
	iuc, err := datahub.GetByFilter(db, new(scmmodel.ItemUnitCost), dbflex.Eqs("ItemID", item.ID,
		"InventDim.SpecID", dim.SpecID,
		"InventDim.WarehouseID", dim.WarehouseID,
		"TrxDate", trxDate))
	if err != nil {
		iuc = CalcUnitCost(db, item, dim, trxDate)
	}
	if iuc.UnitCost == 0 {
		return item.CostUnit
	}
	return iuc.UnitCost
}

func CalcUnitCost(db *datahub.Hub, item tenantcoremodel.Item, dim scmmodel.InventDimension, trxDate *time.Time) *scmmodel.ItemUnitCost {
	dim.Calc()

	if trxDate != nil {
		*trxDate = codekit.String2Date(codekit.Date2String(*trxDate, "yyyy-MM-dd"), "yyyy-MM-dd")
	}

	//get last unit cost before date
	where := dbflex.Eqs("ItemID", item.ID,
		"InventDim.SpecID", dim.SpecID,
		"InventDim.WarehouseID", dim.WarehouseID,
	)
	if trxDate != nil {
		where = dbflex.And(where, dbflex.Lt("TrxDate", trxDate))
	}
	iuc, err := datahub.GetByParm(db, new(scmmodel.ItemUnitCost), dbflex.NewQueryParam().SetSort("-TrxDate").
		SetWhere(where))
	if err != nil {
		iuc = new(scmmodel.ItemUnitCost)
		iuc.ItemID = item.ID
		iuc.InventDim = dim
	} else {
		iuc.ID = ""
	}

	// get transaction
	wheres := []*dbflex.Filter{dbflex.Eqs("Item._id", item.ID,
		"InventDim.SpecID", dim.SpecID,
		"InventDim.WarehouseID", dim.WarehouseID,
		"Status", scmmodel.ItemConfirmed), dbflex.Gt("Qty", 0)}
	if iuc.TrxDate != nil {
		wheres = append(wheres, dbflex.Gt("TrxDate", iuc.TrxDate))
	}
	if trxDate != nil {
		wheres = append(wheres, dbflex.Lt("TrxDate", trxDate.AddDate(0, 0, 1)))
	}
	trxs, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.And(wheres...))
	iuc.UnitCost += lo.SumBy(trxs, func(trx *scmmodel.InventTrx) float64 {
		return (lo.Ternary(trx.AmountFinancial == 0, trx.AmountPhysical, trx.AmountFinancial) + trx.AmountAdjustment) / trx.Qty
	})

	// adjust
	if iuc.UnitCost == 0 {
		return iuc
	}

	resIuc, err := datahub.GetByFilter(db, new(scmmodel.ItemUnitCost), dbflex.Eqs("ItemID", item.ID,
		"InventDim.SpecID", dim.SpecID,
		"InventDim.WarehouseID", dim.WarehouseID,
		"TrxDate", trxDate))
	if err != nil {
		resIuc = iuc
		iuc.TrxDate = trxDate
	}
	resIuc.UnitCost = iuc.UnitCost
	db.Save(resIuc)

	return resIuc
}
