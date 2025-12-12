package ficologic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type EmployeeExpenseHandler struct {
}

type GetsEmployeeExpenseRequest struct {
	Sort  []string
	Skip  int
	Take  int
	Where map[string]interface{}
}

func (m *EmployeeExpenseHandler) Gets(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)

	jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})

	match := bson.M{
		"References": bson.M{
			"$elemMatch": bson.M{
				"Key":   "Submission Type",
				"Value": "Employee Expense",
			},
		},
		"CompanyID": coID,
	}

	if payload != nil {
		if payload.Where != nil {
			match = GetMatchItems(payload.Where.Items, match)
		}
	}

	if len(dimIface) > 0 {
		match["$and"] = GetMatchDimension(dimIface)
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Count": bson.M{"$sum": 1},
			},
		},
	}

	type CountJournal struct {
		Count int
	}

	// get count
	counts := []CountJournal{}
	cmd := dbflex.From(new(ficomodel.VendorJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &counts); err != nil {
		return nil, fmt.Errorf("err when get employee expense journal count: %s", err.Error())
	}

	count := 0
	if len(counts) > 0 {
		count = counts[0].Count
	}

	sort := bson.M{}
	if len(payload.Sort) > 0 {
		field := strings.Split(payload.Sort[0], "-")
		if len(field) > 1 {
			sort[field[1]] = -1
		} else {
			sort[field[0]] = 1
		}
	}

	pipe = []bson.M{
		{
			"$match": match,
		},
		{
			"$sort": sort,
		},
		{
			"$skip": payload.Skip,
		},
		{
			"$limit": payload.Take,
		},
	}

	// get employee expense journal
	journals := []ficomodel.VendorJournal{}
	cmd = dbflex.From(new(ficomodel.VendorJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &journals); err != nil {
		return nil, fmt.Errorf("err when employee expense journal: %s", err.Error())
	}

	result := codekit.M{}.Set("data", journals).Set("count", count)
	return result, nil
}

type GetVendorRequest struct {
	Site string
}

func (m *EmployeeExpenseHandler) GetVendor(ctx *kaos.Context, payload GetVendorRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	vendor := new(tenantcoremodel.Vendor)
	err := h.GetByParm(vendor, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.Site)),
			dbflex.Eq("GroupID", "vendorvirtual"),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get vendor: %s", err.Error())
	}

	return vendor, nil
}
