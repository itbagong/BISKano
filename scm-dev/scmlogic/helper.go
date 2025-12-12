package scmlogic

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/leekchan/accounting"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

func GetCompanyIDFromContext(ctx *kaos.Context) (string, error) {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO00", nil
		//return coID, errors.New("session: Company ID not found, please relogin")
	}
	return coID, nil
}

func GetCompanyAndUserIDFromContext(ctx *kaos.Context) (string, string, error) {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	userID := sebar.GetUserIDFromCtx(ctx)

	if coID == "" || userID == "" {
		return "", "", fmt.Errorf("session expired")
	}

	return coID, userID, nil
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}

// Deserialize decodes query string into a map
func Deserialize(str string, res interface{}) error {
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		return fmt.Errorf("Deserialize Failed | %s:\n%s", err, str)
	}
	return err
}

type GeneralRequest struct {
	Select []string
	Sort   []string
	Skip   int
	Take   int
	Where  *dbflex.Filter
}

func (o *GeneralRequest) GetQueryParam() *dbflex.QueryParam {
	parm := dbflex.NewQueryParam()

	// TODO: do something with o.Select

	if len(o.Sort) > 0 {
		parm = parm.SetSort(o.Sort...)
	}

	parm = parm.SetSkip(o.Skip)
	parm = parm.SetTake(o.Take)

	if o.Where != nil {
		parm = parm.SetWhere(o.Where)
	}

	return parm
}

func (o *GeneralRequest) GetQueryParamWithURLQuery(ctx *kaos.Context) *dbflex.QueryParam {
	parm := dbflex.NewQueryParam()

	// TODO: do something with o.Select

	if len(o.Sort) > 0 {
		parm = parm.SetSort(o.Sort...)
	}

	parm = parm.SetSkip(o.Skip)
	parm = parm.SetTake(o.Take)

	fs := []*dbflex.Filter{}

	if qfs := GetURLQueryParamFilter(ctx); qfs != nil {
		fs = append(fs, qfs)
	}

	if o.Where != nil {
		fs = append(fs, o.Where)
	}

	if len(fs) > 0 {
		parm = parm.SetWhere(dbflex.And(fs...))
	}

	return parm
}

func GetURLQueryParams(ctx *kaos.Context) map[string]string {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return map[string]string{}
	}

	res := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			res[key] = values[0]
		}
	}

	return res
}

func GetURLQueryParamFilter(ctx *kaos.Context, op ...dbflex.FilterOp) *dbflex.Filter {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return nil
	}

	fs := []*dbflex.Filter{}

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			fs = append(fs, dbflex.Eq(key, values[0]))
		} else if len(values) > 1 {
			fsor := []*dbflex.Filter{}
			for _, v := range values {
				fsor = append(fsor, dbflex.Eq(key, v))
			}
			fs = append(fs, dbflex.Or(fsor...))
		}
	}

	if len(fs) == 0 {
		return nil
	}

	if len(op) > 0 && op[0] == dbflex.OpOr {
		return dbflex.Or(fs...)
	}

	return dbflex.And(fs...)
}

// InventTrxSingleJournalSave hanya untuk trx / line yang berasal dari 1 journal saja, ketika trx / line bisa berasal dari beda journal Jangan Pakai Ini !!! pakai yang InventTrxMultiJournalSave
func InventTrxSingleJournalSave(h *datahub.Hub, trxs map[string][]orm.DataModel, companyID string, sourceType interface{}, sourceJournalID string, status scmmodel.ItemBalanceStatus) ([]*scmmodel.InventTrx, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
	if len(inventTrxs) == 0 {
		return nil, nil // prevent deleting without saving
	}

	h.DeleteByFilter(new(scmmodel.InventTrx),
		dbflex.Eqs("CompanyID", companyID, "SourceType", codekit.ToString(sourceType), "SourceJournalID", sourceJournalID),
	)

	for _, trx := range inventTrxs {
		trx.Status = status
		if e := h.Save(trx); e != nil {
			return nil, e
		}
	}

	return inventTrxs, nil
}

// InventTrxMultiJournalSave untuk trx / line yang bisa berasal dari journal yang berbeda
func InventTrxMultiJournalSave(h *datahub.Hub, trxs map[string][]orm.DataModel) ([]*scmmodel.InventTrx, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))

	for _, trx := range inventTrxs {
		h.DeleteByFilter(new(scmmodel.InventTrx), dbflex.Eqs(
			"CompanyID", trx.CompanyID,
			"SourceType", trx.SourceType,
			"SourceJournalID", trx.SourceJournalID,
			"Status", trx.Status,
			"Item._id", trx.Item.ID,
			"SKU", trx.SKU,
			// TODO: butuh SourceLineNo ga ya? khawatir kalau ada 2 Line dengan Item._id dan SKU yang sama
		))

		if e := h.Save(trx); e != nil {
			return nil, e
		}
	}

	return inventTrxs, nil
}

func InventTrxSplitSave(h *datahub.Hub, trxs map[string][]orm.DataModel, sourceStatus scmmodel.ItemBalanceStatus) (map[string][]orm.DataModel, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))

	splittedTrxs := []*scmmodel.InventTrx{}
	spliter := NewInventSplit(h)
	for _, trx := range inventTrxs {
		updatedSourceTrxs, newTrxs, err := spliter.SetOpts(&InventSplitOpts{
			SplitType:       SplitBySource,
			CompanyID:       trx.CompanyID,
			SourceType:      string(trx.SourceType),
			SourceJournalID: trx.SourceJournalID,
			SourceLineNo:    trx.SourceLineNo,
			SourceStatus:    string(sourceStatus),
		}).Split(trx.Qty, string(scmmodel.ItemConfirmed))
		if err != nil {
			return nil, err
		}

		splittedTrxs = append(splittedTrxs, append(updatedSourceTrxs, newTrxs...)...)
	}

	if len(splittedTrxs) == 0 {
		return nil, fmt.Errorf("fail split invent trx, return 0 trx")
	}

	delete(trxs, new(scmmodel.InventTrx).TableName())
	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(splittedTrxs)

	return trxs, nil
}

func FormatMoney(money float64) string {
	return accounting.FormatNumber(money, 2, ",", ".")
}

func FormatDate(paramDate *time.Time) string {
	if paramDate == nil {
		return ""
	}

	return paramDate.Format("02 January 2006")
}

func receiveIssueLineToTrx(db *datahub.Hub, header *scmmodel.InventReceiveIssueJournal, line scmmodel.InventReceiveIssueLine) (*scmmodel.InventTrx, error) {
	// var err error

	trx := new(scmmodel.InventTrx)
	trx.CompanyID = header.CompanyID
	trx.Text = line.Text
	trx.Item = line.Item
	trx.SKU = line.SKU
	trx.Dimension = line.Dimension
	trx.InventDim = line.InventDim

	trx.Qty = line.InventQty
	trx.TrxQty = line.Qty
	trx.TrxUnitID = line.UnitID

	trx.SourceType = line.SourceType
	trx.SourceJournalID = line.SourceJournalID
	trx.SourceTrxType = line.SourceTrxType
	trx.SourceLineNo = line.SourceLineNo

	trx.AmountPhysical = line.UnitCost * line.Qty
	trx.References = trx.References.Set(scmmodel.GetRefKey(line.SourceTrxType), line.SourceJournalID)

	trx.TrxDate = header.TrxDate
	return trx, nil
}

func ItemMinMaxValidation(h *datahub.Hub, inventTrxs []*scmmodel.InventTrx, sourceTrxType string) error {
	errorMsgs := []string{}
	itemSpecIDs := make([]string, len(inventTrxs))
	itemIDs := make([]string, len(inventTrxs))
	InventDimIDs := make([]string, len(inventTrxs))
	for i, trx := range inventTrxs {
		itemSpecIDs[i] = trx.SKU
		itemIDs[i] = trx.Item.ID
		InventDimIDs[i] = trx.InventDim.InventDimID
	}

	specs := []tenantcoremodel.ItemSpec{}
	err := h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemSpecIDs...),
	), &specs)
	if err != nil {
		return fmt.Errorf("error when get item spec: %s", err.Error())
	}

	mapItemSpec := lo.Associate(specs, func(v tenantcoremodel.ItemSpec) (string, tenantcoremodel.ItemSpec) {
		return v.ID, v
	})

	mapAssignItem, err := AssignItem(h, itemIDs, itemSpecIDs)
	if err != nil {
		return nil
	}

	// get in max
	pipe := []bson.M{
		{
			"$match": bson.M{
				"ItemID":                         bson.M{"$in": itemIDs},
				"SKU":                            bson.M{"$in": itemSpecIDs},
				"InventoryDimension.InventDimID": bson.M{"$in": InventDimIDs},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"ItemID":      "$ItemID",
					"SKU":         "$SKU",
					"InventDimID": "$InventoryDimension.InventDimID",
				},
				"Min": bson.M{"$min": "$MinStock"},
				"Max": bson.M{"$max": "$MaxStock"},
			},
		},
		{
			"$project": bson.M{
				"ItemID":      "$_id.ItemID",
				"SKU":         "$_id.SKU",
				"InventDimID": "$_id.InventDimID",
				"Min":         1,
				"Max":         1,
			},
		},
	}

	type item struct {
		ItemID      string
		SKU         string
		InventDimID string
		Min         float64
		Max         float64
	}

	minMaxs := []item{}
	cmd := dbflex.From(new(scmmodel.ItemMinMax).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &minMaxs); err != nil {
		return fmt.Errorf("err when get item min max: %s", err.Error())
	}

	type minmax struct {
		Min float64
		Max float64
	}

	mapMinMax := map[string]minmax{}
	for _, m := range minMaxs {
		mapMinMax[m.ItemID+m.SKU+m.InventDimID] = minmax{
			Min: m.Min,
			Max: m.Max,
		}
	}

	for _, trx := range inventTrxs {
		if v, ok := mapItemSpec[trx.SKU]; ok {
			trx.InventDim.Size = v.SpecSizeID
			trx.InventDim.Grade = v.SpecGradeID
			trx.InventDim.VariantID = v.SpecVariantID
		}
		trx.InventDim.Calc()

		// minStock := 0.0
		// maxStock := 0.0
		// minMaxFound := false

		// imms := []scmmodel.ItemMinMax{}
		// h.GetsByFilter(new(scmmodel.ItemMinMax), dbflex.And(
		// 	dbflex.Eq("ItemID", trx.Item.ID),
		// 	dbflex.Eq("SKU", trx.SKU),
		// 	dbflex.Eq("InventoryDimension.InventDimID", trx.InventDim.InventDimID),
		// ), &imms)
		// if len(imms) > 0 {
		// 	// karena ItemMinMax bisa > 1, kita ambil MinStock paling besar dan MaxStock paling kecil
		// 	minMaxFound = true
		// 	maxStock = imms[0].MaxStock
		// 	for _, imm := range imms {
		// 		if imm.MinStock > minStock {
		// 			minStock = imm.MinStock
		// 		}
		// 		if maxStock == 0 || imm.MaxStock < maxStock {
		// 			maxStock = imm.MaxStock
		// 		}
		// 	}
		// }

		if v, ok := mapMinMax[trx.Item.ID+trx.SKU+trx.InventDim.InventDimID]; ok {
			bal := new(scmmodel.ItemBalance)

			balFilter := struct {
				WarehouseIDs []interface{}
				SectionIDs   []interface{}
				SKUs         []interface{}
			}{}

			balFilter.WarehouseIDs = append(balFilter.WarehouseIDs, trx.InventDim.WarehouseID)
			// if trx.InventDim.WarehouseID != "" {
			// 	balFilter.WarehouseIDs = append(balFilter.WarehouseIDs, trx.InventDim.WarehouseID)
			// }

			balFilter.SectionIDs = append(balFilter.SectionIDs, trx.InventDim.SectionID)
			// if trx.InventDim.SectionID != "" {
			// 	balFilter.SectionIDs = append(balFilter.SectionIDs, trx.InventDim.SectionID)
			// }

			balFilter.SKUs = append(balFilter.SKUs, trx.SKU)
			// if trx.SKU != "" {
			// 	balFilter.SKUs = append(balFilter.SKUs, trx.SKU)
			// }

			ibs, _ := NewItemBalanceHub(h).Get(nil, ItemBalanceOpt{
				CompanyID:       trx.CompanyID,
				ItemIDs:         []string{trx.Item.ID},
				InventDim:       trx.InventDim,
				ConsiderSKU:     true,
				BalanceFilter:   balFilter,
				DisableGrouping: true,
			})
			if len(ibs) > 0 {
				bal = ibs[0]
				bal.Calc()
			}

			if trx.Status == scmmodel.ItemConfirmed {
				bal.Qty = bal.Qty + trx.Qty
			} else if trx.Status == scmmodel.ItemPlanned {
				bal.QtyPlanned = bal.QtyPlanned + trx.Qty
			} else if trx.Status == scmmodel.ItemReserved {
				bal.QtyReserved = bal.QtyReserved + trx.Qty
			}

			isDecrease := false
			beforeQtyAvail := bal.QtyAvail
			bal.Calc() // calc qty avail
			if bal.QtyAvail < beforeQtyAvail {
				isDecrease = true
			}

			qtyAvail := bal.QtyAvail
			if sourceTrxType == "Inventory Receive" || sourceTrxType == "Inventory Issuance" {
				qtyAvail = bal.Qty
			}

			if isDecrease {
				// hanya compare min stock bila terjadi pengurangan, agar bila ada penambahan kurang dari min stock masih diperbolehkan (tidak divalidasi)
				if qtyAvail < v.Min {
					errorMsgs = append(errorMsgs, fmt.Sprintf("Item below Min Stock: %s", mapAssignItem[trx.Item.ID+trx.SKU]))
				}
			} else if qtyAvail > v.Max {
				errorMsgs = append(errorMsgs, fmt.Sprintf("Item above Max Stock: %s", mapAssignItem[trx.Item.ID+trx.SKU]))
			}
		}
	}

	if len(errorMsgs) > 0 {
		return fmt.Errorf(strings.Join(errorMsgs, "\n"))
	}

	return nil
}

type DeductOrIncrease string

var (
	Deduct   DeductOrIncrease = "Deduct"
	Increase DeductOrIncrease = "Increase"
)

func PRLinesUpdateRemainingQty(h *datahub.Hub, po *scmmodel.PurchaseOrderJournal, deductOrIncrease DeductOrIncrease) error {
	linePerPR := lo.GroupBy(po.Lines, func(d scmmodel.PurchaseJournalLine) string {
		return d.PRID
	})

	for prID, poLines := range linePerPR {
		pr, e := datahub.GetByID(h, new(scmmodel.PurchaseRequestJournal), prID)
		if e != nil {
			fmt.Printf("WARNING PO Line doesn't have PRID but has ReffNo | PO ID: %s | ReffNo: %s\n", po.ID, strings.Join(po.ReffNo, ", "))
			continue // bypass if PR not found
		}

		// update each PR Lines
		for prLineIdx, prLine := range pr.Lines {
			// get PO Lines with the same ItemID and SKU
			fpolines := lo.Filter(poLines, func(d scmmodel.PurchaseJournalLine, i int) bool {
				return d.ItemID == prLine.ItemID && d.SKU == prLine.SKU && d.SourceLineNo == prLine.LineNo
			})
			if len(fpolines) == 0 {
				continue // bypass if not exist
			}

			// calculate total converted PO Line Qty
			totalConvertedQty := float64(0)
			for _, fpl := range fpolines {
				convQty, e := ConvertUnit(h, fpl.Qty, fpl.UnitID, prLine.UnitID)
				if e != nil {
					return e
				}

				totalConvertedQty = totalConvertedQty + convQty
			}

			if deductOrIncrease == Deduct {
				// validate remaining
				if totalConvertedQty > prLine.RemainingQty {
					return fmt.Errorf("[validate] | qty exceeded, remaining qty: %v %s | total qty input (converted): %v %s",
						prLine.RemainingQty, prLine.UnitID,
						totalConvertedQty, prLine.UnitID,
					)
				}

				pr.Lines[prLineIdx].RemainingQty = prLine.RemainingQty - totalConvertedQty // deduct
			} else if deductOrIncrease == Increase {
				pr.Lines[prLineIdx].RemainingQty = prLine.RemainingQty + totalConvertedQty // increase
			}
		}

		if e := h.Save(pr); e != nil {
			return e
		}
	}

	return nil
}

func AssignItem(h *datahub.Hub, itemIDs, itemSpecIDs []string) (map[string]string, error) {
	items := []tenantcoremodel.Item{}
	err := h.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return nil, fmt.Errorf("error when get item: %s", err.Error())
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, string) {
		return v.ID, v.Name
	})

	specs := []tenantcoremodel.ItemSpec{}
	err = h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemSpecIDs...),
	), &specs)
	if err != nil {
		return nil, fmt.Errorf("error when get item spec: %s", err.Error())
	}

	specVariantIDs := make([]string, len(specs))
	specSizeIDs := make([]string, len(specs))
	specGradeIDs := make([]string, len(specs))
	mapItemSpec := map[string]tenantcoremodel.ItemSpec{}
	for i, is := range specs {
		mapItemSpec[is.ID] = is

		specVariantIDs[i] = is.SpecVariantID
		specSizeIDs[i] = is.SpecSizeID
		specGradeIDs[i] = is.SpecGradeID
	}

	variants := []tenantcoremodel.SpecVariant{}
	err = h.Gets(new(tenantcoremodel.SpecVariant), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &variants)
	if err != nil {
		return nil, fmt.Errorf("error when get spec variant: %s", err.Error())
	}

	mapVariant := lo.Associate(variants, func(v tenantcoremodel.SpecVariant) (string, string) {
		return v.ID, v.Name
	})

	sizes := []tenantcoremodel.SpecSize{}
	err = h.Gets(new(tenantcoremodel.SpecSize), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &sizes)
	if err != nil {
		return nil, fmt.Errorf("error when get spec size: %s", err.Error())
	}

	mapSize := lo.Associate(sizes, func(v tenantcoremodel.SpecSize) (string, string) {
		return v.ID, v.Name
	})

	grades := []tenantcoremodel.SpecGrade{}
	err = h.Gets(new(tenantcoremodel.SpecGrade), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &grades)
	if err != nil {
		return nil, fmt.Errorf("error when get spec grade: %s", err.Error())
	}

	mapGrade := lo.Associate(grades, func(v tenantcoremodel.SpecGrade) (string, string) {
		return v.ID, v.Name
	})

	mapID := map[string][]string{}
	for i, item := range itemIDs {
		index := item + itemSpecIDs[i]
		if itemSpecIDs[i] == "" {
			mapID[index] = append(mapID[index], mapItem[item])
		} else {
			if v, ok := mapItemSpec[itemSpecIDs[i]]; ok {
				if v.OtherName != "" {
					mapID[index] = append(mapID[index], v.OtherName)
				}

				if name, okVariant := mapVariant[v.SpecVariantID]; okVariant {
					mapID[index] = append(mapID[index], name)
				}

				if name, okSize := mapSize[v.SpecSizeID]; okSize {
					mapID[index] = append(mapID[index], name)
				}

				if name, okGrade := mapGrade[v.SpecGradeID]; okGrade {
					mapID[index] = append(mapID[index], name)
				}
			}
		}
	}

	result := map[string]string{}
	for k, v := range mapID {
		result[k] = strings.Join(v, " - ")
	}

	return result, nil
}

func ItemVariantName(h *datahub.Hub, itemIDs, itemSpecIDs []string) (map[string]string, error) {
	items := []tenantcoremodel.Item{}
	err := h.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return nil, fmt.Errorf("error when get item: %s", err.Error())
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, string) {
		return v.ID, v.Name
	})

	specs := []tenantcoremodel.ItemSpec{}
	err = h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemSpecIDs...),
	), &specs)
	if err != nil {
		return nil, fmt.Errorf("error when get item spec: %s", err.Error())
	}

	specVariantIDs := make([]string, len(specs))
	specSizeIDs := make([]string, len(specs))
	specGradeIDs := make([]string, len(specs))
	mapItemSpec := map[string]tenantcoremodel.ItemSpec{}
	for i, is := range specs {
		mapItemSpec[is.ID] = is

		specVariantIDs[i] = is.SpecVariantID
		specSizeIDs[i] = is.SpecSizeID
		specGradeIDs[i] = is.SpecGradeID
	}

	variants := []tenantcoremodel.SpecVariant{}
	err = h.Gets(new(tenantcoremodel.SpecVariant), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &variants)
	if err != nil {
		return nil, fmt.Errorf("error when get spec variant: %s", err.Error())
	}

	mapVariant := lo.Associate(variants, func(v tenantcoremodel.SpecVariant) (string, string) {
		return v.ID, v.Name
	})

	sizes := []tenantcoremodel.SpecSize{}
	err = h.Gets(new(tenantcoremodel.SpecSize), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &sizes)
	if err != nil {
		return nil, fmt.Errorf("error when get spec size: %s", err.Error())
	}

	mapSize := lo.Associate(sizes, func(v tenantcoremodel.SpecSize) (string, string) {
		return v.ID, v.Name
	})

	grades := []tenantcoremodel.SpecGrade{}
	err = h.Gets(new(tenantcoremodel.SpecGrade), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", specVariantIDs...),
	), &grades)
	if err != nil {
		return nil, fmt.Errorf("error when get spec grade: %s", err.Error())
	}

	mapGrade := lo.Associate(grades, func(v tenantcoremodel.SpecGrade) (string, string) {
		return v.ID, v.Name
	})

	mapID := map[string][]string{}
	for i, item := range itemIDs {
		index := item + itemSpecIDs[i]
		if itemSpecIDs[i] == "" {
			mapID[index] = append(mapID[index], mapItem[item])
		} else {
			if v, ok := mapItemSpec[itemSpecIDs[i]]; ok {
				if v.SKU != "" {
					mapID[index] = append([]string{v.SKU}, mapID[index]...)
				}

				if v.OtherName != "" {
					mapID[index] = append(mapID[index], v.OtherName)
				}

				if name, okVariant := mapVariant[v.SpecVariantID]; okVariant {
					mapID[index] = append(mapID[index], name)
				}

				if name, okSize := mapSize[v.SpecSizeID]; okSize {
					mapID[index] = append(mapID[index], name)
				}

				if name, okGrade := mapGrade[v.SpecGradeID]; okGrade {
					mapID[index] = append(mapID[index], name)
				}
			}
		}
	}

	result := map[string]string{}
	for k, v := range mapID {
		result[k] = strings.Join(v, " - ")
	}

	return result, nil
}

func MongoNotFound(err error) bool {
	if err != nil {
		if err != nil && (err == mongo.ErrNoDocuments || strings.ToLower(err.Error()) == "eof") {
			return true
		}
	}

	return false
}

func ValidatePostingApproval(approvalStatus string) string {
	journalStatus := ""
	switch approvalStatus {
	case "", "DRAFT":
		journalStatus = string(ficomodel.JournalStatusDraft)
	case "PENDING":
		journalStatus = string(ficomodel.JournalStatusSubmitted)
		// case "APPROVED", "READY":
		// 	journalStatus = string(ficomodel.JournalStatusApproved)
		// case "POSTED":
		// 	journalStatus = string(ficomodel.JournalStatusPosted)
	}

	return journalStatus
}
