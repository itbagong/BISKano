package shelogic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/codekit"
)

type SidakLogic struct {
}

type SidakRequest struct {
	CompanyID string
	IDs       []string
	DateFrom  *time.Time
	DateTo    *time.Time
	Dimension []string

	Skip  int
	Take  int
	Where *dbflex.Filter
}

func (obj *SidakLogic) Gets(ctx *kaos.Context, sr *SidakRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := dbflex.NewQueryParam()
	paramEmp := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{}
	filtersEmp := []*dbflex.Filter{}

	if sr.Skip >= 0 && sr.Take > 0 {
		param = param.SetSkip(sr.Skip).SetTake(sr.Take)
	}

	if sr.Where != nil {
		if len(sr.Where.Items) > 0 {
			vVal := ""
			for _, val := range sr.Where.Items {
				fieldVal := val.Field
				opVal := val.Op
				if opVal == "$contains" {
					aInterface := val.Value.([]interface{})
					aString := make([]string, len(aInterface))
					for i, v := range aInterface {
						aString[i] = v.(string)
					}
					if len(aString) > 0 {
						filtersEmp = append(filtersEmp, dbflex.Contains(fieldVal, aString[0]))
					}
					vVal = aString[0]
				}
			}
			paramEmp.SetWhere(dbflex.Or(filtersEmp...))

			dataIDs := []string{}
			dataEmployee := []tenantcoremodel.Employee{}
			if e := h.Gets(new(tenantcoremodel.Employee), paramEmp, &dataEmployee); e != nil {
				return nil, e
			}

			if len(dataEmployee) > 0 {
				for _, vale := range dataEmployee {
					dataIDs = append(dataIDs, vale.ID)
				}

				filters = append(filters, dbflex.In("EmployeeID", dataIDs...))
			}

			// add filter dimension site
			payload := struct {
				Name string
			}{
				Name: vVal,
			}

			// save data asset bagong
			ev, _ := ctx.DefaultEvent()

			if ev == nil {
				return nil, errors.New("nil: EventHub")
			}

			res := struct {
				ID []string
			}{}

			ev.Publish("/v1/bagong/sitesetup/get-site-ids", &payload, &res, nil)
			// fmt.Printf("site id: %s | sitesetup/get-site-ids : %s\n", res, e)

			iDs := []interface{}{}

			if len(res.ID) > 0 {
				for _, val := range res.ID {
					iDs = append(iDs, val)
				}
				filters = append(filters, dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.In("Value", iDs...)))
			}

			param.SetWhere(dbflex.Or(filters...))
		}
	}

	dataSidak := []shemodel.Sidak{}
	if e := h.Gets(new(shemodel.Sidak), param, &dataSidak); e != nil {
		return nil, e
	}

	return codekit.M{"data": dataSidak, "count": len(dataSidak)}, nil
}
