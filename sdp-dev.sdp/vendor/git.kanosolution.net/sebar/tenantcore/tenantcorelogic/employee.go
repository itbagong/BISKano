package tenantcorelogic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type EmployeeEngine struct {
}

type PostRequestEmpWarehouse struct {
	tenantcoremodel.Employee
	Warehouse tenantcoremodel.LocationWarehouse
}
type FindSitePayload struct {
	Name string
}

func (obj *EmployeeEngine) GetEmpWarehouse(ctx *kaos.Context, payload []interface{}) (*PostRequestEmpWarehouse, error) {
	userID := sebar.GetUserIDFromCtx(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if len(payload) > 0 {
		userID = fmt.Sprintf("%v", payload[0])
	}

	emp := new(tenantcoremodel.Employee)
	// if e := h.GetByID(emp, userID); e != nil {
	// 	return nil, errors.New("Employee not found: " + e.Error())
	// }
	h.GetByFilter(emp, dbflex.Eq("UserID", userID))

	empW := new(tenantcoremodel.LocationWarehouse)
	wheres := []*dbflex.Filter{
		dbflex.Ne("_id", ""),
	}
	if len(emp.Dimension) > 0 {
		wheres = append(wheres, emp.Dimension.Where())
	}
	h.GetByFilter(empW, dbflex.And(wheres...))

	res := PostRequestEmpWarehouse{}
	res.ID = emp.ID
	res.Name = emp.Name
	res.Email = emp.Email
	res.EmployeeGroupID = emp.EmployeeGroupID
	res.EmploymentType = emp.EmploymentType
	res.CompanyID = emp.CompanyID
	res.Sites = emp.Sites
	res.JoinDate = emp.JoinDate
	res.Dimension = emp.Dimension
	res.IsActive = emp.IsActive
	res.IsLogin = emp.IsLogin
	res.Created = emp.Created
	res.LastUpdate = emp.LastUpdate
	res.Warehouse = *empW

	return &res, nil
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

func (obj *EmployeeEngine) GetByCurrentUser(ctx *kaos.Context, payload *dbflex.QueryParam) (*tenantcoremodel.Employee, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	emp := new(tenantcoremodel.Employee)
	h.GetByFilter(emp, dbflex.Eq("UserID", userID))
	if emp.ID == "" {
		h.GetByID(emp, userID)
	}

	return emp, nil
}
