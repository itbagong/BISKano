package tenantcorelogic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type ItemSpecEngine struct{}

type ItemSpecGetsDetailRes struct {
	tenantcoremodel.ItemSpec
	Description string
}

func (o *ItemSpecEngine) GetsDetail(ctx *kaos.Context, itemSpecIDs []string) ([]ItemSpecGetsDetailRes, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	specs := []tenantcoremodel.ItemSpec{}
	if e := h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", codekit.ToInterfaceArray(itemSpecIDs)...)), &specs); e != nil {
		return nil, e
	}

	res := []ItemSpecGetsDetailRes{}
	for _, spec := range specs {
		itemName := ""
		variantName := ""
		sizeName := ""
		gradeName := ""
		var wg sync.WaitGroup

		if spec.ItemID != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sd := new(tenantcoremodel.Item)
				h.GetByID(sd, spec.ItemID)
				itemName = sd.Name
			}()
		}

		if spec.SpecVariantID != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sd := new(tenantcoremodel.SpecVariant)
				h.GetByID(sd, spec.SpecVariantID)
				variantName = sd.Name
			}()
		}

		if spec.SpecSizeID != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sd := new(tenantcoremodel.SpecSize)
				h.GetByID(sd, spec.SpecSizeID)
				sizeName = sd.Name
			}()
		}

		if spec.SpecGradeID != "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sd := new(tenantcoremodel.SpecGrade)
				h.GetByID(sd, spec.SpecGradeID)
				gradeName = sd.Name
			}()
		}

		wg.Wait()

		separator := "-"
		desc := itemName
		if variantName != "" {
			desc += fmt.Sprintf("%s%s", separator, variantName)
		}
		if sizeName != "" {
			desc += fmt.Sprintf("%s%s", separator, sizeName)
		}
		if gradeName != "" {
			desc += fmt.Sprintf("%s%s", separator, gradeName)
		}

		res = append(res, ItemSpecGetsDetailRes{
			ItemSpec:    spec,
			Description: desc,
		})
	}

	return res, nil
}

func (o *ItemSpecEngine) GetsInfo(ctx *kaos.Context, _ *interface{}) ([]ItemSpecGetsDetailRes, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	payload := GetURLQueryParams(ctx)

	fs := []*dbflex.Filter{}

	if payload["_id"] != "" {
		fs = append(fs, dbflex.Eq("_id", payload["_id"]))
	}

	if payload["ItemID"] != "" {
		fs = append(fs, dbflex.Eq("ItemID", payload["ItemID"]))
	}

	// get each ids for filters (based on keyword typed in from UI)
	variants := []tenantcoremodel.SpecVariant{}
	variantM := map[string]tenantcoremodel.SpecVariant{}
	sizes := []tenantcoremodel.SpecSize{}
	sizeM := map[string]tenantcoremodel.SpecSize{}
	grades := []tenantcoremodel.SpecGrade{}
	gradeM := map[string]tenantcoremodel.SpecGrade{}

	if payload["Keyword"] != "" {
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			h.GetsByFilter(new(tenantcoremodel.SpecVariant), dbflex.Contains("Name", payload["Keyword"]), &variants)
			if len(variants) > 0 {
				ids := lo.Map(variants, func(d tenantcoremodel.SpecVariant, i int) string {
					return d.ID
				})
				fs = append(fs, dbflex.In("SpecVariantID", codekit.ToInterfaceArray(ids)...))

				variantM = lo.SliceToMap(variants, func(d tenantcoremodel.SpecVariant) (string, tenantcoremodel.SpecVariant) {
					return d.ID, d
				})
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			h.GetsByFilter(new(tenantcoremodel.SpecSize), dbflex.Contains("Name", payload["Keyword"]), &sizes)
			if len(sizes) > 0 {
				ids := lo.Map(sizes, func(d tenantcoremodel.SpecSize, i int) string {
					return d.ID
				})
				fs = append(fs, dbflex.In("SpecSizeID", codekit.ToInterfaceArray(ids)...))

				sizeM = lo.SliceToMap(sizes, func(d tenantcoremodel.SpecSize) (string, tenantcoremodel.SpecSize) {
					return d.ID, d
				})
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			h.GetsByFilter(new(tenantcoremodel.SpecGrade), dbflex.Contains("Name", payload["Keyword"]), &grades)
			if len(grades) > 0 {
				ids := lo.Map(grades, func(d tenantcoremodel.SpecGrade, i int) string {
					return d.ID
				})
				fs = append(fs, dbflex.In("SpecGradeID", codekit.ToInterfaceArray(ids)...))

				gradeM = lo.SliceToMap(grades, func(d tenantcoremodel.SpecGrade) (string, tenantcoremodel.SpecGrade) {
					return d.ID, d
				})
			}
		}()

		wg.Wait()
	}

	param := dbflex.NewQueryParam()
	if len(fs) > 0 {
		param = param.SetWhere(dbflex.And(fs...))
	}

	specs := []tenantcoremodel.ItemSpec{}
	if e := h.Gets(new(tenantcoremodel.ItemSpec), param, &specs); e != nil {
		return nil, e
	}

	// parse results and format name
	res := []ItemSpecGetsDetailRes{}
	for _, spec := range specs {
		separator := "-"
		desc := spec.SKU

		if spec.SpecVariantID != "" {
			if v, ok := variantM[spec.SpecVariantID]; ok {
				desc += fmt.Sprintf("%s%s", separator, v.Name)
			} else {
				v := new(tenantcoremodel.SpecVariant)
				h.GetByID(v, spec.SpecVariantID)
				desc += lo.Ternary(v.ID != "", fmt.Sprintf("%s%s", separator, v.Name), "")
			}
		}

		if spec.SpecSizeID != "" {
			if v, ok := sizeM[spec.SpecSizeID]; ok {
				desc += fmt.Sprintf("%s%s", separator, v.Name)
			} else {
				v := new(tenantcoremodel.SpecSize)
				h.GetByID(v, spec.SpecSizeID)
				desc += lo.Ternary(v.ID != "", fmt.Sprintf("%s%s", separator, v.Name), "")
			}
		}

		if spec.SpecGradeID != "" {
			if v, ok := gradeM[spec.SpecGradeID]; ok {
				desc += fmt.Sprintf("%s%s", separator, v.Name)
			} else {
				v := new(tenantcoremodel.SpecGrade)
				h.GetByID(v, spec.SpecGradeID)
				desc += lo.Ternary(v.ID != "", fmt.Sprintf("%s%s", separator, v.Name), "")
			}
		}

		res = append(res, ItemSpecGetsDetailRes{
			ItemSpec:    spec,
			Description: desc,
		})
	}

	return res, nil
}

type GetsInfoByIDRequest struct {
	ID string
}

func (o *ItemSpecEngine) GetsInfoByID(ctx *kaos.Context, payload *GetsInfoByIDRequest) (*ItemSpecGetsDetailRes, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	itemSpec, e := datahub.GetByID(h, new(tenantcoremodel.ItemSpec), payload.ID)
	if e != nil {
		return nil, e
	}

	res := ItemSpecGetsDetailRes{}
	desc := itemSpec.SKU
	separator := "-"

	if itemSpec.SpecVariantID != "" {
		variant, _ := datahub.GetByID(h, new(tenantcoremodel.SpecVariant), itemSpec.SpecVariantID)
		desc += fmt.Sprintf("%s%s", separator, variant.Name)
	}

	if itemSpec.SpecSizeID != "" {
		size, _ := datahub.GetByID(h, new(tenantcoremodel.SpecSize), itemSpec.SpecSizeID)
		desc += fmt.Sprintf("%s%s", separator, size.Name)
	}

	if itemSpec.SpecGradeID != "" {
		grade, _ := datahub.GetByID(h, new(tenantcoremodel.SpecGrade), itemSpec.SpecGradeID)
		desc += fmt.Sprintf("%s%s", separator, grade.Name)
	}

	res.ItemSpec = *itemSpec
	res.Description = desc
	return &res, nil
}

func (o *ItemSpecEngine) SaveMultiple(ctx *kaos.Context, payload []tenantcoremodel.ItemSpec) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if len(payload) == 0 {
		return nil, errors.New("missing: payload")
	}

	userLogin := sebar.GetUserIDFromCtx(ctx)
	itemID := payload[0].ItemID
	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))

	// check SKU usage
	exItemSpecs := []tenantcoremodel.ItemSpec{}
	h.GetsByFilter(new(tenantcoremodel.ItemSpec), dbflex.Eq("ItemID", itemID), &exItemSpecs)

	deletedItemSpecs := []tenantcoremodel.ItemSpec{}
	payloadMap := lo.SliceToMap(payload, func(spec tenantcoremodel.ItemSpec) (string, bool) {
		return spec.ID, true
	})

	for _, exSpec := range exItemSpecs {
		if _, found := payloadMap[exSpec.ID]; !found {
			deletedItemSpecs = append(deletedItemSpecs, exSpec)
		}
	}

	// cari item yang baru, yang belum ada di exItemSpecs, utk validasi duplikat SKU
	exItemSpecsMap := lo.SliceToMap(exItemSpecs, func(d tenantcoremodel.ItemSpec) (string, bool) {
		return d.ID, true
	})

	newItemSpecSKUs := []string{}
	for _, spec := range payload {
		if _, found := exItemSpecsMap[spec.ID]; !found {
			newItemSpecSKUs = append(newItemSpecSKUs, spec.SKU)
		}
	}

	// cek duplikat SKU
	duplicatedSKUs := []tenantcoremodel.ItemSpec{}
	h.GetsByFilter(new(tenantcoremodel.ItemSpec), dbflex.In("SKU", newItemSpecSKUs...), &duplicatedSKUs)
	if len(duplicatedSKUs) > 0 {
		// duplicate SKU: "SKU-001" di Item "BEARING-001" | "SKU-002" di Item "BEARING-002"
		return nil, fmt.Errorf("duplicate SKU: %s", strings.Join(lo.Map(duplicatedSKUs, func(d tenantcoremodel.ItemSpec, i int) string {
			item, _ := itemORM.Get(d.ItemID)
			return fmt.Sprintf("\"%s\" di Item \"%s\"", d.SKU, lo.Ternary(item.Name != "", item.Name, d.ItemID))
		}), ", "))
	}

	if len(deletedItemSpecs) > 0 {
		deletedItemSpecIDs := lo.Map(deletedItemSpecs, func(d tenantcoremodel.ItemSpec, i int) string {
			return d.ID
		})

		ev, _ := ctx.DefaultEvent()
		if ev == nil {
			return false, nil
		}

		resp := new(UsageCheckResponse)
		if e := ev.Publish("/v1/scm/item/spec/usage-check", &struct{ SpecIDs []string }{SpecIDs: deletedItemSpecIDs}, resp, nil); e != nil {
			return false, fmt.Errorf("failed to check SKU usage: %s", e.Error())
		}

		if resp.IsUsed {
			return false, fmt.Errorf("deleted Specification is already used journals")
		}
	}

	logParams := []tenantcoremodel.LogParam{}

	// h.DeleteByFilter(new(tenantcoremodel.ItemSpec), dbflex.Eq("ItemID", itemID))

	// instead of delete all, delete only deleted ones
	if len(deletedItemSpecs) > 0 {
		deletedIDs := lo.Map(deletedItemSpecs, func(d tenantcoremodel.ItemSpec, i int) string { return d.ID })
		h.DeleteByFilter(new(tenantcoremodel.ItemSpec), dbflex.In("_id", deletedIDs...))

		logParams = append(logParams, lo.Map(deletedItemSpecs, func(d tenantcoremodel.ItemSpec, i int) tenantcoremodel.LogParam {
			return tenantcoremodel.LogParam{
				Hub:           h,
				Menu:          string(tenantcoremodel.LogMenuItemSpec),
				Action:        string(tenantcoremodel.LogActionDelete),
				TransactionID: d.ID,
				Name:          d.SKU,
				UserLogin:     userLogin,
			}
		})...)
	}

	for _, spec := range payload {
		if e := h.GetByID(new(tenantcoremodel.ItemSpec), spec.ID); e != nil {
			if e := h.Insert(&spec); e != nil {
				return nil, errors.New("error insert Item Spec: " + e.Error())
			}

			logParams = append(logParams, tenantcoremodel.LogParam{
				Hub:           h,
				Menu:          string(tenantcoremodel.LogMenuItemSpec),
				Action:        string(tenantcoremodel.LogActionCreate),
				TransactionID: spec.ID,
				Name:          spec.SKU,
				UserLogin:     userLogin,
			})
		} else {
			if e := h.Save(&spec); e != nil {
				return nil, errors.New("error update Item Spec: " + e.Error())
			}

			logParams = append(logParams, tenantcoremodel.LogParam{
				Hub:           h,
				Menu:          string(tenantcoremodel.LogMenuItemSpec),
				Action:        string(tenantcoremodel.LogActionUpdate),
				TransactionID: spec.ID,
				Name:          spec.SKU,
				UserLogin:     userLogin,
			})
		}
	}

	new(tenantcoremodel.Log).AddMultiple(logParams)

	return payload, nil
}

type ItemSpecSearchParam struct {
	dbflex.QueryParam
	Keyword string
	// Skip    int
	// Take    int
}

type ItemSpecSearchResult struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `grid:"hide" bson:"_id" json:"_id"`
	ItemID            string    `label:"ItemID"`
	ItemName          string    `label:"Item Name"`
	SKU               string    `label:"SKU"`
	OtherName         string    `label:"Other Name"`
	SpecVariantID     string    `grid:"hide"`
	SpecSizeID        string    `grid:"hide"`
	SpecGradeID       string    `grid:"hide"`
	SpecID            string    `grid:"hide"`
	Variant           string    `label:"Variant"`
	Size              string    `label:"Size"`
	Grade             string    `label:"Grade"`
	IsActive          bool      `label:"Is Active"`
	Created           time.Time `label:"Created"`
	LastUpdate        time.Time `label:"Last Update"`
}

func (o *ItemSpecEngine) Search(ctx *kaos.Context, payload *ItemSpecSearchParam) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	if payload.Skip < 0 {
		payload.Skip = 0
	}

	if payload.Take <= 0 {
		payload.Take = 10
	}

	results := []ItemSpecSearchResult{}
	totalCount, err := o.search(h, payload, &results)
	if err != nil {
		return nil, err
	}

	res := codekit.M{
		"data":  results,
		"count": totalCount,
	}

	return res, nil
}

func (o *ItemSpecEngine) Gets(ctx *kaos.Context, payload *ItemSpecSearchParam) (codekit.M, error) {
	return o.Search(ctx, payload)
}

func (o *ItemSpecEngine) Download() func(h *datahub.Hub) ([]codekit.M, error) {
	return func(h *datahub.Hub) ([]codekit.M, error) {
		results := []codekit.M{}

		_, err := o.search(h, nil, &results)
		if err != nil {
			return nil, err
		}

		return results, nil
	}
}

func (o *ItemSpecEngine) search(h *datahub.Hub, payload *ItemSpecSearchParam, results interface{}) (count int, err error) {
	if payload == nil {
		payload = &ItemSpecSearchParam{}
	}

	// TODO: bisa lebih di optimasi lagi dengan lookup filter, bila payload.Keyword is provided
	pipe := []bson.M{}
	e := Deserialize(fmt.Sprintf(`
    [
		{"$lookup":{
			"from":"Items",
			"localField":"ItemID",
			"foreignField":"_id",
			"as":"items"
		}},
		{"$unwind":{"path":"$items","preserveNullAndEmptyArrays": false}},
		
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
			"Grade":"$spec_grades.Name",
			"Size":"$spec_sizes.Name",
			"Variant":"$spec_variants.Name"
		}}
	]
    `), &pipe)
	if e != nil {
		return 0, e
	}

	if payload.Keyword != "" {
		payload.Keyword = strings.TrimSpace(payload.Keyword)
		escapedKeyword := regexp.QuoteMeta(payload.Keyword)
		pipe = append(pipe, bson.M{"$match": bson.M{"$or": []bson.M{
			{"items._id": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"SKU": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"OtherName": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"ItemName": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"Grade": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"Size": bson.M{"$regex": escapedKeyword, "$options": "i"}},
			{"Variant": bson.M{"$regex": escapedKeyword, "$options": "i"}},
		}}})
	}

	pipe = append(pipe, bson.M{"$sort": bson.M{"ItemName": 1, "SKU": 1}})

	pipeSkipTake := make([]bson.M, len(pipe))
	copy(pipeSkipTake, pipe)

	if payload.Skip >= 0 && payload.Take > 0 {
		pipeSkipTake = append(pipeSkipTake, bson.M{"$skip": payload.Skip}, bson.M{"$limit": payload.Take})
	}

	cmd := dbflex.From(new(tenantcoremodel.ItemSpec).TableName()).Command("pipe", pipeSkipTake)
	if _, e := h.Populate(cmd, results); e != nil {
		return 0, fmt.Errorf("error when fetching data : %s", e)
	}

	countCmd := dbflex.From(new(tenantcoremodel.ItemSpec).TableName()).Command("pipe", append(pipe, bson.M{"$count": "totalCount"}))
	countResult := []codekit.M{}
	if _, e := h.Populate(countCmd, &countResult); e != nil {
		return 0, fmt.Errorf("error when counting data : %s", e)
	}

	var totalCount int
	if len(countResult) > 0 {
		totalCount = countResult[0].GetInt("totalCount")
	} else {
		totalCount = 0
	}

	return totalCount, nil
}
