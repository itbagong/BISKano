package hcmlogic

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO"
	}
	return coID
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}
