package bagonglogic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeHandler struct {
}

type PostRequestEmployee struct {
	tenantcoremodel.Employee
	Detail bagongmodel.EmployeeDetail
}

type CreateUserRequest struct {
	ID       string `bson:"_id" json:"_id"`
	Email    string
	Name     string
	Password string
}

func (obj *EmployeeHandler) Get(ctx *kaos.Context, payload []interface{}) (*PostRequestEmployee, error) {
	if len(payload) == 0 {
		return nil, errors.New("invalid request")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	emp := new(tenantcoremodel.Employee)
	if e := h.GetByID(emp, payload[0]); e != nil {
		return nil, fmt.Errorf("employee not found: %s", payload)
	}

	empDetail := new(bagongmodel.EmployeeDetail)
	h.GetByFilter(empDetail, dbflex.Eq("EmployeeID", emp.ID))

	res := PostRequestEmployee{}
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
	res.UserID = emp.UserID
	res.Created = emp.Created
	res.LastUpdate = emp.LastUpdate
	res.Detail = *empDetail

	return &res, nil
}

func (obj *EmployeeHandler) Save(ctx *kaos.Context, payload *PostRequestEmployee) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", errors.New("nil: EventHub")
	}

	if payload.Employee.ID == "" {
		tenantcorelogic.MWPreAssignSequenceNo("Employee", false, "_id")(ctx, &payload.Employee)

		if e := h.GetByID(new(tenantcoremodel.Employee), payload.Employee.ID); e == nil {
			ctx.Log().Errorf("error duplicate key: %s", e.Error())
			return nil, errors.New("error duplicate key: " + e.Error())
		}

		userReq := rbaclogic.GetUserByRequest{
			FindBy: "email",
			FindID: payload.Email,
		}
		existsCandidate := new(rbacmodel.User)
		err := ev.Publish("/v1/iam/user/get-by", &userReq, existsCandidate, nil)
		if err != nil && err.Error() != io.EOF.Error() {
			return "", fmt.Errorf("error when check user: %s", err.Error())
		}

		if existsCandidate.ID != "" {
			return "", fmt.Errorf("duplicate: User: %s", payload.Email)
		}

		if e := h.Insert(&payload.Employee); e != nil {
			return nil, errors.New("error insert Employee: " + e.Error())
		}
	} else {
		existingUser := new(tenantcoremodel.Employee)
		e := h.GetByID(existingUser, payload.ID)
		if e != nil {
			return nil, errors.New("error when get existing user: " + e.Error())
		}

		// check if email is changed
		if existingUser.Email != payload.Email {
			userReq := rbaclogic.GetUserByRequest{
				FindBy: "email",
				FindID: payload.Email,
			}
			existsCandidate := new(rbacmodel.User)
			err := ev.Publish("/v1/iam/user/get-by", &userReq, existsCandidate, nil)
			if err != nil && err.Error() != io.EOF.Error() {
				return "", fmt.Errorf("error when check user: %s", err.Error())
			}

			if existsCandidate.ID != "" {
				return "", fmt.Errorf("duplicate: User: %s", payload.Email)
			}
		}

		if e := h.Save(&payload.Employee); e != nil {
			return nil, errors.New("error update Employee: " + e.Error())
		}
	}

	payload.Detail.EmployeeID = payload.ID
	if e := h.GetByID(new(bagongmodel.EmployeeDetail), payload.Detail.ID); e != nil {
		payload.Detail.EmployeeID = payload.Employee.ID
		if e := h.Insert(&payload.Detail); e != nil {
			return nil, errors.New("error insert EmployeeDetail: " + e.Error())
		}
	} else {
		if e := h.Save(&payload.Detail); e != nil {
			return nil, errors.New("error update EmployeeDetail: " + e.Error())
		}
	}

	// update dimension to rbac user when islogin is true
	if payload.Employee.IsLogin {
		req := struct {
			EmployeeID        string
			EmployeeDimension tenantcoremodel.Dimension
			Name              string
			Email             string
		}{
			EmployeeID:        payload.Employee.ID,
			EmployeeDimension: payload.Employee.Dimension,
			Name:              payload.Name,
			Email:             payload.Email,
		}

		res := ""
		err := ev.Publish("/v1/iam/user/update-user", &req, &res, nil)
		if err != nil {
			fmt.Printf("employee name: %s | result : %s | user/update-user error: %s\n", payload.Employee.Name, res, err)
		}
	}

	return payload, nil
}

func (obj *EmployeeHandler) CreateUser(ctx *kaos.Context, payload *CreateUserRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("payload is nil")
	} else {
		if payload.ID == "" || payload.Email == "" || payload.Name == "" || payload.Password == "" {
			return nil, errors.New("ID, Email, Name, and Password is required")
		}
	}

	req := rbaclogic.CreateUserRequest{
		User: &rbacmodel.User{
			ID:          payload.ID,
			Email:       payload.Email,
			LoginID:     payload.Email,
			DisplayName: payload.Name,
			Status:      "Active",
			Enable:      true,
		},
		Password: payload.Password,
	}
	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", errors.New("nil: EventHub")
	}
	userid := ""
	err := ev.Publish("/v1/iam/user/create", &req, &userid, nil)
	if err != nil {
		return "", err
	}

	// update islogin true
	employees := []tenantcoremodel.Employee{}
	e := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", payload.ID)), &employees)
	if e != nil {
		return employees, errors.New("Failed populate data driver: " + e.Error())
	}

	if len(employees) > 0 {
		employee := employees[0]
		employee.IsLogin = true
		employee.UserID = userid

		if e := h.Save(&employee); e != nil {
			return nil, errors.New("error update Employee: " + e.Error())
		}
	}

	return payload, nil
}

type ResponseJournal struct {
	JournalType ficomodel.LedgerJournalType
	UserID      string
	Dimension   string
	EmployeeID  string
}

func (obj *EmployeeHandler) GetJournal(ctx *kaos.Context, payload []interface{}) (ficomodel.LedgerJournalType, error) {

	userID := sebar.GetUserIDFromCtx(ctx)
	result := ficomodel.LedgerJournalType{}
	// hardcode ini mas
	userID = "admin_admin"

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return result, errors.New("missing: connection")
	}

	emp := new(tenantcoremodel.Employee)
	if e := h.GetByID(emp, userID); e != nil {
		return result, fmt.Errorf("employee not found: %s", payload)
	}

	ledgerJT := new(ficomodel.LedgerJournalType)
	h.GetByFilter(ledgerJT, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", emp.Dimension.Get("Site"))))

	result = *ledgerJT

	return result, nil
}

func (obj *EmployeeHandler) GetEmployeeByPosition(ctx *kaos.Context, payload *dbflex.QueryParam) ([]tenantcoremodel.Employee, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	r := ctx.Data().Get("http_request", nil).(*http.Request)
	SiteID := strings.TrimSpace(r.URL.Query().Get("SiteID"))
	Position := strings.TrimSpace(r.URL.Query().Get("Position"))
	Op := strings.TrimSpace(r.URL.Query().Get("Op"))
	if Op == "" {
		Op = "Contain"
	}

	var detailIds []interface{}
	if Position != "" {
		md := []tenantcoremodel.MasterData{}
		if Op == "Contain" {
			err := hub.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
				dbflex.And(dbflex.Eq("MasterDataTypeID", "PTE"), dbflex.Contains("Name", Position)),
			), &md)
			if err != nil {
				return nil, err
			}
		} else {
			err := hub.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
				dbflex.And(dbflex.Eq("MasterDataTypeID", "PTE"), dbflex.Not(dbflex.Contains("Name", Position))),
			), &md)
			if err != nil {
				return nil, err
			}
		}

		positionIds := lo.Map(md, func(m tenantcoremodel.MasterData, index int) interface{} {
			return m.ID
		})

		ed := []bagongmodel.EmployeeDetail{}
		err := hub.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
			dbflex.And(dbflex.In("Position", positionIds...)),
		), &ed)
		if err != nil {
			return nil, err
		}

		detailIds = lo.Map(ed, func(m bagongmodel.EmployeeDetail, index int) interface{} {
			return m.EmployeeID
		})
	}

	query := dbflex.NewQueryParam()
	if payload.Skip != 0 {
		query = query.SetSkip(payload.Skip)
	}

	if payload.Take != 0 {
		query = query.SetTake(payload.Take)
	}

	if len(payload.Sort) > 0 {
		query = query.SetSort(payload.Sort...)
	}

	filters := []*dbflex.Filter{dbflex.Eq("IsActive", true)}
	if len(detailIds) > 0 {
		filters = append(filters, dbflex.In("_id", detailIds...))
	}
	if SiteID != "" {
		filters = append(filters, dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", SiteID)))
	}

	if payload != nil {
		if payload.Where != nil {
			filters2 := []*dbflex.Filter{}
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
				if len(vItems) > 0 {
					for _, val := range vItems {
						fieldVal := val.Field
						opVal := val.Op
						if opVal == dbflex.OpContains {
							aInterface := val.Value.([]interface{})
							aString := make([]string, len(aInterface))
							for i, v := range aInterface {
								aString[i] = v.(string)
							}
							if len(aString) > 0 {
								if aString[0] != "" {
									filters2 = append(filters2, dbflex.Contains(fieldVal, aString[0]))
								}
							}
						}
					}
				}
			} else {
				fieldVal := payload.Where.Field
				opVal := payload.Where.Op
				if opVal == dbflex.OpEq {
					aInterface := payload.Where.Value.(string)
					filters2 = append(filters2, dbflex.Eq(fieldVal, aInterface))
				}
			}
			if len(filters2) > 0 {
				filters = append(filters, dbflex.Or(filters2...))
			}
		}
	}

	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	employees := []tenantcoremodel.Employee{}
	err := hub.Gets(new(tenantcoremodel.Employee), query, &employees)
	if err != nil {
		return nil, err
	}

	return employees, nil
}

type GetsFilterResponse struct {
	tenantcoremodel.Employee
	EmployeeDetail bagongmodel.EmployeeDetail
}

func (obj *EmployeeHandler) GetsFilter(ctx *kaos.Context, _ *interface{}) ([]tenantcoremodel.Employee, error) {
	p := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// TODO: change this filter in query param to in body param using Where clauses

	filters := []*dbflex.Filter{}

	// filtering by bagongmodel EmployeeDetail fields
	bgFilters := []*dbflex.Filter{}

	if p["Department"] != "" {
		bgFilters = append(bgFilters, dbflex.Eq("Department", p["Department"]))
	}

	if p["Rank"] != "" {
		bgFilters = append(bgFilters, dbflex.Eq("Rank", p["Rank"]))
	}

	if p["Level"] != "" {
		bgFilters = append(bgFilters, dbflex.Eq("Level", p["Level"]))
	}

	if p["Group"] != "" {
		bgFilters = append(bgFilters, dbflex.Eq("Group", p["Group"]))
	}

	if p["SubGroup"] != "" {
		bgFilters = append(bgFilters, dbflex.Eq("SubGroup", p["SubGroup"]))
	}

	if len(bgFilters) > 0 {
		eds := []bagongmodel.EmployeeDetail{}
		if e := h.GetsByFilter(new(bagongmodel.EmployeeDetail), dbflex.And(bgFilters...), &eds); e != nil {
			return nil, e
		}

		empFilterIDs := lo.Map(eds, func(d bagongmodel.EmployeeDetail, i int) string {
			return d.EmployeeID
		})

		filters = append(filters, dbflex.In("_id", empFilterIDs...))
	}

	// filtering by tenantcoremodel Employee fields
	if p["_id"] != "" {
		filters = append(filters, dbflex.Eq("_id", p["_id"]))
	}

	if p["ids"] != "" {
		_ids := strings.Split(p["ids"], ",")
		if len(_ids) > 0 {
			filters = append(filters, dbflex.In("_id", _ids...))
		}
	}

	if p["Name"] != "" {
		filters = append(filters, dbflex.Contains("Name", p["Name"]))
	}

	if p["Email"] != "" {
		filters = append(filters, dbflex.Contains("Email", p["Email"]))
	}

	if p["EmployeeGroupID"] != "" {
		filters = append(filters, dbflex.Eq("EmployeeGroupID", p["EmployeeGroupID"]))
	}

	if p["EmploymentType"] != "" {
		filters = append(filters, dbflex.Eq("EmploymentType", p["EmploymentType"]))
	}

	if p["CompanyID"] != "" {
		filters = append(filters, dbflex.Eq("CompanyID", p["CompanyID"]))
	}

	filter := new(dbflex.Filter)
	if len(filters) > 0 {
		filter = dbflex.And(filters...)
	}

	res := []tenantcoremodel.Employee{}
	if e := h.GetsByFilter(new(tenantcoremodel.Employee), filter, &res); e != nil {
		return nil, e
	}

	return res, nil
}

type GetEmployeeRequest struct {
	ID string
}

func (obj *EmployeeHandler) GetEmployeeByID(ctx *kaos.Context, payload *GetEmployeeRequest) (*tenantcoremodel.EmployeeTenantBagong, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if len(payload.ID) == 0 {
		return nil, fmt.Errorf("no ID provided")
	}

	emps := []tenantcoremodel.EmployeeTenantBagong{}
	err := h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(dbflex.Eq("EmployeeID", payload.ID)), &emps)
	if err != nil {
		return nil, err
	}

	emp := tenantcoremodel.EmployeeTenantBagong{}
	if len(emps) > 0 {
		emp = emps[0]

		empHeader := new(tenantcoremodel.Employee)
		if e := h.GetByID(empHeader, emp.EmployeeID); e != nil {
			return nil, fmt.Errorf("employee not found: %s", emp.EmployeeID)
		}

		emp.ID = empHeader.ID
		emp.Name = empHeader.Name
		emp.Email = empHeader.Email
		emp.EmployeeGroupID = empHeader.EmployeeGroupID
		emp.EmploymentType = empHeader.EmploymentType
		emp.CompanyID = empHeader.CompanyID
		emp.Sites = empHeader.Sites
		emp.Dimension = empHeader.Dimension
		emp.IsActive = empHeader.IsActive
		emp.IsLogin = empHeader.IsLogin
		emp.UserID = empHeader.UserID
		emp.Created = empHeader.Created
		emp.LastUpdate = empHeader.LastUpdate
	}

	return &emp, nil
}

type SaveEmployeeResumeRequest struct {
	Name   string
	Detail *bagongmodel.EmployeeDetail
}

func (obj *EmployeeHandler) SaveEmployeeResume(ctx *kaos.Context, payload *SaveEmployeeResumeRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	payload.Detail.EmployeeID = payload.Detail.ID

	err := h.Save(payload.Detail)
	if err != nil {
		return nil, fmt.Errorf("error when save resume: %s", err.Error())
	}

	employee := new(tenantcoremodel.Employee)
	err = h.GetByID(employee, payload.Detail.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	employee.Name = payload.Name
	err = h.Save(employee)
	if err != nil {
		return nil, fmt.Errorf("error when save employee: %s", err.Error())
	}

	return payload, nil
}

func (obj *EmployeeHandler) GetEmployeeResume(ctx *kaos.Context, param *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	employeeDetail := new(bagongmodel.EmployeeDetail)
	h.GetByID(employeeDetail, sebar.GetUserIDFromCtx(ctx))

	employee := new(tenantcoremodel.Employee)
	h.GetByID(employee, employeeDetail.ID)

	result := struct {
		*bagongmodel.EmployeeDetail
		Name string
	}{
		Name:           employee.Name,
		EmployeeDetail: employeeDetail,
	}
	return result, nil
}

func (obj *EmployeeHandler) GetEmployees(ctx *kaos.Context, param *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	filterEmployeeNo := new(dbflex.Filter)
	if param.Where != nil {
		for _, item := range param.Where.Items {
			if item.Op == dbflex.OpOr {
				for _, orItem := range item.Items {
					if orItem.Field == "EmployeeNo" {
						filterEmployeeNo = orItem
						break
					}
				}
			}
		}
	}

	// filter employee no to employee detail table
	if filterEmployeeNo.Field != "" {
		employees := []bagongmodel.EmployeeDetail{}
		e := h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(filterEmployeeNo), &employees)
		if e != nil {
			return nil, errors.New("error when get employee: " + e.Error())
		}

		ids := lo.Map(employees, func(emp bagongmodel.EmployeeDetail, index int) string {
			return emp.EmployeeID
		})

		if len(ids) > 0 {
			// set new filter based on ids from employee detail table
			// param.Where.Items[0].Items = append(param.Where.Items[0].Items, dbflex.In("_id", ids...))
			param.Where.Items[0].Items = []*dbflex.Filter{
				dbflex.In("_id", ids...),
			}
		}
	}

	employees := []codekit.M{}
	err := h.Gets(new(tenantcoremodel.Employee), param, &employees)
	if err != nil {
		return nil, errors.New("error when get employee: " + err.Error())
	}

	employeeGroupIDs := make([]string, len(employees))
	ids := make([]string, len(employees))
	siteIDs := make([]interface{}, 0)
	for i, m := range employees {
		employeeGroupIDs[i] = m.GetString("EmployeeGroupID")
		ids[i] = m.GetString("_id")
		if m["Sites"] != nil {
			sites := m["Sites"].(primitive.A)
			if len(sites) > 0 {
				siteIDs = append(siteIDs, sites...)
			}
		}
	}

	sites := []bagongmodel.Site{}
	err = h.GetsByFilter(new(bagongmodel.Site), dbflex.In("_id", siteIDs...), &sites)
	if err != nil {
		return nil, errors.New("error when get site: " + err.Error())
	}

	mapSites := lo.Associate(sites, func(site bagongmodel.Site) (string, string) {
		return site.ID, site.Name
	})

	employeeGroups := []tenantcoremodel.EmployeeGroup{}
	err = h.GetsByFilter(new(tenantcoremodel.EmployeeGroup), dbflex.In("_id", employeeGroupIDs...), &employeeGroups)
	if err != nil {
		return nil, errors.New("error when get employee group: " + err.Error())
	}

	mapEmployeeGroups := lo.Associate(employeeGroups, func(group tenantcoremodel.EmployeeGroup) (string, string) {
		return group.ID, group.Name
	})

	employeeDetails := []bagongmodel.EmployeeDetail{}
	err = h.GetsByFilter(new(bagongmodel.EmployeeDetail), dbflex.In("EmployeeID", ids...), &employeeDetails)
	if err != nil {
		return nil, errors.New("error when get employee detail: " + err.Error())
	}

	mapEmployeeDetail := lo.Associate(employeeDetails, func(emp bagongmodel.EmployeeDetail) (string, string) {
		return emp.EmployeeID, emp.EmployeeNo
	})

	for _, m := range employees {
		if v, ok := mapEmployeeGroups[m.GetString("EmployeeGroupID")]; ok {
			m.Set("EmployeeGroupID", v)
		}

		if v, ok := mapEmployeeDetail[m.GetString("_id")]; ok {
			m.Set("EmployeeNo", v)
		}

		if m["Sites"] != nil {
			sites := m["Sites"].(primitive.A)
			if len(sites) > 0 {
				if v, ok := mapSites[sites[0].(string)]; ok {
					m.Set("Sites", []interface{}{v})
				}
			}
		}
	}

	var count int
	count, err = h.Count(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(param.Where))
	if err != nil {
		return nil, errors.New("error when get employee count: " + err.Error())
	}

	return codekit.M{"count": count, "data": employees}, nil
}
