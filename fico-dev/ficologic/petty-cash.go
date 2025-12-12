package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PettyCashHandler struct {
}

type GetsWhereRequest struct {
	CompanyID string
	Text      string
	StartDate *time.Time
	EndDate   *time.Time
	Items     []*dbflex.Filter
	Op        dbflex.FilterOp
}

type GetsRequest struct {
	Skip  int
	Take  int
	Where GetsWhereRequest
}

func (m *PettyCashHandler) Gets(ctx *kaos.Context, payload GetsRequest) (interface{}, error) {
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
				"Value": "Petty Cash",
			},
		},
	}

	isGetMatchItems := true

	if len(dimIface) > 0 {
		match["$and"] = GetMatchDimension(dimIface)
	}

	filterDate := bson.M{}
	if payload.Where.StartDate != nil {
		filterDate["$gte"] = payload.Where.StartDate
	}

	if payload.Where.EndDate != nil {
		filterDate["$lte"] = payload.Where.EndDate
	}

	if len(filterDate) > 0 {
		match["TrxDate"] = filterDate
		isGetMatchItems = false
	}

	if payload.Where.CompanyID != "" {
		match["CompanyID"] = payload.Where.CompanyID
	} else {
		match["CompanyID"] = coID
	}

	if payload.Where.Text != "" {
		match["Text"] = bson.M{
			"$regex": primitive.Regex{Pattern: payload.Where.Text, Options: "i"},
		}
		isGetMatchItems = false
	}

	if isGetMatchItems {
		match = GetMatchItems(payload.Where.Items, match)
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
	cmd := dbflex.From(new(ficomodel.CashJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &counts); err != nil {
		return nil, fmt.Errorf("err when get petty cash journal count: %s", err.Error())
	}

	count := 0
	if len(counts) > 0 {
		count = counts[0].Count
	}

	pipe = []bson.M{
		{
			"$match": match,
		},
		{
			"$sort": bson.M{"TrxDate": -1},
		},
		{
			"$skip": payload.Skip,
		},
		{
			"$limit": payload.Take,
		},
	}

	// get cash journal petty cash
	journals := []ficomodel.CashJournal{}
	cmd = dbflex.From(new(ficomodel.CashJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &journals); err != nil {
		return nil, fmt.Errorf("err when get petty cash journal: %s", err.Error())
	}

	result := codekit.M{}.Set("data", journals).Set("count", count)
	return result, nil
}

type GetSubmissionHistoryRequest struct {
	JournalTypeID string
	SiteID        string
	CashBankID    string
	TrxDate       time.Time
}

func (m *PettyCashHandler) GetSubmissionHistory(ctx *kaos.Context, p *GetSubmissionHistoryRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)

	// get cash journal petty cash
	cashJournals := []ficomodel.CashJournal{}
	err := h.Gets(new(ficomodel.CashJournal), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("CompanyID", coID),
			dbflex.Eq("JournalTypeID", p.JournalTypeID),
			dbflex.Eq("CashBookID", p.CashBankID),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted),
			dbflex.Lt("TrxDate", p.TrxDate),
			dbflex.ElemMatch("References", dbflex.Eq("Key", "Submission Type"), dbflex.Eq("Value", "Petty Cash")),
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", p.SiteID)),
		),
	).SetSort("-TrxDate").SetTake(1), &cashJournals)
	if err != nil {
		return nil, fmt.Errorf("err when get cash journal: %s", err.Error())
	}

	cashTransactions := []ficomodel.CashTransaction{}
	filters := []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
		dbflex.Eq("SourceJournalType", p.JournalTypeID),
		dbflex.Eq("CashBank._id", p.CashBankID),
		dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", p.SiteID)),
	}
	switch len(cashJournals) {
	case 0:
		filters = append(filters, dbflex.Lte("TrxDate", p.TrxDate))
	case 1:
		filters = append(filters, dbflex.Gte("TrxDate", cashJournals[0].TrxDate), dbflex.Lte("TrxDate", p.TrxDate))
	}

	// get cash transaction
	err = h.Gets(new(ficomodel.CashTransaction), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	), &cashTransactions)
	if err != nil {
		return nil, fmt.Errorf("err when get cash transaction: %s", err.Error())
	}

	return cashTransactions, nil
}
