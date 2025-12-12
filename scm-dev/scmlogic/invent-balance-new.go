package scmlogic

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type ItemBalanceOpt struct {
	CompanyID   string
	ItemIDs     []string
	ConsiderSKU bool
	InventDim   scmmodel.InventDimension
	//InventTrxIDs    []string
	DisableGrouping bool
	GroupBy         []string
	BalanceFilter   struct {
		WarehouseIDs []interface{}
		SectionIDs   []interface{}
		SKUs         []interface{}
	}
}

type ItemBalanceProvider struct {
	db      *datahub.Hub
	itemIDs []interface{}
}

func NewItemBalanceHub2(db *datahub.Hub, ev kaos.EventHub, opt ItemBalanceOpt) *ficologic.BalanceHub[*scmmodel.ItemBalance, *scmmodel.InventTrx, ItemBalanceOpt] {
	// pvd := ficologic.BalanceHubProvider[*scmmodel.ItemBalance, *scmmodel.InventTrx, ItemBalaceOpt](LedgerBalanceProvider)
	itemBalance := new(ItemBalanceProvider)
	itemBalance.db = db
	if len(opt.ItemIDs) > 0 {
		itemBalance.itemIDs = lo.Map(opt.ItemIDs, func(itemID string, index int) interface{} {
			return itemID
		})
	}

	itemProvider := ficologic.BalanceHubProvider[*scmmodel.ItemBalance, *scmmodel.InventTrx, ItemBalanceOpt](itemBalance)
	calc := ficologic.NewBalanceHub(db, itemProvider, ev, opt)
	return calc
}

func NewItemBalanceHub(db *datahub.Hub) *ficologic.BalanceHub[*scmmodel.ItemBalance, *scmmodel.InventTrx, ItemBalanceOpt] {
	itemBalance := new(ItemBalanceProvider)
	itemBalance.db = db
	itemProvider := ficologic.BalanceHubProvider[*scmmodel.ItemBalance, *scmmodel.InventTrx, ItemBalanceOpt](itemBalance)
	calc := ficologic.NewBalanceHub(db, itemProvider, nil, ItemBalanceOpt{})
	return calc
}

func (o *ItemBalanceProvider) BalanceFilter(obj *scmmodel.ItemBalance, opt ItemBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
	}

	if len(opt.ItemIDs) > 0 {
		wheres = append(wheres, dbflex.In("ItemID", opt.ItemIDs...))

		dimIDs := []string{}
		for _, itemID := range opt.ItemIDs {
			itemDimIds, _ := FindInventDimIDs(o.db, opt.CompanyID, itemID, "InventDim", opt.InventDim)
			dimIDs = append(dimIDs, itemDimIds...)
		}

		if len(opt.BalanceFilter.WarehouseIDs) > 0 || len(opt.BalanceFilter.SectionIDs) > 0 {
			if len(opt.BalanceFilter.WarehouseIDs) > 0 {
				wheres = append(wheres, dbflex.In("InventDim.WarehouseID", opt.BalanceFilter.WarehouseIDs...))
			}

			if len(opt.BalanceFilter.SectionIDs) > 0 {
				wheres = append(wheres, dbflex.In("InventDim.SectionID", opt.BalanceFilter.SectionIDs...))
			}
		} else {
			dimIDs = lo.Uniq(dimIDs)
			if len(dimIDs) > 0 {
				wheres = append(wheres, dbflex.In("InventDim.InventDimID", dimIDs...))
			}
		}
	}

	if len(opt.BalanceFilter.SKUs) > 0 {
		wheres = append(wheres, dbflex.In("SKU", opt.BalanceFilter.SKUs...))
	}

	return dbflex.And(wheres...)
}

func (o *ItemBalanceProvider) TransactionFilter(obj *scmmodel.InventTrx, opt ItemBalanceOpt) *dbflex.Filter {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
	}

	if len(opt.ItemIDs) > 0 {
		wheres = append(wheres, dbflex.In("Item._id", opt.ItemIDs...))

		dimIDs := []string{}
		for _, itemID := range opt.ItemIDs {
			itemDimIds, _ := FindInventDimIDs(o.db, opt.CompanyID, itemID, "InventDim", opt.InventDim)
			dimIDs = append(dimIDs, itemDimIds...)
		}
		dimIDs = lo.Uniq(dimIDs)
		if len(dimIDs) > 0 {
			wheres = append(wheres, dbflex.In("InventDim.InventDimID", dimIDs...))
		}
	}

	return dbflex.And(wheres...)
}

func (o *ItemBalanceProvider) BalanceGrouping(obj *scmmodel.ItemBalance, opt ItemBalanceOpt) string {
	groupMap := lo.SliceToMap(opt.GroupBy, func(groupBy string) (string, bool) {
		return groupBy, true
	})

	finalGroup := []string{opt.CompanyID, obj.ItemID}
	if opt.ConsiderSKU {
		finalGroup = append(finalGroup, obj.SKU)
	}

	if len(opt.GroupBy) > 0 {
		dim := new(scmmodel.InventDimension)
		if _, ok := groupMap["WarehouseID"]; ok {
			dim.WarehouseID = obj.InventDim.WarehouseID
			// finalGroup = append(finalGroup, obj.InventDim.WarehouseID)
		}

		if _, ok := groupMap["SectionID"]; ok {
			dim.SectionID = obj.InventDim.SectionID
			// finalGroup = append(finalGroup, obj.InventDim.SectionID)
		}

		if _, ok := groupMap["AisleID"]; ok {
			dim.AisleID = obj.InventDim.AisleID
			// finalGroup = append(finalGroup, obj.InventDim.AisleID)
		}

		if _, ok := groupMap["BoxID"]; ok {
			dim.BoxID = obj.InventDim.BoxID
			// finalGroup = append(finalGroup, obj.InventDim.BoxID)
		}

		if _, ok := groupMap["BatchID"]; ok {
			dim.BatchID = obj.InventDim.BatchID
			// finalGroup = append(finalGroup, obj.InventDim.BatchID)
		}

		if _, ok := groupMap["VariantID"]; ok {
			dim.VariantID = obj.InventDim.VariantID
			// finalGroup = append(finalGroup, obj.InventDim.VariantID)
		}

		if _, ok := groupMap["SerialNumber"]; ok {
			dim.SerialNumber = obj.InventDim.SerialNumber
			// finalGroup = append(finalGroup, obj.InventDim.SerialNumber)
		}

		if _, ok := groupMap["Size"]; ok {
			dim.Size = obj.InventDim.Size
			// finalGroup = append(finalGroup, obj.InventDim.Size)
		}

		if _, ok := groupMap["Grade"]; ok {
			dim.Grade = obj.InventDim.Grade
			// finalGroup = append(finalGroup, obj.InventDim.Grade)
		}

		dim.Calc()
		finalGroup = append(finalGroup, dim.InventDimID)
	} else if opt.DisableGrouping {
		finalGroup = append(finalGroup, obj.InventDim.InventDimID)
	}

	return strings.Join(finalGroup, "|")
}

func (o *ItemBalanceProvider) AggregateBalance(origmodel *scmmodel.ItemBalance, balModels []*scmmodel.ItemBalance, add bool, opt ItemBalanceOpt) error {
	var origDim scmmodel.InventDimension
	if len(balModels) > 0 {
		origmodel.CompanyID = balModels[0].CompanyID
		origmodel.ItemID = balModels[0].ItemID
		if opt.ConsiderSKU {
			origmodel.SKU = balModels[0].SKU
		}
		origDim = balModels[0].InventDim
	}

	totalQtyPlanned := 0.0
	totalQtyReserved := 0.0
	totalQty := 0.0
	totalAmountPhysical := 0.0
	totalAmountFinancial := 0.0
	totalAmountAdjustment := 0.0

	lo.ForEach(balModels, func(balance *scmmodel.ItemBalance, index int) {
		totalQtyPlanned += balance.QtyPlanned
		totalQtyReserved += balance.QtyReserved
		totalQty += balance.Qty
		totalAmountAdjustment += balance.AmountAdjustment
		totalAmountFinancial += balance.AmountFinancial
		totalAmountPhysical += balance.AmountPhysical
	})

	origmodel.QtyPlanned = totalQtyPlanned
	origmodel.QtyReserved = totalQtyReserved
	origmodel.Qty = totalQty
	origmodel.AmountAdjustment = totalAmountAdjustment
	origmodel.AmountFinancial = totalAmountFinancial
	origmodel.AmountPhysical = totalAmountPhysical

	origmodel.CompanyID = balModels[0].CompanyID
	origmodel.ItemID = balModels[0].ItemID
	origmodel.InventDim = scmmodel.InventDimension{}
	if len(opt.GroupBy) > 0 && len(balModels) > 0 {
		for _, gb := range opt.GroupBy {
			switch gb {
			case "VariantID":
				origmodel.InventDim.VariantID = origDim.VariantID

			case "Size":
				origmodel.InventDim.Size = origDim.Size

			case "Grade":
				origmodel.InventDim.Grade = origDim.Grade

			case "WarehouseID":
				origmodel.InventDim.WarehouseID = origDim.WarehouseID

			case "SectionID":
				origmodel.InventDim.SectionID = origDim.SectionID

			case "AisleID":
				origmodel.InventDim.AisleID = origDim.AisleID

			case "BoxID":
				origmodel.InventDim.BoxID = origDim.BoxID

			case "BatchID":
				origmodel.InventDim.BatchID = origDim.BatchID

			case "SerialNumber":
				origmodel.InventDim.SerialNumber = origDim.SerialNumber
			}

			origmodel.InventDim = *origmodel.InventDim.Calc()
		}
	} else if opt.DisableGrouping {
		origmodel.InventDim = origDim
	}

	return nil
}

func (o *ItemBalanceProvider) GetTransactions(db *datahub.Hub, dateFrom *time.Time, dateTo *time.Time, initialWhere *dbflex.Filter, includeDateTo bool, opt ItemBalanceOpt) ([]*scmmodel.InventTrx, error) {
	trxWheres := []*dbflex.Filter{initialWhere}
	if dateFrom != nil && !dateFrom.IsZero() {
		trxWheres = append(trxWheres, dbflex.Gt("TrxDate", dateFrom))
	}

	if dateTo != nil && !dateTo.IsZero() {
		if includeDateTo {
			trxWheres = append(trxWheres, dbflex.Lte("TrxDate", dateTo))
		} else {
			trxWheres = append(trxWheres, dbflex.Lt("TrxDate", dateTo))
		}
	}

	res, err := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.And(trxWheres...))
	if err != nil {
		return nil, err
	}

	res = lo.Map(res, func(re *scmmodel.InventTrx, index int) *scmmodel.InventTrx {
		dim := scmmodel.InventDimension{}
		if len(opt.GroupBy) > 0 {
			for _, gb := range opt.GroupBy {
				switch gb {
				case "Variant":
					dim.VariantID = re.InventDim.VariantID

				case "Size":
					dim.Size = re.InventDim.Size

				case "Grade":
					dim.Grade = re.InventDim.Grade

				case "Warehouse":
					dim.WarehouseID = re.InventDim.WarehouseID

				case "Section":
					dim.SectionID = re.InventDim.SectionID

				case "Aisle":
					dim.AisleID = re.InventDim.AisleID

				case "Box":
					dim.BoxID = re.InventDim.BoxID

				case "Batch":
					dim.BatchID = re.InventDim.BatchID

				case "SerialNumber":
					dim.SerialNumber = re.InventDim.SerialNumber
				}
			}

			dim = *dim.Calc()
			re.InventDim = dim
		} else if !opt.DisableGrouping {
			re.InventDim = dim
		}
		return re
	})

	return res, nil
}

func (o *ItemBalanceProvider) Update(bals []*scmmodel.ItemBalance, trxs []*scmmodel.InventTrx, balDate *time.Time, deduct bool) ([]*scmmodel.ItemBalance, error) {
	mapBals := lo.SliceToMap(bals, func(bal *scmmodel.ItemBalance) (string, *scmmodel.ItemBalance) {
		return fmt.Sprintf("%s|%s|%s|%s", bal.CompanyID, bal.ItemID, bal.SKU, bal.InventDim.InventDimID), bal
	})

	lo.ForEach(trxs, func(trx *scmmodel.InventTrx, index int) {
		trxMapID := fmt.Sprintf("%s|%s|%s|%s", trx.CompanyID, trx.Item.ID, trx.SKU, trx.InventDim.InventDimID)
		bal, inMap := mapBals[trxMapID]
		if !inMap {
			bal = new(scmmodel.ItemBalance)
			bal.CompanyID = trx.CompanyID
			bal.InventDim = trx.InventDim
			bal.ItemID = trx.Item.ID
			bal.SKU = trx.SKU
			mapBals[trxMapID] = bal
		}

		if !deduct {
			switch trx.Status {
			case scmmodel.ItemPlanned:
				bal.QtyPlanned += trx.Qty
			case scmmodel.ItemConfirmed:
				bal.Qty += trx.Qty
			case scmmodel.ItemReserved:
				bal.QtyReserved += trx.Qty
			}

			bal.AmountAdjustment += trx.AmountAdjustment
			bal.AmountPhysical += trx.AmountPhysical
			bal.AmountFinancial += trx.AmountFinancial
		} else {
			switch trx.Status {
			case scmmodel.ItemPlanned:
				bal.QtyPlanned -= trx.Qty
			case scmmodel.ItemConfirmed:
				bal.Qty -= trx.Qty
			case scmmodel.ItemReserved:
				bal.QtyReserved -= trx.Qty
			}

			bal.AmountAdjustment -= trx.AmountAdjustment
			bal.AmountPhysical -= trx.AmountPhysical
			bal.AmountFinancial -= trx.AmountFinancial
		}
	})

	res := lo.MapToSlice(mapBals, func(_ string, v *scmmodel.ItemBalance) *scmmodel.ItemBalance {
		return v
	})

	return res, nil
}

func FindInventDims(db *datahub.Hub, companyID, itemID, inventDimFieldName string, dim scmmodel.InventDimension) ([]scmmodel.InventDimension, error) {
	eqs := []*dbflex.Filter{dbflex.Eqs("CompanyID", companyID, "Item._id", itemID)}
	if dim.WarehouseID != "" {
		eqs = append(eqs, dbflex.Eq(inventDimFieldName+".WarehouseID", dim.WarehouseID))
	}
	dimIdFieldName := inventDimFieldName + ".InventDimID"
	param := dbflex.NewQueryParam().
		SetWhere(dbflex.And(eqs...)).
		SetAggr(dbflex.NewAggrItem("InventDim", dbflex.AggrFirst, "InventDim")).
		SetGroupBy(dimIdFieldName)

	type dimAggr struct {
		ID struct {
			InventDimID string `json:"InventDimID" bson:"InventDimID"`
		} `json:"_id" bson:"_id"`
		InventDim scmmodel.InventDimension
	}

	dest := []dimAggr{}
	err := db.PopulateByParm(new(scmmodel.InventTrx).TableName(), param, &dest)
	if err != nil {
		return nil, err
	}

	trxs := lo.Filter(dest, func(t dimAggr, index int) bool {
		if dim.VariantID != "" && dim.VariantID != t.InventDim.VariantID {
			return false
		}

		if dim.Size != "" && dim.VariantID != t.InventDim.Size {
			return false
		}

		if dim.Grade != "" && dim.VariantID != t.InventDim.Grade {
			return false
		}

		if dim.SectionID != "" && dim.VariantID != t.InventDim.SectionID {
			return false
		}

		if dim.AisleID != "" && dim.VariantID != t.InventDim.AisleID {
			return false
		}

		if dim.BoxID != "" && dim.VariantID != t.InventDim.BoxID {
			return false
		}

		if dim.BatchID != "" && dim.VariantID != t.InventDim.BatchID {
			return false
		}

		if dim.SerialNumber != "" && dim.VariantID != t.InventDim.SerialNumber {
			return false
		}

		return true
	})

	dims := lo.Map(trxs, func(t dimAggr, index int) scmmodel.InventDimension {
		return t.InventDim
	})

	return dims, nil
}

func FindInventDimIDs(db *datahub.Hub, companyID, itemID, inventDimFieldName string, dim scmmodel.InventDimension) ([]string, error) {
	dims, err := FindInventDims(db, companyID, itemID, inventDimFieldName, dim)
	if err != nil {
		return nil, err
	}

	return lo.Map(dims, func(t scmmodel.InventDimension, index int) string {
		return t.InventDimID
	}), nil
}
