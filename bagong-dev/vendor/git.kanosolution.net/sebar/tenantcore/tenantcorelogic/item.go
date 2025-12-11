package tenantcorelogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

const (
	SeparatorID string = "~~"
)

type ItemEngine struct{}

type ItemGetsDetailRes struct {
	ID       string
	ItemID   string
	SKU      string
	Text     string
	Item     tenantcoremodel.Item
	ItemSpec *tenantcoremodel.ItemSpec
}

func MWPreAssignItem(id string, isSerial bool) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		if h == nil {
			return false, errors.New("missing: connection")
		}

		splitItem := strings.Split(id, "~~")
		itemID := splitItem[0]
		itemSpecID := splitItem[1]

		separator := " - "
		texts := []string{}

		itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		specORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
		specVariantORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
		specSizeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize))
		specGradeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade))

		itemData, _ := itemORM.Get(itemID)
		texts = append(texts, itemData.Name)

		if itemSpecID == "" {
			m.Set("_id", itemData.Name)
			serde.Serde(m, payload)
			return true, nil
		}

		spec, _ := specORM.Get(itemSpecID)
		// if spec.SKU != "" {
		// 	texts = append([]string{spec.SKU}, texts...)
		// }

		if spec.OtherName != "" {
			texts = append(texts, spec.OtherName)
		}

		if data, _ := specVariantORM.Get(spec.SpecVariantID); data != nil && data.Name != "" {
			texts = append(texts, data.Name)
		}

		if data, _ := specSizeORM.Get(spec.SpecSizeID); data != nil && data.Name != "" {
			texts = append(texts, data.Name)
		}

		if data, _ := specGradeORM.Get(spec.SpecGradeID); data != nil && data.Name != "" {
			texts = append(texts, data.Name)
		}

		m.Set("_id", strings.Join(texts, separator))
		serde.Serde(m, payload)

		return true, nil
	}
}

func (o *ItemEngine) GetsDetail(ctx *kaos.Context, req *GeneralRequest) ([]ItemGetsDetailRes, error) {
	reqQuery := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if req == nil {
		req = &GeneralRequest{
			Skip: 0,
			Take: 10,
		}
	}

	reqItemID := ""
	reqItemSpecID := ""
	reqText := ""
	reqItemGroupID := ""
	reqExcludeItemGroupIDs := []string{}
	if req.Where != nil {
		// handle filter first level
		if req.Where.Value != "" {
			switch req.Where.Field {
			case "_id":
				reqItemID, reqItemSpecID = getsDetailSplitPayloadID(codekit.ToString(req.Where.Value))
			case "Text":
				reqText = codekit.ToString(req.Where.Value)
			case "ItemGroupID":
				reqItemGroupID = codekit.ToString(req.Where.Value)
			case "ExcludeItemGroupID":
				reqExcludeItemGroupIDs = strings.Split(codekit.ToString(req.Where.Value), ",")
			}
		}

		// handle filter by second level
		for _, item := range req.Where.Items {
			if val, ok := item.Value.([]interface{}); ok && len(val) > 0 {
				switch item.Field {
				case "_id":
					reqItemID, reqItemSpecID = getsDetailSplitPayloadID(codekit.ToString(req.Where.Value))
				case "Text":
					reqText = codekit.ToString(val[0])
				case "ItemGroupID":
					reqItemGroupID = codekit.ToString(val[0])
				case "ExcludeItemGroupID":
					reqExcludeItemGroupIDs = lo.Map(val, func(d interface{}, i int) string { return codekit.ToString(d) })
				}
			}
		}
	}

	if reqQuery["_id"] != "" {
		reqItemID, reqItemSpecID = getsDetailSplitPayloadID(reqQuery["_id"])
	}

	if reqQuery["Text"] != "" {
		reqText = reqQuery["Text"]
	}

	if reqQuery["ItemGroupID"] != "" {
		reqItemGroupID = reqQuery["ItemGroupID"]
	}

	if reqQuery["ExcludeItemGroupID"] != "" {
		reqExcludeItemGroupIDs = strings.Split(reqQuery["ExcludeItemGroupID"], ",")
	}

	// handle if _id="undefined"
	if reqItemID == "undefined" {
		reqItemID = ""
	}

	specParam := GetSpecParam{
		reqItemID:              reqItemID,
		reqItemGroupID:         reqItemGroupID,
		reqItemSpecID:          reqItemSpecID,
		reqText:                reqText,
		reqExcludeItemGroupIDs: reqExcludeItemGroupIDs,
	}

	if reqItemID != "" {
		specParam.ignoreSpecIsActive = true // ignore is active if showing only specific item
	}

	// #1 - Item
	items, specByItems, e := getSpecsByItem(h, req, specParam)
	if e != nil {
		return nil, e
	}

	// kalo search nya "" tampilan cuman dari item table aja untuk performa saja
	if reqItemID == "" && reqText == "" && req.Take == 1 {
		return toGetsDetailResult(h, items, nil, req), nil
	}

	// kalo cuman menampilkan 1 item tanpa spec (tidak perlu search ke table lain)
	if req.Take == 1 && reqItemID != "" && reqItemSpecID == "" && reqText == "" {
		return toGetsDetailResult(h, items, nil, req), nil
	}

	// #2 - ItemSpec
	itemSpecs, e := getSpecsBySpec(h, req, specParam)
	if e != nil {
		return nil, e
	}

	// #3 - SpecVariant
	specByVariants, e := getSpecsByVariant(h, req, specParam)
	if e != nil {
		return nil, e
	}

	// #4 - SpecSize
	specBySizes, e := getSpecsBySize(h, req, specParam)
	if e != nil {
		return nil, e
	}

	// #5 - SpecGrade
	specByGrades, e := getSpecsByGrade(h, req, specParam)
	if e != nil {
		return nil, e
	}

	allSpecs := [][]*tenantcoremodel.ItemSpec{
		specByItems,
		itemSpecs,
		specByVariants,
		specBySizes,
		specByGrades,
	}

	// merge uniquely all specs from item, spec itself, and variant
	mergedSpecs := lo.UniqBy(CombineSlices(allSpecs...), func(d *tenantcoremodel.ItemSpec) string {
		return d.ID
	})

	// filter items that doesn't have any spec
	specItemIDs := lo.Uniq(lo.Map(mergedSpecs, func(d *tenantcoremodel.ItemSpec, i int) string {
		return d.ItemID
	}))

	items = lo.Filter(items, func(d *tenantcoremodel.Item, i int) bool {
		return lo.Contains(specItemIDs, d.ID) == false
	})

	// filter merged specs when ItemSpecID is given
	if reqItemSpecID != "" {
		// filter to get one data only for displaying existing data
		mergedSpecs = lo.Filter(mergedSpecs, func(d *tenantcoremodel.ItemSpec, i int) bool {
			return d.ID == reqItemSpecID
		})
	}

	return toGetsDetailResult(h, items, mergedSpecs, req), nil
}

type SKUGetsDetailRes struct {
	ID       string
	ItemID   string
	Text     string
	ItemSpec *tenantcoremodel.ItemSpec
}

func (o *ItemEngine) GetsSkuDetailByItem(ctx *kaos.Context, req *interface{}) ([]SKUGetsDetailRes, error) {
	reqQuery := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if _, ok := reqQuery["ItemID"]; !ok {
		return nil, errors.New("Missing payload Item ID")
	}

	itemID := reqQuery["ItemID"]
	specs := []tenantcoremodel.ItemSpec{}

	h.GetsByFilter(new(tenantcoremodel.ItemSpec), dbflex.Eq("ItemID", itemID), &specs)
	results := lo.Map(specs, func(spec tenantcoremodel.ItemSpec, index int) SKUGetsDetailRes {
		res := SKUGetsDetailRes{}
		res.ID = spec.ID
		res.ItemID = itemID
		text := []string{spec.SKU}

		//get variant
		variant, _ := datahub.GetByID(h, new(tenantcoremodel.SpecVariant), spec.SpecVariantID)
		if variant.Name != "" {
			text = append(text, variant.Name)
		}

		// size
		size, _ := datahub.GetByID(h, new(tenantcoremodel.SpecSize), spec.SpecSizeID)
		if size.Name != "" {
			text = append(text, size.Name)
		}

		// get grade
		grade, _ := datahub.GetByID(h, new(tenantcoremodel.SpecGrade), spec.SpecGradeID)
		if grade.Name != "" {
			text = append(text, grade.Name)
		}

		res.Text = strings.Join(text, " - ")
		res.ItemSpec = &spec

		return res
	})

	return results, nil
}

func Deserialize(str string, res interface{}) error {
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		return fmt.Errorf("%s:\n%s", err, str)
	}
	return err
}

func (o *ItemEngine) Gets(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	items := []tenantcoremodel.CustomItemDownload{}
	res := codekit.M{}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: connection")
	}

	idx, conn, e := h.GetConnection()
	if e != nil {
		return res, e
	}
	defer h.CloseConnection(idx, conn)
	pipes := []codekit.M{}

	e = Deserialize(`[{"$lookup":{
        "from":"Items",
        "localField":"ItemID",
        "foreignField":"_id",
        "as":"items"
}},
{"$unwind":{"path":"$items","preserveNullAndEmptyArrays": true}},    
{"$lookup":{
    "from":"SpecGrade",
    "localField":"SpecGradeID",
    "foreignField":"_id",
    "as":"spec_grades"
}},
{"$unwind":{"path":"$spec_grades","preserveNullAndEmptyArrays": true}},

{"$lookup":{
    "from":"SpecSize",
    "localField":"SpecSizeID",
    "foreignField":"_id",
    "as":"spec_sizes"
}},
{"$unwind":{"path":"$spec_sizes","preserveNullAndEmptyArrays": true}},


{"$lookup":{
    "from":"SpecVariant",
    "localField":"SpecVariantID",
    "foreignField":"_id",
    "as":"spec_variants"
}},
{"$unwind":{"path":"$spec_variants","preserveNullAndEmptyArrays": true}},

{"$set":{
    "ItemName":"$items.Name",
    "SpecGradeName":"$spec_grades.Name",
    "SpecSizeName":"$spec_sizes.Name",
    "SpecVariantName":"$spec_variants.Name"
}}]`, &pipes)
	if req.Take > 0 {
		pipes = append(pipes, codekit.M{"$limit": req.Take})
	}

	if req.Skip > 0 {
		pipes = append(pipes, codekit.M{"$skip": req.Skip})
	}
	if e != nil {
		return res, nil
	}

	cmd := dbflex.From(new(tenantcoremodel.ItemSpec).TableName()).Command("pipe", pipes)
	_, e = h.Populate(cmd, &items)
	if e != nil {
		return res, e
	}

	res = codekit.M{
		"data":  items,
		"count": len(items),
	}

	return res, nil
}

func (o *ItemEngine) Download() func(h *datahub.Hub) ([]codekit.M, error) {
	return func(h *datahub.Hub) ([]codekit.M, error) {
		items := []codekit.M{}

		if h == nil {
			return items, errors.New("missing: connection")
		}

		idx, conn, e := h.GetConnection()
		if e != nil {
			return items, e
		}
		defer h.CloseConnection(idx, conn)
		pipes := []codekit.M{}

		e = Deserialize(`[{"$lookup":{
			"from":"Items",
			"localField":"ItemID",
			"foreignField":"_id",
			"as":"items"
	}},
	{"$unwind":{"path":"$items","preserveNullAndEmptyArrays": true}},    
	{"$lookup":{
		"from":"SpecGrade",
		"localField":"SpecGradeID",
		"foreignField":"_id",
		"as":"spec_grades"
	}},
	{"$unwind":{"path":"$spec_grades","preserveNullAndEmptyArrays": true}},

	{"$lookup":{
		"from":"SpecSize",
		"localField":"SpecSizeID",
		"foreignField":"_id",
		"as":"spec_sizes"
	}},
	{"$unwind":{"path":"$spec_sizes","preserveNullAndEmptyArrays": true}},


	{"$lookup":{
		"from":"SpecVariant",
		"localField":"SpecVariantID",
		"foreignField":"_id",
		"as":"spec_variants"
	}},
	{"$unwind":{"path":"$spec_variants","preserveNullAndEmptyArrays": true}},

	{"$set":{
		"ItemName":"$items.Name",
		"SpecGradeName":"$spec_grades.Name",
		"SpecSizeName":"$spec_sizes.Name",
		"SpecVariantName":"$spec_variants.Name"
	}}]`, &pipes)
		if e != nil {
			return items, nil
		}

		cmd := dbflex.From(new(tenantcoremodel.ItemSpec).TableName()).Command("pipe", pipes)
		_, e = h.Populate(cmd, &items)
		if e != nil {
			return items, e
		}

		return items, nil
	}
}

type GetSpecParam struct {
	reqItemID              string
	reqItemGroupID         string
	reqItemSpecID          string
	reqText                string
	reqExcludeItemGroupIDs []string
	ignoreSpecIsActive     bool
}

func getSpecsByItem(h *datahub.Hub, req *GeneralRequest, parm GetSpecParam) ([]*tenantcoremodel.Item, []*tenantcoremodel.ItemSpec, error) {
	qparam := req.GetQueryParam().SetWhere(nil)
	filters := []*dbflex.Filter{}

	if parm.reqItemID != "" {
		filters = append(filters, dbflex.Contains("_id", parm.reqItemID))
	}

	if parm.reqText != "" {
		filters = append(filters, dbflex.Contains("Name", parm.reqText))
	}

	if parm.reqItemGroupID != "" {
		if len(filters) > 0 {
			filters = []*dbflex.Filter{dbflex.And(dbflex.Eq("ItemGroupID", parm.reqItemGroupID), dbflex.Or(filters...))}
		} else {
			filters = []*dbflex.Filter{dbflex.Eq("ItemGroupID", parm.reqItemGroupID)}
		}
	}

	if len(parm.reqExcludeItemGroupIDs) > 0 {
		if len(filters) > 0 {
			filters = []*dbflex.Filter{dbflex.And(dbflex.Nin("ItemGroupID", codekit.ToInterfaceArray(parm.reqExcludeItemGroupIDs)...), dbflex.Or(filters...))}
		} else {
			filters = []*dbflex.Filter{dbflex.Nin("ItemGroupID", codekit.ToInterfaceArray(parm.reqExcludeItemGroupIDs)...)}
		}
	}

	if len(filters) > 0 {
		qparam.SetWhere(dbflex.Or(filters...))
	}

	items := []*tenantcoremodel.Item{}
	if e := h.Gets(new(tenantcoremodel.Item), qparam, &items); e != nil {
		return nil, nil, e
	}

	itemIDs := lo.Map(items, func(d *tenantcoremodel.Item, i int) string {
		return d.ID
	})

	where := dbflex.And(dbflex.Eq("IsActive", true), dbflex.In("ItemID", itemIDs...))
	if parm.ignoreSpecIsActive {
		where = dbflex.In("ItemID", itemIDs...)
	}

	specByItems := []*tenantcoremodel.ItemSpec{}
	specByItemsParm := req.GetQueryParam().SetWhere(where)
	if e := h.Gets(new(tenantcoremodel.ItemSpec), specByItemsParm, &specByItems); e != nil {
		return nil, nil, e
	}

	return items, specByItems, nil
}

func getSpecsBySpec(h *datahub.Hub, req *GeneralRequest, parm GetSpecParam) ([]*tenantcoremodel.ItemSpec, error) {
	itemSpecs, e := GetDataByGeneralRequest(h, new(tenantcoremodel.ItemSpec), req, []GetDataByGRParam{
		{Field: "_id", Value: parm.reqItemSpecID},
		{Field: "SKU", Value: parm.reqText},
		{Field: "OtherName", Value: parm.reqText},
	}...)
	if e != nil {
		return nil, e
	}

	itemSpecs, e = filterSpecByItemGroupID(h, itemSpecs, parm)
	if e != nil {
		return nil, e
	}

	if !parm.ignoreSpecIsActive {
		// filter spec by IsActive only
		itemSpecs = lo.Filter(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) bool {
			return d.IsActive
		})
	}

	return itemSpecs, nil
}

func getSpecsByVariant(h *datahub.Hub, req *GeneralRequest, parm GetSpecParam) ([]*tenantcoremodel.ItemSpec, error) {
	specVariants, e := GetDataByGeneralRequest(h, new(tenantcoremodel.SpecVariant), req, []GetDataByGRParam{
		{Field: "Name", Value: parm.reqText},
	}...)
	if e != nil {
		return nil, e
	}

	varIDs := lo.Map(specVariants, func(d *tenantcoremodel.SpecVariant, i int) string {
		return d.ID
	})

	where := dbflex.And(dbflex.Eq("IsActive", true), dbflex.In("SpecVariantID", varIDs...))
	if parm.ignoreSpecIsActive {
		where = dbflex.In("SpecVariantID", varIDs...)
	}

	res := []*tenantcoremodel.ItemSpec{}
	if e := h.GetsByFilter(new(tenantcoremodel.ItemSpec), where, &res); e != nil {
		return nil, e
	}

	res, e = filterSpecByItemGroupID(h, res, parm)
	if e != nil {
		return nil, e
	}

	return res, nil
}

func getSpecsBySize(h *datahub.Hub, req *GeneralRequest, parm GetSpecParam) ([]*tenantcoremodel.ItemSpec, error) {
	specVariants, e := GetDataByGeneralRequest(h, new(tenantcoremodel.SpecSize), req, []GetDataByGRParam{
		{Field: "Name", Value: parm.reqText},
	}...)
	if e != nil {
		return nil, e
	}

	varIDs := lo.Map(specVariants, func(d *tenantcoremodel.SpecSize, i int) string {
		return d.ID
	})

	where := dbflex.And(dbflex.Eq("IsActive", true), dbflex.In("SpecSizeID", varIDs...))
	if parm.ignoreSpecIsActive {
		where = dbflex.In("SpecSizeID", varIDs...)
	}

	res := []*tenantcoremodel.ItemSpec{}
	if e := h.GetsByFilter(new(tenantcoremodel.ItemSpec), where, &res); e != nil {
		return nil, e
	}

	res, e = filterSpecByItemGroupID(h, res, parm)
	if e != nil {
		return nil, e
	}

	return res, nil
}

func getSpecsByGrade(h *datahub.Hub, req *GeneralRequest, parm GetSpecParam) ([]*tenantcoremodel.ItemSpec, error) {
	specVariants, e := GetDataByGeneralRequest(h, new(tenantcoremodel.SpecGrade), req, []GetDataByGRParam{
		{Field: "Name", Value: parm.reqText},
	}...)
	if e != nil {
		return nil, e
	}

	varIDs := lo.Map(specVariants, func(d *tenantcoremodel.SpecGrade, i int) string {
		return d.ID
	})

	where := dbflex.And(dbflex.Eq("IsActive", true), dbflex.In("SpecGradeID", varIDs...))
	if parm.ignoreSpecIsActive {
		where = dbflex.In("SpecGradeID", varIDs...)
	}

	res := []*tenantcoremodel.ItemSpec{}
	if e := h.GetsByFilter(new(tenantcoremodel.ItemSpec), where, &res); e != nil {
		return nil, e
	}

	res, e = filterSpecByItemGroupID(h, res, parm)
	if e != nil {
		return nil, e
	}

	return res, nil
}

func filterSpecByItemGroupID(h *datahub.Hub, itemSpecs []*tenantcoremodel.ItemSpec, parm GetSpecParam) ([]*tenantcoremodel.ItemSpec, error) {
	if len(itemSpecs) == 0 || parm.reqItemGroupID == "" {
		return itemSpecs, nil
	}

	itemIDs := lo.Uniq(lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.ItemID }))

	filters := []*dbflex.Filter{dbflex.In("_id", itemIDs...)}
	if len(parm.reqExcludeItemGroupIDs) > 0 {
		filters = append(filters, dbflex.Nin("ItemGroupID", codekit.ToInterfaceArray(parm.reqExcludeItemGroupIDs)...))
	}

	if parm.reqItemGroupID != "" {
		filters = append(filters, dbflex.Eq("ItemGroupID", parm.reqItemGroupID))
	}

	filtItems := []*tenantcoremodel.Item{}
	if e := h.GetsByFilter(new(tenantcoremodel.Item), dbflex.And(filters...), &filtItems); e != nil {
		return nil, e
	}

	filtItemIDs := lo.Map(filtItems, func(d *tenantcoremodel.Item, i int) string { return d.ID })
	itemSpecs = lo.Filter(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) bool {
		return lo.Contains(filtItemIDs, d.ItemID)
	})

	return itemSpecs, nil
}

func toGetsDetailResult(h *datahub.Hub, items []*tenantcoremodel.Item, itemSpecs []*tenantcoremodel.ItemSpec, req *GeneralRequest) []ItemGetsDetailRes {
	resItems := lo.Map(items, func(d *tenantcoremodel.Item, i int) ItemGetsDetailRes {
		txts := []string{d.Name}

		return ItemGetsDetailRes{
			ID:     d.ID,
			ItemID: d.ID,
			Text:   strings.Join(txts, " - "),
			Item:   *d,
		}
	})

	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))

	param := HelperItemVariantNameByMapParam{
		itemORM:        itemORM,
		itemSpecORM:    sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec)),
		specVariantORM: sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant)),
		specSizeORM:    sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize)),
		specGradeORM:   sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade)),
	}

	// build param Map
	param.itemMap = getDataItemInMap(h, new(tenantcoremodel.Item),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.ItemID }),
	)

	param.itemSpecMap = getDataItemInMap(h, new(tenantcoremodel.ItemSpec),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.ID }),
	)

	param.specVariantMap = getDataItemInMap(h, new(tenantcoremodel.SpecVariant),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecVariantID }),
	)

	param.specSizeMap = getDataItemInMap(h, new(tenantcoremodel.SpecSize),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecSizeID }),
	)

	param.specGradeMap = getDataItemInMap(h, new(tenantcoremodel.SpecGrade),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecGradeID }),
	)

	resSpecs := []ItemGetsDetailRes{}
	lo.ForEach(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) {
		item, _ := itemORM.Get(d.ItemID)
		if item != nil && item.ID != "" && item.IsActive == true {
			resSpecs = append(resSpecs, ItemGetsDetailRes{
				ID:       fmt.Sprintf("%s%s%s", d.ItemID, SeparatorID, d.ID),
				ItemID:   d.ItemID,
				SKU:      d.ID,
				Text:     ItemVariantNameByMap(h, d.ItemID, d.ID, param),
				Item:     *item,
				ItemSpec: d,
			})
		}
	})

	res := append(resItems, resSpecs...)

	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})

	if len(res) > req.Take {
		res = res[:req.Take]
	}

	return res
}

func toGetsDetailResultV2(h *datahub.Hub, items []*tenantcoremodel.Item, itemSpecs []*tenantcoremodel.ItemSpec, req *GeneralRequest) []ItemGetsDetailRes {
	resItems := lo.Map(items, func(d *tenantcoremodel.Item, i int) ItemGetsDetailRes {
		txts := []string{d.Name}

		return ItemGetsDetailRes{
			ID:     d.ID,
			ItemID: d.ID,
			Text:   strings.Join(txts, " - "),
			Item:   *d,
		}
	})

	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))

	param := HelperItemVariantNameByMapParam{
		itemORM:        itemORM,
		itemSpecORM:    sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec)),
		specVariantORM: sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant)),
		specSizeORM:    sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize)),
		specGradeORM:   sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade)),
	}

	// build param Map
	param.itemMap = getDataItemInMap(h, new(tenantcoremodel.Item),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.ItemID }),
	)

	param.itemSpecMap = getDataItemInMap(h, new(tenantcoremodel.ItemSpec),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.ID }),
	)

	param.specVariantMap = getDataItemInMap(h, new(tenantcoremodel.SpecVariant),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecVariantID }),
	)

	param.specSizeMap = getDataItemInMap(h, new(tenantcoremodel.SpecSize),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecSizeID }),
	)

	param.specGradeMap = getDataItemInMap(h, new(tenantcoremodel.SpecGrade),
		lo.Map(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) string { return d.SpecGradeID }),
	)

	resSpecs := []ItemGetsDetailRes{}
	lo.ForEach(itemSpecs, func(d *tenantcoremodel.ItemSpec, i int) {
		item, _ := itemORM.Get(d.ItemID)
		if item != nil && item.ID != "" && item.IsActive == true {
			resSpecs = append(resSpecs, ItemGetsDetailRes{
				ID:       fmt.Sprintf("%s%s%s", d.ItemID, SeparatorID, d.ID),
				ItemID:   d.ItemID,
				SKU:      d.ID,
				Text:     ItemVariantNameByMapV2(h, d.ItemID, d.ID, param),
				Item:     *item,
				ItemSpec: d,
			})
		}
	})

	res := append(resItems, resSpecs...)

	sort.Slice(res, func(i, j int) bool {
		return res[i].Text < res[j].Text
	})

	if len(res) > req.Take {
		res = res[:req.Take]
	}

	return res
}

type GetDataByGRParam struct {
	Field string
	Value interface{}
}

func GetDataByGeneralRequest[T orm.DataModel](h *datahub.Hub, any T, req *GeneralRequest, param ...GetDataByGRParam) ([]T, error) {
	qparam := req.GetQueryParam().SetWhere(nil)
	filters := []*dbflex.Filter{}

	for _, p := range param {
		if val, _ := p.Value.(string); p.Value == nil || val == "" {
			continue // bypass empty
		}

		switch val := p.Value.(type) {
		case string:
			filters = append(filters, dbflex.Contains(p.Field, val))
		default:
			filters = append(filters, dbflex.Eq(p.Field, val))
		}
	}

	if len(filters) > 0 {
		qparam.SetWhere(dbflex.Or(filters...))
	}

	items := []T{}
	if e := h.Gets(any, qparam, &items); e != nil {
		return nil, e
	}

	return items, nil
}

func getsDetailSplitPayloadID(id string) (reqItemID, reqItemSpecID string) {
	if id == "" {
		return
	}

	reqIDs := strings.Split(id, SeparatorID)
	reqItemID = reqIDs[0]
	if len(reqIDs) > 1 {
		reqItemSpecID = reqIDs[1]
	}

	return
}

func getDataItemInMap[T orm.DataModel](h *datahub.Hub, any T, ids []string) map[string]T {
	itemDatas := []T{}
	h.GetsByFilter(any, dbflex.In("_id", ids...), &itemDatas)
	res := lo.SliceToMap(itemDatas, func(d T) (string, T) {
		id := ""
		reflector.From(d).GetTo("ID", &id)
		return id, d
	})

	return res
}
