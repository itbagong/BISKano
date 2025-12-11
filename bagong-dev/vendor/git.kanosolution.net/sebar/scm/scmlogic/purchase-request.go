package scmlogic

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type PurchaseRequestEngine struct{}

type PurchaseRequestGetLinesRequest struct {
	VendorID               string
	Site                   string
	DeliveryTo             string   // WarehouseID
	Statuses               []string // empty -> POSTED only
	ReffNo                 string
	PRRequestor            string
	SortPurchaseRequestIDs []string
	Skip                   int
	Take                   int
}

type PurchaseRequestGetLinesResponse struct {
	scmmodel.PurchaseJournalLine
	Header       PRHeader
	PRID         string
	PRDate       *time.Time
	WarehouseID  string
	RemainQty    float64 // from item balance - qty planned
	VendorID     string
	VendorName   string
	SourceLineNo int
}

type PurchaseRequestGetLinesResponseV1 struct {
	scmmodel.PurchaseJournalLine
	Header       PRHeaderLine
	PRID         string
	PRDate       *time.Time
	WarehouseID  string
	RemainQty    float64 // from item balance - qty planned
	VendorID     string
	VendorName   string
	SourceLineNo int
}

type PRHeader struct {
	scmmodel.PurchaseRequestJournal
	DueDate time.Time
}

type PRHeaderLine struct {
	PurchaseRequestJournal scmmodel.PurchaseRequestJournalWithLine
	DueDate                time.Time
}

func (o *PurchaseRequestEngine) GetLines(ctx *kaos.Context, payload *PurchaseRequestGetLinesRequest) ([]PurchaseRequestGetLinesResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	filters := []*dbflex.Filter{}

	if payload.VendorID != "" {
		filters = append(filters, dbflex.Eq("VendorID", payload.VendorID))
	}

	if payload.DeliveryTo != "" {
		filters = append(filters, dbflex.Eq("Location.WarehouseID", payload.DeliveryTo))
	}

	if payload.Site != "" {
		filters = append(filters, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Site)))
	}

	if len(payload.Statuses) > 0 {
		filters = append(filters, dbflex.In("Status", payload.Statuses...))
	} else {
		filters = append(filters, dbflex.Eq("Status", ficomodel.JournalStatusPosted))
	}

	if payload.ReffNo != "" {
		filters = append(filters, dbflex.Contains("_id", payload.ReffNo))
	}

	if payload.PRRequestor != "" {
		filters = append(filters, dbflex.Eq("Requestor", payload.PRRequestor))
	}

	var f *dbflex.Filter
	if len(filters) > 0 {
		f = dbflex.And(filters...)
	}

	prs := []scmmodel.PurchaseRequestJournal{}
	if e := h.GetsByFilter(new(scmmodel.PurchaseRequestJournal), f, &prs); e != nil {
		return nil, e
	}

	skuIDs := make([]string, 0)
	itemIDs := make([]string, 0)
	for _, pr := range prs {
		for _, l := range pr.Lines {
			skuIDs = append(skuIDs, l.SKU)
			itemIDs = append(itemIDs, l.ItemID)
		}
	}

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
		dbflex.In("_id", skuIDs...),
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

	res := []PurchaseRequestGetLinesResponse{}
	for _, pr := range prs {
		lines := lo.Map(pr.Lines, func(item scmmodel.PurchaseJournalLine, index int) PurchaseRequestGetLinesResponse {
			prHeader := pr
			prHeader.ReffNo = []string{pr.ID}
			prHeader.Lines = []scmmodel.PurchaseJournalLine{}
			return PurchaseRequestGetLinesResponse{
				PurchaseJournalLine: item,
				Header:              PRHeader{PurchaseRequestJournal: prHeader, DueDate: *pr.PRDate},
				PRID:                pr.ID,
				PRDate:              pr.PRDate,
				WarehouseID:         pr.WarehouseID,
				VendorID:            pr.VendorID,
				VendorName:          pr.VendorName,
				SourceLineNo:        item.LineNo,
			}
		})

		// tambah query find remaining qty > 0
		prg := []PurchaseRequestGetLinesResponse{}
		for li, l := range lines {
			spec := mapItemSpec[l.SKU]

			texts := []string{}
			if v, ok := mapItem[spec.ItemID]; ok {
				texts = append(texts, v)
			}

			if spec.SKU != "" {
				texts = append([]string{spec.SKU}, texts...)
			}

			if spec.OtherName != "" {
				texts = append(texts, spec.OtherName)
			}

			if v, ok := mapVariant[spec.SpecVariantID]; ok {
				texts = append(texts, v)
			}

			if v, ok := mapSize[spec.SpecSizeID]; ok {
				texts = append(texts, v)
			}

			if v, ok := mapGrade[spec.SpecGradeID]; ok {
				texts = append(texts, v)
			}

			specName := strings.Join(texts, " - ")
			lines[li].Text = lo.Ternary(specName != "", specName, mapItem[l.ItemID])
			if l.InventJournalLine.RemainingQty > 0 {
				prg = append(prg, lines[li])
			}
		}

		res = append(res, prg...)
	}

	sort.SliceStable(res, func(i, j int) bool {
		return lo.Contains(payload.SortPurchaseRequestIDs, res[i].PRID) && !lo.Contains(payload.SortPurchaseRequestIDs, res[j].PRID)
	})

	return res, nil
}

func (o *PurchaseRequestEngine) GetLinesV1(ctx *kaos.Context, payload *PurchaseRequestGetLinesRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	filters := []*dbflex.Filter{}
	pipeFilters := []bson.M{}

	if payload.VendorID != "" {
		filters = append(filters, dbflex.Eq("VendorID", payload.VendorID))
		pipeFilters = append(pipeFilters, bson.M{"VendorID": payload.VendorID})
	}

	if payload.DeliveryTo != "" {
		filters = append(filters, dbflex.Eq("Location.WarehouseID", payload.DeliveryTo))
		pipeFilters = append(pipeFilters, bson.M{"Location.WarehouseID": payload.DeliveryTo})
	}

	if payload.Site != "" {
		filters = append(filters, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Site)))
		pipeFilters = append(pipeFilters, bson.M{"$and": []bson.M{
			{"Dimension.Key": "Site"},
			{"Dimension.Value": payload.Site},
		}})
	}

	if len(payload.Statuses) > 0 {
		filters = append(filters, dbflex.In("Status", payload.Statuses...))
		pipeFilters = append(pipeFilters, bson.M{"Status": bson.M{"$in": payload.Statuses}})
	} else {
		filters = append(filters, dbflex.Eq("Status", ficomodel.JournalStatusPosted))
		pipeFilters = append(pipeFilters, bson.M{"Status": ficomodel.JournalStatusPosted})
	}

	if payload.ReffNo != "" {
		filters = append(filters, dbflex.Contains("_id", payload.ReffNo))
		pipeFilters = append(pipeFilters, bson.M{"_id": payload.ReffNo})
	}

	if payload.PRRequestor != "" {
		filters = append(filters, dbflex.Eq("Requestor", payload.PRRequestor))
		pipeFilters = append(pipeFilters, bson.M{"Requestor": payload.PRRequestor})
	}

	pipe := []bson.M{}

	if len(pipeFilters) > 0 {
		pipe = append(pipe, bson.M{"$match": bson.M{"$and": pipeFilters}})
	}

	pipeUnwind := []bson.M{}
	e := Deserialize(fmt.Sprintf(`
    [
		{
			"$unwind": {
				"path": "$Lines",
    			"preserveNullAndEmptyArrays": true
			}
		}
    ]
    `), &pipeUnwind)
	if e != nil {
		return nil, e
	}

	pipe = append(pipe, pipeUnwind...)

	pipe = append(pipe, bson.M{"$match": bson.M{"Lines.InventJournalLine.RemainingQty": bson.M{"$gt": 0}}})

	if len(payload.SortPurchaseRequestIDs) > 0 {
		pipe = append(pipe, bson.M{"$addFields": bson.M{
			"SortPriority": bson.M{"$concat": []interface{}{"_id", "|", bson.M{"$toString": "$Lines.InventJournalLine.LineNo"}}},
		}})

		pipe = append(pipe, bson.M{"$addFields": bson.M{
			"SortPriority": bson.M{"$cond": []interface{}{
				bson.M{"$in": []interface{}{"$SortPriority", payload.SortPurchaseRequestIDs}},
				// bson.M{"$arrayElemAt": []interface{}{sortOrder, bson.M{"$indexOfArray": []interface{}{payload.SortSourceJournalIDs, "$SourceJournalID"}}}},
				bson.M{"$indexOfArray": []interface{}{payload.SortPurchaseRequestIDs, "$SortPriority"}},
				len(payload.SortPurchaseRequestIDs),
			}},
		}})
	}
	pipeSkipTake := make([]bson.M, len(pipe))
	copy(pipeSkipTake, pipe)
	pipeSkipTake = append(pipeSkipTake, bson.M{"$skip": payload.Skip}, bson.M{"$limit": payload.Take})

	prs := []scmmodel.PurchaseRequestJournalWithLine{}
	cmd := dbflex.From(new(scmmodel.PurchaseRequestJournal).TableName()).Command("pipe", pipeSkipTake)
	if _, e := h.Populate(cmd, &prs); e != nil {
		return nil, fmt.Errorf("error when fetching data : %s:\n%s", e, codekit.JsonStringIndent(pipeSkipTake, "\t"))
	}

	countCmd := dbflex.From(new(scmmodel.PurchaseRequestJournal).TableName()).Command("pipe", append(pipe, bson.M{"$count": "totalCount"}))
	countResult := []codekit.M{}
	if _, e := h.Populate(countCmd, &countResult); e != nil {
		return nil, fmt.Errorf("error when counting data : %s", e)
	}

	var totalCount int
	if len(countResult) > 0 {
		totalCount = countResult[0].GetInt("totalCount")
	} else {
		totalCount = 0
	}

	// get item and specs
	skuIDs := make([]string, 0)
	itemIDs := make([]string, 0)
	for _, pr := range prs {
		skuIDs = append(skuIDs, pr.Lines.SKU)
		itemIDs = append(itemIDs, pr.Lines.ItemID)
	}

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
		dbflex.In("_id", skuIDs...),
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

	res := []PurchaseRequestGetLinesResponseV1{}
	for _, pr := range prs {
		spec := mapItemSpec[pr.Lines.SKU]

		texts := []string{}
		if v, ok := mapItem[spec.ItemID]; ok {
			texts = append(texts, v)
		}

		if spec.SKU != "" {
			texts = append([]string{spec.SKU}, texts...)
		}

		if spec.OtherName != "" {
			texts = append(texts, spec.OtherName)
		}

		if v, ok := mapVariant[spec.SpecVariantID]; ok {
			texts = append(texts, v)
		}

		if v, ok := mapSize[spec.SpecSizeID]; ok {
			texts = append(texts, v)
		}

		if v, ok := mapGrade[spec.SpecGradeID]; ok {
			texts = append(texts, v)
		}

		specName := strings.Join(texts, " - ")
		pr.Lines.Text = lo.Ternary(specName != "", specName, mapItem[pr.Lines.ItemID])

		// if pr.Lines.InventJournalLine.RemainingQty > 0 {

		prHeader := pr
		prHeader.ReffNo = []string{pr.ID}
		// prHeader.Lines = []scmmodel.PurchaseJournalLine{}
		res = append(res, PurchaseRequestGetLinesResponseV1{
			PurchaseJournalLine: pr.Lines,
			Header:              PRHeaderLine{PurchaseRequestJournal: pr, DueDate: *pr.PRDate},
			PRID:                pr.ID,
			PRDate:              pr.PRDate,
			WarehouseID:         pr.WarehouseID,
			VendorID:            pr.VendorID,
			VendorName:          pr.VendorName,
			SourceLineNo:        pr.Lines.LineNo,
		})
		// }
	}

	response := codekit.M{
		"data":  res,
		"count": totalCount,
	}

	return response, nil
}

func (o *PurchaseRequestEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.PurchaseRequestJournal{}
	resFinal := codekit.M{
		"count": len(res),
		"data":  res,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return resFinal, fmt.Errorf("missing: connection")
	}

	if req == nil {
		return resFinal, fmt.Errorf("missing: payload")
	}

	//clear param
	keyword := ""
	if req.Where != nil {
		filter := req.Where
		if filter.Op == dbflex.OpAnd {
			clearItems := []*dbflex.Filter{}
			for _, filterItem := range filter.Items {
				if filterItem.Field == "Keyword" {
					keyword = filterItem.Value.(string)
				} else {
					clearItems = append(clearItems, filterItem)
				}
			}

			filter.Items = clearItems
		}

		if len(filter.Items) > 0 {
			req.Where = filter
		} else {
			req.Where = nil
		}
	}

	//get posting profile purchase request
	jtPurchaseRequests := []scmmodel.PurchaseRequestJournalType{}
	e := h.Gets(new(scmmodel.PurchaseRequestJournalType), dbflex.NewQueryParam(), &jtPurchaseRequests)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppPurchaseRequests := lo.Map(jtPurchaseRequests, func(jt scmmodel.PurchaseRequestJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppPurchaseRequests[0])
	if len(sourceIDs) > 0 {
		if req.Where != nil {
			filter := req.Where
			if filter.Op == dbflex.OpAnd {
				clearItems := []*dbflex.Filter{}
				for _, filterItem := range filter.Items {
					if filterItem.Field == "Keyword" {
						keyword = filterItem.Value.(string)
					} else {
						if filterItem.Op == dbflex.OpOr {
							filterItem.Items = append(filterItem.Items, dbflex.In("_id", sourceIDs...))
						}

						clearItems = append(clearItems, filterItem)
					}
				}

				filter.Items = clearItems
			}

			if len(filter.Items) > 0 {
				req.Where = filter
			} else {
				req.Where = nil
			}
		}
	}

	fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)
	parm := req
	if len(fs) == 1 {
		parm = CombineQueryParam(req, dbflex.NewQueryParam().SetWhere(fs[0]))
	} else if len(fs) > 1 {
		parm = CombineQueryParam(req, dbflex.NewQueryParam().SetWhere(dbflex.And(fs...)))
	}

	if parm.Where != nil {
		if parm.Where.Op == dbflex.OpAnd {
			if len(parm.Where.Items) == 1 {
				if parm.Where.Items[0].Op == dbflex.OpOr {
					parm.Where = parm.Where.Items[0]
				}
			}
		}
	}

	// Parse "TrxDate" in query's "Where" clause
	// Convert string to RFC3339 time format if present
	// Ensures proper formatting for database query
	if parm.Where != nil && parm.Where.Items != nil {
		for i, item := range parm.Where.Items {
			if item.Field == "TrxDate" {
				if strValue, ok := item.Value.(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, strValue); err == nil {
						parm.Where.Items[i].Value = parsedTime
					}
				}
			}
		}
	}

	e = h.Gets(new(scmmodel.PurchaseRequestJournal), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.PurchaseRequestJournal).TableName())
	if parm != nil && parm.Where != nil {
		cmd.Where(parm.Where)
	}
	connIdx, conn, err := h.GetConnection()
	if err == nil {
		defer h.CloseConnection(connIdx, conn)
		resFinal.Set("count", conn.Cursor(cmd, nil).Count())
	}

	// fill Reff PO: PO mana yang sudah menggunakan PR ini
	prIDs := lo.Map(res, func(item scmmodel.PurchaseRequestJournal, index int) string {
		return item.ID
	})

	poReffs := []scmmodel.PurchaseOrderJournal{}
	h.GetsByFilter(new(scmmodel.PurchaseOrderJournal), dbflex.In("ReffNo", prIDs...), &poReffs)

	poReffsMap := map[string]string{}
	for _, poReff := range poReffs {
		for _, ref := range poReff.ReffNo {
			poReffsMap[ref] = poReff.ID
		}
	}

	res = lo.Map(res, func(d scmmodel.PurchaseRequestJournal, index int) scmmodel.PurchaseRequestJournal {
		if poID, ok := poReffsMap[d.ID]; ok {
			d.POReff = append(d.POReff, poID)
		}
		return d
	})

	resFinal.Set("data", res)
	return resFinal, nil
}

type PurchaseRequestSyncJournalStatusRequest struct {
	PurchaseRequestID string
}

func (p *PurchaseRequestEngine) SyncJournalStatus(ctx *kaos.Context, req *PurchaseRequestSyncJournalStatusRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "NOK", fmt.Errorf("missing: connection")
	}

	if req == nil {
		return "NOK", fmt.Errorf("missing: payload")
	}

	//get purchase order journal
	pr, err := datahub.GetByID(h, new(scmmodel.PurchaseRequestJournal), req.PurchaseRequestID)
	if err != nil && err != mongo.ErrNoDocuments {
		return "NOK", fmt.Errorf("Error when get purchase request journal with error: %s", err.Error())
	}

	//get posting profile approval
	param := dbflex.NewQueryParam()
	param.SetSort("-LastUpdate")
	param.SetWhere(dbflex.Eq("SourceID", req.PurchaseRequestID))

	postingApproval, err := datahub.GetByParm(h, new(ficomodel.PostingApproval), param)
	if err != nil {
		return "NOK", fmt.Errorf("Error when get posting approval with error: %s", err.Error())
	}

	journalStatus := ValidatePostingApproval(postingApproval.Status)

	if pr.Status != ficomodel.JournalStatus(journalStatus) && journalStatus != "" {
		pr.Status = ficomodel.JournalStatus(journalStatus)
		err = datahub.Update(h, pr)
		if err != nil {
			return "NOK", fmt.Errorf("Error when update journal status with error: %s", err.Error())
		}
	}

	return "OK", nil
}
