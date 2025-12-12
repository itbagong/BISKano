package tenantcorelogic

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoreconfig"
	"github.com/sebarcode/codekit"
)

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

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		coID = tenantcoreconfig.Config.DefaultCompanyID
	}
	if coID == "" {
		coID = "DEMO"
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
