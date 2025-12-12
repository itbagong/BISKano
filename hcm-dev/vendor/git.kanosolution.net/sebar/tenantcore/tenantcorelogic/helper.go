package tenantcorelogic

import (
	"bytes"
	"html/template"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoreconfig"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type GeneralRequest struct {
	Select []string // TODO: currently unused
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

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		coID = tenantcoreconfig.Config.DefaultCompanyID
	}
	return coID
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
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

func ParseStringTemplate(source string, data codekit.M) (string, error) {
	w := bytes.NewBufferString("")
	tt, e := template.New("tmp").Parse(source)
	if e != nil {
		return source, e
	}

	e = tt.Execute(w, data)
	if e != nil {
		return source, e
	}

	return w.String(), nil
}

func ItemVariantName(h *datahub.Hub, itemID string, itemSpecID ...string) (itemVariantName string) {
	separator := " - "
	texts := []string{}
	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	specORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
	specVariantORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
	specSizeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize))
	specGradeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade))

	itemData, _ := itemORM.Get(itemID)
	texts = append(texts, itemData.Name)

	if len(itemSpecID) == 0 {
		return itemData.Name
	}

	spec, _ := specORM.Get(itemSpecID[0])
	if spec.SKU != "" {
		texts = append([]string{spec.SKU}, texts...)
	}

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

	return strings.Join(texts, separator)
}

func ItemVariantNameV2(h *datahub.Hub, itemID string, itemSpecID ...string) (itemVariantName string) {
	separator := " - "
	texts := []string{}
	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	specORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
	specVariantORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
	specSizeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize))
	specGradeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade))

	itemData, _ := itemORM.Get(itemID)
	texts = append(texts, itemData.Name)

	if len(itemSpecID) == 0 {
		return itemData.Name
	}

	spec, _ := specORM.Get(itemSpecID[0])
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

	return strings.Join(texts, separator)
}

type HelperItemVariantNameByMapParam struct {
	itemMap        map[string]*tenantcoremodel.Item
	itemSpecMap    map[string]*tenantcoremodel.ItemSpec
	specVariantMap map[string]*tenantcoremodel.SpecVariant
	specSizeMap    map[string]*tenantcoremodel.SpecSize
	specGradeMap   map[string]*tenantcoremodel.SpecGrade
	itemORM        *sebar.MapRecord[*tenantcoremodel.Item]
	itemSpecORM    *sebar.MapRecord[*tenantcoremodel.ItemSpec]
	specVariantORM *sebar.MapRecord[*tenantcoremodel.SpecVariant]
	specSizeORM    *sebar.MapRecord[*tenantcoremodel.SpecSize]
	specGradeORM   *sebar.MapRecord[*tenantcoremodel.SpecGrade]
}

func ItemVariantNameByMap(h *datahub.Hub, itemID, itemSpecID string, param HelperItemVariantNameByMapParam) (itemVariantName string) {
	separator := " - "
	texts := []string{}

	// #1 - Item
	itemData, exist := param.itemMap[itemID]
	if !exist {
		itemData, _ = param.itemORM.Get(itemID)
	}

	texts = append(texts, itemData.Name)

	if itemSpecID == "" {
		return itemData.Name
	}

	// #2 - ItemSpec
	spec, exist := param.itemSpecMap[itemSpecID]
	if !exist {
		spec, _ = param.itemSpecORM.Get(itemSpecID)
	}

	if spec != nil && spec.SKU != "" {
		texts = append([]string{spec.SKU}, texts...)
	}

	if spec != nil && spec.OtherName != "" {
		texts = append(texts, spec.OtherName)
	}

	// #3 - SpecVariant
	dataSpecVariant, exist := param.specVariantMap[spec.SpecVariantID]
	if !exist && spec.SpecVariantID != "" {
		dataSpecVariant, _ = param.specVariantORM.Get(spec.SpecVariantID)
	}

	if dataSpecVariant != nil && dataSpecVariant.Name != "" {
		texts = append(texts, dataSpecVariant.Name)
	}

	// #4 - SpecSize
	dataSpecSize, exist := param.specSizeMap[spec.SpecSizeID]
	if !exist && spec.SpecSizeID != "" {
		dataSpecSize, _ = param.specSizeORM.Get(spec.SpecSizeID)
	}

	if dataSpecSize != nil && dataSpecSize.Name != "" {
		texts = append(texts, dataSpecSize.Name)
	}

	// #5 - SpecGrade
	dataSpecGrade, exist := param.specGradeMap[spec.SpecGradeID]
	if !exist && spec.SpecGradeID != "" {
		dataSpecGrade, _ = param.specGradeORM.Get(spec.SpecGradeID)
	}

	if dataSpecGrade != nil && dataSpecGrade.Name != "" {
		texts = append(texts, dataSpecGrade.Name)
	}

	return strings.Join(texts, separator)
}

func ItemVariantNameByMapV2(h *datahub.Hub, itemID, itemSpecID string, param HelperItemVariantNameByMapParam) (itemVariantName string) {
	separator := " - "
	texts := []string{}

	// #2 - ItemSpec
	spec, exist := param.itemSpecMap[itemSpecID]
	if !exist {
		spec, _ = param.itemSpecORM.Get(itemSpecID)
	}

	if spec != nil && spec.OtherName != "" {
		texts = append(texts, spec.OtherName)
	}

	// #3 - SpecVariant
	dataSpecVariant, exist := param.specVariantMap[spec.SpecVariantID]
	if !exist && spec.SpecVariantID != "" {
		dataSpecVariant, _ = param.specVariantORM.Get(spec.SpecVariantID)
	}

	if dataSpecVariant != nil && dataSpecVariant.Name != "" {
		texts = append(texts, dataSpecVariant.Name)
	}

	// #4 - SpecSize
	dataSpecSize, exist := param.specSizeMap[spec.SpecSizeID]
	if !exist && spec.SpecSizeID != "" {
		dataSpecSize, _ = param.specSizeORM.Get(spec.SpecSizeID)
	}

	if dataSpecSize != nil && dataSpecSize.Name != "" {
		texts = append(texts, dataSpecSize.Name)
	}

	// #5 - SpecGrade
	dataSpecGrade, exist := param.specGradeMap[spec.SpecGradeID]
	if !exist && spec.SpecGradeID != "" {
		dataSpecGrade, _ = param.specGradeORM.Get(spec.SpecGradeID)
	}

	if dataSpecGrade != nil && dataSpecGrade.Name != "" {
		texts = append(texts, dataSpecGrade.Name)
	}

	return strings.Join(texts, separator)
}

// CombineSlices menggabungkan beberapa slice menjadi satu slice tunggal.
func CombineSlices[T any](slices ...[]T) []T {
	var combinedSlice []T
	for _, slice := range slices {
		combinedSlice = append(combinedSlice, slice...)
	}
	return combinedSlice
}
