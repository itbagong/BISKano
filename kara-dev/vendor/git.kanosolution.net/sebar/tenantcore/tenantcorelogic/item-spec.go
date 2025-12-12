package tenantcorelogic

import (
	"errors"
	"fmt"
	"sync"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type ItemSpecEngine struct{}

type ItemGetsDetailRes struct {
	tenantcoremodel.ItemSpec
	Description string
}

func (o *ItemSpecEngine) GetsDetail(ctx *kaos.Context, itemSpecIDs []string) ([]ItemGetsDetailRes, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	specs := []tenantcoremodel.ItemSpec{}
	if e := h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", codekit.ToInterfaceArray(itemSpecIDs)...)), &specs); e != nil {
		return nil, e
	}

	res := []ItemGetsDetailRes{}
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

		res = append(res, ItemGetsDetailRes{
			ItemSpec:    spec,
			Description: desc,
		})
	}

	return res, nil
}

func (o *ItemSpecEngine) GetsInfo(ctx *kaos.Context, _ *interface{}) ([]ItemGetsDetailRes, error) {
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
	res := []ItemGetsDetailRes{}
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

		res = append(res, ItemGetsDetailRes{
			ItemSpec:    spec,
			Description: desc,
		})
	}

	return res, nil
}

func (o *ItemSpecEngine) SaveMultiple(ctx *kaos.Context, payload []tenantcoremodel.ItemSpec) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if len(payload) == 0 {
		return nil, errors.New("missing: payload")
	}

	itemID := payload[0].ItemID
	h.DeleteByFilter(new(tenantcoremodel.ItemSpec), dbflex.Eq("ItemID", itemID))

	for _, spec := range payload {
		if e := h.GetByID(new(tenantcoremodel.ItemSpec), spec.ID); e != nil {
			if e := h.Insert(&spec); e != nil {
				return nil, errors.New("error insert Item Spec: " + e.Error())
			}
		} else {
			if e := h.Save(&spec); e != nil {
				return nil, errors.New("error update Item Spec: " + e.Error())
			}
		}
	}

	return payload, nil
}
