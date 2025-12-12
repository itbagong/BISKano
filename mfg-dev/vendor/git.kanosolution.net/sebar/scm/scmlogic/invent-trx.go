package scmlogic

import (
	"errors"
	"fmt"
	"math"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/golang-module/carbon/v2"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"gopkg.in/mgo.v2/bson"
)

type InventTrxEngine struct{}

type InventTrxGetsFilterRequest struct {
	CompanyID      string   `form_required:"1"`
	WarehouseID    string   `form_required:"1"`
	Statuses       []string // scmmodel.ItemBalanceStatus
	ItemIDs        []string
	SourceTrxTypes []scmmodel.InventTrxType
	TrxDate        *struct {
		From time.Time
		To   time.Time
	}
	SourceType      tenantcoremodel.TrxModule
	SourceJournalID string
}

type InventTrxSettledResult struct {
	SourceType      string
	SourceJournalID string
	SourceLineNo    int
	SettledQty      float64 // sum of TrxQty
	Qty             float64 // sum of Qty
}

type CalcInventTrxSettledQtyOptionalParam struct {
	SourceJournalID string
	SourceLineNo    int
}

type GetsByBalanceRequest struct {
	ID           []string
	DateFrom     time.Time
	DateTo       time.Time
	Dimension    []string
	DimInventory struct {
		WarehouseID []interface{}
		SectionID   []interface{}
		SKU         []interface{}
	}
	TrxType string
	Skip    int
	Take    int
}

type GetsByBalanceResponse struct {
	Closing     scmmodel.TrxBalance
	Opening     scmmodel.TrxBalance
	Transaction []scmmodel.InventTrxPerDimension
}

type DisplayPrevPo struct {
	// No PO, Tanggal, Nama Vendor, Item, DPP Net
	PoNo           string    `label:"PO No"`            // InventTransactions.Item.Name
	TransationDate time.Time `label:"Transaction date"` // PurchaseOrders.PODate
	VendorName     string    `label:"Vendor"`           // PurchaseOrders.VendorName
	ItemVarian     string    `grid:"hide"`              // InventTransactions.Item.ItemGroupID + _id + Name
	UoM            string    `label:"UoM"`
	CostUnit       float64   `label:"Unit cost"` // InventTransactions.Item.CostUnit
	DppNet         float64   `grid:"hide"`
	Qty            float64   `label:"Quantity"`
	TotalAmount    float64   `label:"Total amount"`
}

func (o *InventTrxEngine) GetsFilter(ctx *kaos.Context, payload *InventTrxGetsFilterRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	// if e := suim.Validate(payload); e != nil {
	// 	return nil, e
	// }

	if len(payload.Statuses) == 0 {
		payload.Statuses = []string{string(scmmodel.ItemReserved), string(scmmodel.ItemPlanned)}
	}

	pipe := []bson.M{}
	pipeFilters := []bson.M{
		{"CompanyID": payload.CompanyID},
		{"InventDim.WarehouseID": payload.WarehouseID},
		{"Status": bson.M{"$in": payload.Statuses}},
	}

	if len(payload.ItemIDs) > 0 {
		pipeFilters = append(pipeFilters, bson.M{"Item._id": bson.M{"$in": payload.ItemIDs}})
	}
	if len(payload.SourceTrxTypes) > 0 {
		pipeFilters = append(pipeFilters, bson.M{"SourceTrxType": bson.M{"$in": payload.SourceTrxTypes}})
	}
	if payload.TrxDate != nil {
		pipeFilters = append(pipeFilters, bson.M{"$and": []bson.M{
			{"TrxDate": bson.M{"$gte": payload.TrxDate.From}},
			{"TrxDate": bson.M{"$lte": payload.TrxDate.To}},
		}})
	}
	if payload.SourceType != "" {
		pipeFilters = append(pipeFilters, bson.M{"SourceType": payload.SourceType})
	}
	if payload.SourceJournalID != "" {
		pipeFilters = append(pipeFilters, bson.M{"SourceJournalID": payload.SourceJournalID})
	}
	pipe = append(pipe, bson.M{"$match": bson.M{"$and": pipeFilters}})

	pipeGroup := []bson.M{}
	e := Deserialize(fmt.Sprintf(`
    [
        {"$group":{
			"_id":{
				"SourceType":"$SourceType",
				"SourceJournalID":"$SourceJournalID",
				"SourceLineNo":"$SourceLineNo"
			},
			"TrxDate":{"$first":"$TrxDate"},
			"SourceType":{"$first":"$SourceType"},
			"SourceTrxType":{"$first":"$SourceTrxType"},
			"SourceJournalID":{"$first":"$SourceJournalID"},
			"SourceLineNo":{"$first":"$SourceLineNo"},
			"Item":{"$first":"$Item"},
			"SKU":{"$first":"$SKU"},
			"Text":{"$first":"$Text"},
			"TrxQty":{"$sum":"$TrxQty"},
			"Qty":{"$sum":"$Qty"},
			"TrxUnitID":{"$first":"$TrxUnitID"},
			"InventDim":{"$first":"$InventDim"},
			"Dimension":{"$first":"$Dimension"},
			"References":{"$first":"$References"}
		}},
		{"$set":{"_id":""}}
    ]
    `), &pipeGroup)
	if e != nil {
		return nil, e
	}

	pipe = append(pipe, pipeGroup...)
	results := []scmmodel.InventTrxReceipt{}
	cmd := dbflex.From(new(scmmodel.InventTrx).TableName()).Command("pipe", pipe)
	if _, e := h.Populate(cmd, &results); e != nil {
		return nil, fmt.Errorf("error when fetching data : %s", e)
	}

	resSettle, e := CalcInventTrxSettledQty(h, payload.CompanyID)
	if e != nil {
		return nil, e
	}

	settleM := lo.Associate(resSettle, func(d InventTrxSettledResult) (string, InventTrxSettledResult) {
		return fmt.Sprintf("%s||%s||%v", d.SourceType, d.SourceJournalID, d.SourceLineNo), d
	})

	sourceJournalIDs := []string{}
	results = lo.Map(results, func(d scmmodel.InventTrxReceipt, index int) scmmodel.InventTrxReceipt {
		sourceJournalIDs = append(sourceJournalIDs, d.SourceJournalID)
		uniq := fmt.Sprintf("%s||%s||%v", d.SourceType, d.SourceJournalID, d.SourceLineNo)
		d.SettledQty = settleM[uniq].SettledQty
		d.OriginalQty = d.SettledQty + d.TrxQty
		return d
	})

	// add source journal line data
	invjORM := sebar.NewMapRecordWithORM(h, new(scmmodel.InventJournal))
	prjORM := sebar.NewMapRecordWithORM(h, new(scmmodel.PurchaseRequestJournal))
	pojORM := sebar.NewMapRecordWithORM(h, new(scmmodel.PurchaseOrderJournal))
	specs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))

	results = lo.Map(results, func(d scmmodel.InventTrxReceipt, index int) scmmodel.InventTrxReceipt {
		switch d.SourceTrxType {
		case string(scmmodel.JournalMovementIn), string(scmmodel.JournalMovementOut), string(scmmodel.JournalTransfer):
			j, _ := invjORM.Get(d.SourceJournalID)
			line, _, exist := lo.FindIndexOf(j.Lines, func(item scmmodel.InventJournalLine) bool {
				return item.LineNo == d.SourceLineNo
			})
			if exist {
				d.InventJournalLine = line
			}
		case string(scmmodel.PurchRequest):
			j, _ := prjORM.Get(d.SourceJournalID)
			line, _, exist := lo.FindIndexOf(j.Lines, func(item scmmodel.PurchaseJournalLine) bool {
				return item.LineNo == d.SourceLineNo
			})
			if exist {
				d.InventJournalLine = line.InventJournalLine
			}
		case string(scmmodel.PurchOrder):
			j, _ := pojORM.Get(d.SourceJournalID)
			line, _, exist := lo.FindIndexOf(j.Lines, func(item scmmodel.PurchaseJournalLine) bool {
				return item.LineNo == d.SourceLineNo
			})
			if exist {
				d.InventJournalLine = line.InventJournalLine
				d.VendorID = j.VendorID
				d.VendorName = j.VendorName
			}
		default:
			d.InventJournalLine = scmmodel.InventJournalLine{}
		}

		sk, _ := specs.Get(d.SKU)
		d.SKUName = sk.SKU
		d.InventJournalLine.UnitCost = CalcUnitCostFromSourceID(h, d.SourceJournalID)
		d.ItemName = tenantcorelogic.ItemVariantName(h, d.Item.ID, d.SKU)
		return d
	})

	res := codekit.M{
		"data":  results,
		"count": len(results),
	}

	return res, nil
}

func (o *InventTrxEngine) GetsByBalance(ctx *kaos.Context, payload *GetsByBalanceRequest) (interface{}, error) {
	res := GetsByBalanceResponse{}
	//get user and company from context
	companyID, _, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	param := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{
		dbflex.Eq("CompanyID", companyID),
	}

	if len(payload.ID) > 0 {
		filters = append(filters, dbflex.In("Item._id", lo.Map(payload.ID, func(item string, _ int) string {
			return item
		})...))
	}
	opt := ItemBalanceOpt{
		CompanyID: companyID,
		ItemIDs:   payload.ID,
		BalanceFilter: struct {
			WarehouseIDs []interface{}
			SectionIDs   []interface{}
			SKUs         []interface{}
		}{
			WarehouseIDs: payload.DimInventory.WarehouseID,
			SectionIDs:   payload.DimInventory.SectionID,
			SKUs:         payload.DimInventory.SKU,
		},
	}

	//get opening balance
	itemBalances, err := NewItemBalanceHub(h).Get(&payload.DateFrom, opt)
	if err != nil {
		return nil, err
	}

	openingBalance := 0.0
	lo.ForEach(itemBalances, func(balance *scmmodel.ItemBalance, _ int) {
		openingBalance += balance.Qty
	})

	//get closing balance
	itemBalances, err = NewItemBalanceHub(h).Get(&payload.DateTo, opt)
	if err != nil {
		return nil, err
	}

	closingBalance := 0.0
	lo.ForEach(itemBalances, func(balance *scmmodel.ItemBalance, _ int) {
		closingBalance += balance.Qty
	})

	if payload.DateFrom.IsZero() == false {
		payload.DateFrom = carbon.CreateFromStdTime(payload.DateFrom, carbon.Local).StartOfDay().StdTime()
		filters = append(filters, dbflex.Gte("TrxDate", payload.DateFrom))
	}

	if payload.DateTo.IsZero() == false {
		payload.DateTo = carbon.CreateFromStdTime(payload.DateTo, carbon.Local).EndOfDay().StdTime()
		filters = append(filters, dbflex.Lte("TrxDate", payload.DateTo))
	}

	if len(payload.DimInventory.WarehouseID) > 0 {
		filters = append(filters, dbflex.In("InventDim.WarehouseID", payload.DimInventory.WarehouseID...))
	}

	if len(payload.DimInventory.SectionID) > 0 {
		filters = append(filters, dbflex.In("InventDim.SectionID", payload.DimInventory.SectionID...))
	}

	if len(payload.DimInventory.SKU) > 0 {
		filters = append(filters, dbflex.In("SKU", payload.DimInventory.SKU...))
	}
	if payload.TrxType != "" {
		filters = append(filters, dbflex.Eq("SourceTrxType", payload.TrxType))
	}

	filters = append(filters, dbflex.Eq("Status", scmmodel.ItemConfirmed))
	param.SetWhere(dbflex.And(filters...)).
		SetSort("TrxDate")
	// 	SetGroupBy(
	// 		"SourceType", "SourceTrxType", "SourceJournalID", "SourceLineNo", "Status",
	// 	).SetAggr(
	// 	dbflex.NewAggrItem("SourceType", dbflex.AggrFirst, "SourceType"),
	// 	dbflex.NewAggrItem("SourceTrxType", dbflex.AggrFirst, "SourceTrxType"),
	// 	dbflex.NewAggrItem("SourceJournalID", dbflex.AggrFirst, "SourceJournalID"),
	// 	dbflex.NewAggrItem("SourceLineNo", dbflex.AggrFirst, "SourceLineNo"),
	// 	dbflex.NewAggrItem("Status", dbflex.AggrFirst, "Status"),
	// 	dbflex.NewAggrItem("TrxDate", dbflex.AggrFirst, "TrxDate"),
	// 	dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
	// 	dbflex.NewAggrItem("TrxQty", dbflex.AggrSum, "TrxQty"),
	// 	dbflex.NewAggrItem("AmountPhysical", dbflex.AggrSum, "AmountPhysical"),
	// 	dbflex.NewAggrItem("AmountFinancial", dbflex.AggrSum, "AmountFinancial"),
	// 	dbflex.NewAggrItem("AmountAdjustment", dbflex.AggrSum, "AmountAdjustment"),
	// )

	inventTrxs := []scmmodel.InventTrx{}
	if err := h.PopulateByParm(new(scmmodel.InventTrx).TableName(), param, &inventTrxs); err != nil {
		return inventTrxs, err
	}

	// mapTrxs := lo.GroupBy(inventTrxs, func(trx scmmodel.InventTrx) string {
	// 	return fmt.Sprintf("%s|%s|%s", trx.SourceType, trx.SourceJournalID, trx.SourceLineNo)
	// })

	finalTrxs := lo.Map(inventTrxs, func(trx scmmodel.InventTrx, index int) scmmodel.InventTrxPerDimension {
		trxDim := scmmodel.InventTrxPerDimension{}
		trxDim.TrxDate = trx.TrxDate
		trxDim.SourceType = string(trx.SourceType)
		trxDim.SourceTrxType = trx.SourceTrxType
		trxDim.SourceJournalID = trx.SourceJournalID
		trxDim.SourceLineNo = trx.SourceLineNo
		trxDim.Status = trx.Status
		trxDim.Qty = trx.Qty
		trxDim.TrxQty = trx.TrxQty
		trxDim.AmountPhysical = trx.AmountPhysical
		trxDim.AmountFinancial = trx.AmountFinancial
		trxDim.AmountAdjustment = trx.AmountAdjustment

		if trx.AmountFinancial != 0 {
			trxDim.Amount = trx.AmountFinancial + trx.AmountAdjustment
		} else {
			trxDim.Amount = trx.AmountPhysical
		}
		// 	if len(trxs) > 0 {
		// 		trx.SourceType = trxs[0].SourceType
		// 		trx.SourceJournalID = trxs[0].SourceJournalID
		// 		trx.SourceLineNo = trxs[0].SourceLineNo
		// 		trx.SourceTrxType = trxs[0].SourceTrxType
		// 		trx.Qty = trxs[0].Qty
		// 		trx.TrxQty = trxs[0].TrxQty
		// 		trx.TrxDate = trxs[0].TrxDate
		// 	}

		// 	lo.ForEach(trxs, func(trxChild scmmodel.InventTrxPerDimension, index int) {
		// 		trx.Status = trxChild.Status
		// 		if trxChild.Status == scmmodel.ItemConfirmed {
		// 			trx.QtyConfirmed += trxChild.Qty
		// 		} else if trxChild.Status == scmmodel.ItemPlanned {
		// 			trx.QtyPlanned += trxChild.Qty
		// 		} else if trxChild.Status == scmmodel.ItemReserved {
		// 			trx.QtyReserved += trxChild.Qty
		// 		}

		// 		trx.AmountFinancial += trxChild.AmountFinancial
		// 		trx.AmountPhysical += trxChild.AmountPhysical
		// 		trx.AmountAdjustment += trxChild.AmountAdjustment

		// 		if trxChild.AmountFinancial != 0 {
		// 			trx.Amount += (trxChild.AmountFinancial + trxChild.AmountAdjustment)
		// 		} else {
		// 			trx.Amount = trxChild.AmountPhysical
		// 		}
		// 	})

		return trxDim
	})

	// sort.Slice(finalTrxs, func(a, b int) bool {
	// 	return finalTrxs[a].TrxDate.After(finalTrxs[b].TrxDate)
	// })
	//  finalTrxs
	res = GetsByBalanceResponse{
		Closing: scmmodel.TrxBalance{
			Balance: closingBalance,
			Date:    &payload.DateFrom,
		},
		Opening: scmmodel.TrxBalance{
			Balance: openingBalance,
			Date:    &payload.DateFrom,
		},
		Transaction: finalTrxs,
	}

	return res, nil
}

func FindInventTrxBySource(db *datahub.Hub, companyID string, sourceType, sourceID string, sourceLineNo int, status ...scmmodel.ItemBalanceStatus) []*scmmodel.InventTrx {
	filters := []*dbflex.Filter{dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceID", sourceID)}
	if sourceLineNo != 0 {
		filters = append(filters, dbflex.Eq("SourceLineNo", sourceLineNo))
	}
	if len(status) > 0 {
		filters = append(filters, dbflex.In("Status", status...))
	}

	res, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.And(filters...))
	return res
}

func FindInventTrxByItemSpec(db *datahub.Hub, companyID string, itemID, specID, whsID, sectionID string, status ...scmmodel.ItemBalanceStatus) []*scmmodel.InventTrx {
	filters := []*dbflex.Filter{dbflex.Eqs("CompanyID", companyID, "ItemID", itemID)}
	if specID != "" {
		filters = append(filters, dbflex.Eq("InventDim.SpecID", specID))
	}
	if whsID != "" {
		filters = append(filters, dbflex.Eq("WarehouseID", whsID))
	}
	if sectionID != "" {
		filters = append(filters, dbflex.Eq("SectionID", sectionID))
	}
	if len(status) > 0 {
		filters = append(filters, dbflex.In("Status", status...))
	}

	res, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.And(filters...))
	return res
}

func CalcInventTrxSettledQty(h *datahub.Hub, companyID string, optparam ...CalcInventTrxSettledQtyOptionalParam) ([]InventTrxSettledResult, error) {
	pipeFilter := []bson.M{
		{"CompanyID": companyID},
		{"Status": scmmodel.ItemConfirmed},
	}

	if len(optparam) > 0 {
		if optparam[0].SourceJournalID != "" {
			pipeFilter = append(pipeFilter, bson.M{"SourceJournalID": optparam[0].SourceJournalID})
		}
		if optparam[0].SourceLineNo != 0 {
			pipeFilter = append(pipeFilter, bson.M{"SourceLineNo": optparam[0].SourceLineNo})
		}
	}

	pipes := []bson.M{}
	pipes = append(pipes, bson.M{"$match": bson.M{"$and": pipeFilter}})

	pipeGroup := []bson.M{}
	e := Deserialize(fmt.Sprintf(`
    [
		{"$group":{
			"_id":{
				"SourceType":"$SourceType",
				"SourceJournalID":"$SourceJournalID",
				"SourceLineNo":"$SourceLineNo"
			},
			"SourceType":{"$first":"$SourceType"},
			"SourceJournalID":{"$first":"$SourceJournalID"},
			"SourceLineNo":{"$first":"$SourceLineNo"},
			"SettledQty":{"$sum":"$TrxQty"},
			"Qty":{"$sum":"$Qty"}
		}}
    ]
    `), &pipeGroup)
	if e != nil {
		return nil, e
	}
	pipes = append(pipes, pipeGroup...)

	resSettle := []InventTrxSettledResult{}
	cmd := dbflex.From(new(scmmodel.InventTrx).TableName()).Command("pipe", pipes)
	if _, e := h.Populate(cmd, &resSettle); e != nil {
		return nil, fmt.Errorf("error when fetching data : %s", e)
	}

	return resSettle, nil
}

func validateInventTrxQty(h *datahub.Hub, sourceType tenantcoremodel.TrxModule, companyID, sourceJournalID string, sourceLine int, remainingQty, qty float64) bool {
	if remainingQty > 0 && qty > 0 || remainingQty < 0 && qty < 0 {
		return math.Abs(remainingQty) >= math.Abs(qty)
	}

	settles, e := CalcInventTrxSettledQty(h, companyID, CalcInventTrxSettledQtyOptionalParam{
		SourceJournalID: sourceJournalID,
		SourceLineNo:    sourceLine,
	})
	if e != nil || len(settles) == 0 {
		return false
	}

	return qty < settles[0].Qty
}

func (o *InventTrxEngine) findLowerBalanceDate(h *datahub.Hub, companyID string, balanceDate time.Time) *time.Time {
	snapshotBal, _ := datahub.GetByParm(h, new(scmmodel.ItemBalance), dbflex.NewQueryParam().
		SetWhere(dbflex.And(dbflex.Eq("CompanyID", companyID), dbflex.Lte("BalanceDate", balanceDate))).
		SetSort("-BalanceDate").
		SetSelect("BalanceDate"))
	return snapshotBal.BalanceDate
}

func CalcUnitCostFromSourceID(h *datahub.Hub, sourceID string) float64 {
	trxs := []scmmodel.InventTrx{}
	e := h.GetsByFilter(new(scmmodel.InventTrx), dbflex.And(
		dbflex.Eq("SourceJournalID", sourceID),
		dbflex.Eq("Status", scmmodel.ItemPlanned),
	), &trxs)
	if e != nil {
		return 0
	}

	unitCost := 0.0
	lo.ForEach(trxs, func(tr scmmodel.InventTrx, _ int) {
		if tr.Qty != 0 {
			unitCost += (tr.AmountPhysical / tr.Qty)
		}
	})

	return unitCost
}

type GetDisplayPrevRequest struct {
	ItemID string
	SKU    string
}

func (obj *InventTrxEngine) GetDisplayPrev(ctx *kaos.Context, req *GetDisplayPrevRequest) ([]DisplayPrevPo, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	arrRes := []DisplayPrevPo{}

	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if req == nil {
		return nil, errors.New("missing payload")
	}

	wheres := []*dbflex.Filter{
		dbflex.Eq("Item._id", req.ItemID),
		dbflex.Eq("SKU", req.SKU),
		dbflex.Eq("SourceTrxType", "Purchase Order"),
	}
	query := dbflex.NewQueryParam()
	query = query.SetWhere(dbflex.And(wheres...))
	// query = query.SetWhere(dbflex.Eq("Item._id", ItemID))
	query = query.SetSort("-Created")
	query = query.SetTake(5)

	IT := []scmmodel.InventTrx{}
	err := h.Gets(new(scmmodel.InventTrx), query, &IT)
	if err != nil {
		return nil, fmt.Errorf("error when get InventTrx: %s", err.Error())
	}

	for _, b := range IT {
		PO := new(scmmodel.PurchaseOrderJournal)
		if e := h.GetByID(PO, b.SourceJournalID); e != nil {
			return arrRes, fmt.Errorf("PurchaseOrderJournal not found: %s", b.SourceJournalID)
		}

		poItems := lo.Filter(PO.Lines, func(d scmmodel.PurchaseJournalLine, index int) bool {
			return d.ItemID == req.ItemID && d.SKU == req.SKU
		})

		lo.ForEach(poItems, func(poItem scmmodel.PurchaseJournalLine, index int) {
			res := DisplayPrevPo{}
			itemVariantName := tenantcorelogic.ItemVariantName(h, req.ItemID, poItem.SKU)
			uomORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))
			uom, _ := uomORM.Get(poItem.UnitID)

			res.PoNo = b.SourceJournalID

			if PO.TrxDate.IsZero() == false {
				res.TransationDate = PO.TrxDate
			}

			res.VendorName = PO.VendorName
			// res.ItemVarian = b.Item.ItemGroupID + "-" + b.Item.ID + "-" + b.Item.Name
			res.ItemVarian = itemVariantName
			res.UoM = lo.Ternary(uom.Name != "", uom.Name, poItem.UnitID)
			res.CostUnit = poItem.UnitCost
			res.DppNet = 0.0
			res.Qty = b.Qty
			res.TotalAmount = res.Qty * res.CostUnit

			arrRes = append(arrRes, res)
		})
	}

	arrRes = lo.UniqBy(arrRes, func(d DisplayPrevPo) string {
		return fmt.Sprintf("%s||%s", d.ItemVarian, d.PoNo)
	})

	return arrRes, nil
}
