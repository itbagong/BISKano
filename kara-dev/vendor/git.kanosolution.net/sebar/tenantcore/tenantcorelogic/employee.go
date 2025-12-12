package tenantcorelogic

import (
	"errors"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type EmployeeEngine struct {
}

type FindSitePayload struct {
	Name string
}

func (o *EmployeeEngine) FindBySite(ctx *kaos.Context, payload *dbflex.QueryParam) ([]tenantcoremodel.Employee, error) {
	employees := []tenantcoremodel.Employee{}
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return employees, errors.New("missing: connection")
	}

	r := ctx.Data().Get("http_request", nil).(*http.Request)
	siteID := strings.TrimSpace(r.URL.Query().Get("SiteID"))
	name := strings.TrimSpace(r.URL.Query().Get("Name"))

	filters := []*dbflex.Filter{}

	if siteID != "" {
		filters = append(filters, dbflex.In("Sites", siteID))
	}
	if name != "" {
		filters = append(filters, dbflex.Contains("Name", name))
	}

	if payload != nil {
		if payload.Where != nil {
			fieldVal := payload.Where.Field
			if fieldVal == "_id" {
				aInterface := payload.Where.Value.(interface{})
				vID := aInterface.(string)
				if vID != "" {
					filters = append(filters, dbflex.Eq("_id", vID))
				}
			} else if fieldVal == "Name" {
				aInterface := payload.Where.Value.([]interface{})
				aString := make([]string, len(aInterface))
				for i, v := range aInterface {
					aString[i] = v.(string)
				}
				if len(aString) > 0 {
					filters = append(filters, dbflex.Contains("Name", aString[0]))
				}
			}
		}
	}

	if len(filters) > 0 {
		e := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.And(filters...)).SetSort("Name"), &employees)
		if e != nil {
			return employees, errors.New("Failed populate data driver: " + e.Error())
		}
	} else {
		e := h.Gets(new(tenantcoremodel.Employee), nil, &employees)
		if e != nil {
			return employees, errors.New("Failed populate data driver: " + e.Error())
		}
	}

	return employees, nil
}
