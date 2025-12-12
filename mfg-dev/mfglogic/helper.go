package mfglogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/leekchan/accounting"
	"github.com/sebarcode/codekit"
)

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

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO"
	}
	return coID
}

func GetCompanyAndUserIDFromContext(ctx *kaos.Context) (string, string, error) {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	userID := sebar.GetUserIDFromCtx(ctx)

	if coID == "" || userID == "" {
		return "", "", fmt.Errorf("Session expired")
	}

	return coID, userID, nil
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}

func GetItemSpecDescription(h *datahub.Hub, itemSpecID string) (string, error) {
	spec := new(tenantcoremodel.ItemSpec)
	if e := h.GetByID(spec, itemSpecID); e != nil {
		return "", e
	}

	variants := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
	sizes := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize))
	grades := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade))

	separator := "-"
	desc := spec.SKU

	if spec.SpecVariantID != "" {
		if v, e := variants.Get(spec.SpecVariantID); e != nil {
			desc += fmt.Sprintf("%s%s", separator, v.Name)
		}
	}

	if spec.SpecSizeID != "" {
		if v, e := sizes.Get(spec.SpecSizeID); e != nil {
			desc += fmt.Sprintf("%s%s", separator, v.Name)
		}
	}

	if spec.SpecGradeID != "" {
		if v, e := grades.Get(spec.SpecGradeID); e != nil {
			desc += fmt.Sprintf("%s%s", separator, v.Name)
		}
	}

	return desc, nil
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

func FormatDateTime(paramDateTime *time.Time) string {
	if paramDateTime == nil {
		return ""
	}

	return paramDateTime.Format("02 January 2006 15:04:05")
}

func FormatFloatDecimal2(number float64) string {
	return fmt.Sprintf("%.2f", number)
}

func FormatNumberNoDecimal(number float64) string {
	return fmt.Sprintf("%.0f", number)
}
