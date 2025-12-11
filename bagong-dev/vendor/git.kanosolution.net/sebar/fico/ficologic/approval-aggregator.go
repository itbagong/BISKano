package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type ApprovalAggregatorHandler struct {
}

type ApprovalAggregatorGroupBy string

const (
	ApprovalAggregatorGroupByModule ApprovalAggregatorGroupBy = "Module"
	ApprovalAggregatorGroupBySite   ApprovalAggregatorGroupBy = "Site"
	ApprovalAggregatorGroupByObject ApprovalAggregatorGroupBy = "Object"
)

type ApprovalAggregatorGroupByRequest struct {
	GroupBy ApprovalAggregatorGroupBy
	Status  string
	Start   *time.Time
	End     *time.Time
}

type ApprovalAggregatorGroupByResponse struct {
	SiteID string
	Group  string
	Total  float64
}

func (m *ApprovalAggregatorHandler) GroupBy(ctx *kaos.Context, payload *ApprovalAggregatorGroupByRequest) ([]*ApprovalAggregatorGroupByResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}
	user := sebar.GetUserIDFromCtx(ctx)

	ApprovalFilter := []*dbflex.Filter{
		dbflex.Eq("Status", payload.Status),
		dbflex.Eq("UserID", user),
	}
	filter := []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
	}

	if payload.Start != nil {
		ApprovalFilter = append(ApprovalFilter, dbflex.Gte("Assigned", payload.Start))
	}

	if payload.End != nil {
		ApprovalFilter = append(ApprovalFilter, dbflex.Lte("Assigned", payload.End))
	}

	if payload.GroupBy == ApprovalAggregatorGroupByObject {
		filter = append(filter, dbflex.In("SourceType", []string{"CASHBANK", "CUSTOMER", "LEDGERACCOUNT", "VENDOR"}...))
	}

	filter = append(filter, dbflex.ElemMatch("Approvals", dbflex.And(ApprovalFilter...)))

	profiles := []ficomodel.PostingApproval{}
	err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filter...),
	).SetSelect("SourceType", "Approvals", "Amount", "Dimension", "SourceID"), &profiles)
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	// get existing journal
	mapJournal, err := m.getExistingJournal(h, profiles)
	if err != nil {
		return nil, fmt.Errorf("error when get existing journal: %s", err.Error())
	}

	mapProfile := map[string]*ApprovalAggregatorGroupByResponse{}
	siteIDs := []string{}
	for _, prof := range profiles {
		// skip if journal doesn't exist
		if _, ok := mapJournal[prof.SourceID]; ok {
			group := ""
			if payload.GroupBy == ApprovalAggregatorGroupBySite {
				group = prof.Dimension.Get("Site")
				siteIDs = append(siteIDs, group)
			} else {
				group = prof.SourceType
			}

			key := group
			if v, ok := mapProfile[key]; ok {
				v.Total += prof.Amount
			} else {
				mapProfile[key] = &ApprovalAggregatorGroupByResponse{
					Group: group,
					Total: prof.Amount,
				}

				if payload.GroupBy == ApprovalAggregatorGroupBySite {
					siteIDs = append(siteIDs, group)
				}
			}
		}
	}

	// get site
	sites := []tenantcoremodel.DimensionMaster{}
	err = h.Gets(new(tenantcoremodel.DimensionMaster), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", siteIDs...),
	).SetSelect("_id", "Label"), &sites)
	if err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	mapSource := lo.Associate(sites, func(s tenantcoremodel.DimensionMaster) (string, string) {
		return s.ID, s.Label
	})

	result := make([]*ApprovalAggregatorGroupByResponse, len(mapProfile))
	i := 0
	for _, p := range mapProfile {
		// only for site
		if payload.GroupBy == ApprovalAggregatorGroupBySite {
			p.SiteID = p.Group
			p.Group = mapSource[p.SiteID]
		}

		result[i] = p
		i++
	}

	return result, nil
}

type ApprovalAggregatorGetJournal struct {
	GroupBy ApprovalAggregatorGroupBy
	Status  string
	Type    string
}

type Profile struct {
	AccountID        string
	Text             string
	TrxDate          time.Time
	Amount           float64
	SourceType       string
	SourceID         string
	VoucherNo        string
	PostingProfileID string
	JournalTypeID    string
	Status           string
}

func (m *ApprovalAggregatorHandler) GetJournal(ctx *kaos.Context, payload *ApprovalAggregatorGetJournal) ([]*Profile, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	filters := m.getFilterPostingProfile(ctx, payload)
	profiles := []*Profile{}
	err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	).SetSelect("SourceType", "SourceID", "AccountID", "Text", "TrxDate", "Amount", "PostingProfileID").
		SetSort("-TrxDate"), &profiles)
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	mapIDBySourceType := map[string][]*Profile{}
	for _, prof := range profiles {
		mapIDBySourceType[prof.SourceType] = append(mapIDBySourceType[prof.SourceType], prof)
	}

	type journal struct {
		ID            string `bson:"_id"`
		JournalTypeID string
		Status        string
	}

	result := []*Profile{}
	for key, profiles := range mapIDBySourceType {
		param := getDetailPayload{
			h:          h,
			data:       make([]*Profile, 0),
			accountIDs: make([]string, 0),
			journalIDs: make([]string, 0),
		}

		var model orm.DataModel
		switch key {
		case string(ficomodel.SubledgerCashBank):
			param.sourceModel = new(tenantcoremodel.CashBank)
			model = new(ficomodel.CashJournal)
		case string(ficomodel.SubledgerVendor):
			param.sourceModel = new(tenantcoremodel.Vendor)
			model = new(ficomodel.VendorJournal)
		case string(ficomodel.SubledgerCustomer):
			param.sourceModel = new(tenantcoremodel.Customer)
			model = new(ficomodel.CustomerJournal)
		case string(ficomodel.SubledgerAccounting):
			param.sourceModel = new(tenantcoremodel.LedgerAccount)
			model = new(ficomodel.LedgerJournal)
		// SCM
		case string(ficomodel.Inventory):
			param.sourceModel = new(scmmodel.InventJournal)
			model = new(scmmodel.InventJournal)
		case string(ficomodel.GoodReceive):
			param.sourceModel = new(scmmodel.InventReceiveIssueJournal)
			model = new(scmmodel.InventReceiveIssueJournal)
		case string(ficomodel.GoodIssuance):
			param.sourceModel = new(scmmodel.InventReceiveIssueJournal)
			model = new(scmmodel.InventReceiveIssueJournal)
		case string(ficomodel.Transfer):
			param.sourceModel = new(scmmodel.InventJournal)
			model = new(scmmodel.InventJournal)
		case string(ficomodel.ItemRequest):
			param.sourceModel = new(scmmodel.ItemRequest)
			model = new(scmmodel.ItemRequest)
		case string(ficomodel.PurchOrder):
			param.sourceModel = new(scmmodel.PurchaseOrderJournal)
			model = new(scmmodel.PurchaseOrderJournal)
		case string(ficomodel.PurchRequest):
			param.sourceModel = new(scmmodel.PurchaseRequestJournal)
			model = new(scmmodel.PurchaseRequestJournal)
		case string(ficomodel.AssetAcquisition):
			param.sourceModel = new(scmmodel.AssetAcquisitionJournal)
			model = new(scmmodel.AssetAcquisitionJournal)

		// MFG
		case string(ficomodel.WorkRequest):
			param.sourceModel = new(mfgmodel.WorkRequest)
			model = new(mfgmodel.WorkRequest)
		case string(ficomodel.WorkOrder):
			param.sourceModel = new(mfgmodel.WorkOrderPlan)
			model = new(mfgmodel.WorkOrderPlan)
		case string(ficomodel.WorkOrderReportConsumption):
			param.sourceModel = new(mfgmodel.WorkOrderPlanReportConsumption)
			model = new(mfgmodel.WorkOrderPlanReportConsumption)
		case string(ficomodel.WorkOrderReportResource):
			param.sourceModel = new(mfgmodel.WorkOrderPlanReportResource)
			model = new(mfgmodel.WorkOrderPlanReportResource)
		case string(ficomodel.WorkOrderReportOutput):
			param.sourceModel = new(mfgmodel.WorkOrderPlanReportOutput)
			model = new(mfgmodel.WorkOrderPlanReportOutput)
		default:
			continue
		}

		journalIDs := lo.Map(profiles, func(m *Profile, index int) string {
			return m.SourceID
		})

		// get existing journal
		journals := []journal{}
		err = h.Gets(model, dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", journalIDs...),
		).SetSelect("_id", "JournalTypeID", "Status"), &journals)
		if err != nil {
			return nil, fmt.Errorf("error when get existing journal: %s", err.Error())
		}

		mapJournal := map[string]journal{}
		for _, j := range journals {
			mapJournal[j.ID] = j
		}

		for i, prof := range profiles {
			// show only existing journal
			if v, ok := mapJournal[prof.SourceID]; ok {
				param.accountIDs = append(param.accountIDs, prof.AccountID)
				param.journalIDs = append(param.journalIDs, prof.SourceID)

				profiles[i].JournalTypeID = v.JournalTypeID
				profiles[i].Status = v.Status

				param.data = append(param.data, profiles[i])
			}
		}

		err = m.getDetail(&param)
		if err != nil {
			return nil, fmt.Errorf("error when mapping name object: %s", err.Error())
		}

		result = append(result, param.data...)
	}

	return result, nil
}

func (m *ApprovalAggregatorHandler) getFilterPostingProfile(ctx *kaos.Context, payload *ApprovalAggregatorGetJournal) []*dbflex.Filter {
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}
	user := sebar.GetUserIDFromCtx(ctx)

	filter := []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
		dbflex.ElemMatch("Approvals",
			dbflex.And(
				dbflex.Eq("Status", payload.Status),
				dbflex.Eq("UserID", user),
			),
		),
	}

	if payload.GroupBy == ApprovalAggregatorGroupBySite {
		filter = append(filter, dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.Type)))
	} else {
		filter = append(filter, dbflex.Eq("SourceType", payload.Type))
	}

	return filter
}

type getDetailPayload struct {
	h           *datahub.Hub
	sourceModel orm.DataModel
	accountIDs  []string
	journalIDs  []string
	data        []*Profile
}

func (m *ApprovalAggregatorHandler) getDetail(p *getDetailPayload) error {
	// get name
	sources := []codekit.M{}
	err := p.h.Gets(p.sourceModel, dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", p.accountIDs...),
	).SetSelect("_id", "Name"), &sources)
	if err != nil {
		return err
	}

	mapSource := lo.Associate(sources, func(source codekit.M) (string, codekit.M) {
		return source["_id"].(string), source
	})

	schedules := []ficomodel.CashSchedule{}
	err = p.h.Gets(new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(
		dbflex.In("SourceJournalID", p.journalIDs...),
	), &schedules)
	if err != nil {
		return fmt.Errorf("error when get cash schedule: %s", err.Error())
	}

	mapVoucherNo := lo.Associate(schedules, func(source ficomodel.CashSchedule) (string, string) {
		return source.SourceJournalID, source.VoucherNo
	})

	for i := range p.data {
		if v, ok := mapSource[p.data[i].AccountID]; ok {
			p.data[i].AccountID = v["Name"].(string)
		}

		if v, ok := mapVoucherNo[p.data[i].SourceID]; ok {
			p.data[i].VoucherNo = v
		}
	}

	return nil
}

type ApprovalAggregatorPostByGroupRequest struct {
	GroupBy ApprovalAggregatorGroupBy
	Type    string
	Op      PostOp
	Text    string
}

func (m *ApprovalAggregatorHandler) PostByGroup(ctx *kaos.Context, payloads []ApprovalAggregatorPostByGroupRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	postParams := []PostRequest{}
	for _, payload := range payloads {
		postParam := PostRequest{
			Op:   payload.Op,
			Text: payload.Text,
		}

		param := ApprovalAggregatorGetJournal{
			GroupBy: payload.GroupBy,
			Status:  "PENDING",
			Type:    payload.Type,
		}

		filters := m.getFilterPostingProfile(ctx, &param)
		profiles := codekit.Ms{}
		err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
			dbflex.And(filters...),
		).SetSelect("SourceType", "SourceID"), &profiles)
		if err != nil {
			return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
		}

		for _, profile := range profiles {
			switch profile["SourceType"].(string) {
			case string(ficomodel.SubledgerCashBank):
				postParam.JournalType = ficomodel.SubledgerCashBank
			case string(ficomodel.SubledgerVendor):
				postParam.JournalType = ficomodel.SubledgerVendor
			case string(ficomodel.SubledgerCustomer):
				postParam.JournalType = ficomodel.SubledgerCustomer
			case string(ficomodel.SubledgerAccounting):
				postParam.JournalType = ficomodel.SubledgerAccounting
			default:
				continue
			}

			postParam.JournalID = profile["SourceID"].(string)
			postParams = append(postParams, postParam)
		}
	}

	_, err := new(PostingProfileHandler).Post(ctx, postParams)
	if err != nil {
		return nil, fmt.Errorf("error when posting journals : %s", err.Error())
	}

	return "Success", nil
}

func (m *ApprovalAggregatorHandler) Post(ctx *kaos.Context, payloads []PostRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	type Profile struct {
		SourceType string
		SourceID   string
	}

	ids := lo.Map(payloads, func(m PostRequest, index int) string {
		return m.JournalID
	})

	profiles := []Profile{}
	err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
		dbflex.In("SourceID", ids...),
	).SetSelect("SourceType", "SourceID"), &profiles)
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	mapProfile := lo.Associate(profiles, func(p Profile) (string, string) {
		return p.SourceID, p.SourceType
	})

	postParams := make([]PostRequest, len(payloads))
	for i, payload := range payloads {
		switch mapProfile[payload.JournalID] {
		case string(ficomodel.SubledgerCashBank):
			payload.JournalType = ficomodel.SubledgerCashBank
		case string(ficomodel.SubledgerVendor):
			payload.JournalType = ficomodel.SubledgerVendor
		case string(ficomodel.SubledgerCustomer):
			payload.JournalType = ficomodel.SubledgerCustomer
		case string(ficomodel.SubledgerAccounting):
			payload.JournalType = ficomodel.SubledgerAccounting
		default:
			continue
		}

		postParams[i] = payload
	}

	_, err = new(PostingProfileHandler).Post(ctx, postParams)
	if err != nil {
		return nil, fmt.Errorf("error when posting journals : %s", err.Error())
	}

	return "Success", nil
}

// getExistingJournal return existing journal
func (m *ApprovalAggregatorHandler) getExistingJournal(hub *datahub.Hub, profiles []ficomodel.PostingApproval) (map[string]bool, error) {
	var err error
	mapIDBySourceType := map[string][]interface{}{}
	for _, prof := range profiles {
		mapIDBySourceType[prof.SourceType] = append(mapIDBySourceType[prof.SourceType], prof.SourceID)
	}

	type journal struct {
		ID string `bson:"_id"`
	}

	result := map[string]bool{}
	for key, ids := range mapIDBySourceType {
		var model orm.DataModel
		switch key {
		case string(ficomodel.SubledgerCashBank):
			model = new(ficomodel.CashJournal)
		case string(ficomodel.SubledgerVendor):
			model = new(ficomodel.VendorJournal)
		case string(ficomodel.SubledgerCustomer):
			model = new(ficomodel.CustomerJournal)
		case string(ficomodel.SubledgerAccounting):
			model = new(ficomodel.LedgerJournal)

		// SCM
		case string(ficomodel.Inventory):
			model = new(scmmodel.InventJournal)
		case string(ficomodel.GoodReceive):
			model = new(scmmodel.InventReceiveIssueJournal)
		case string(ficomodel.GoodIssuance):
			model = new(scmmodel.InventReceiveIssueJournal)
		case string(ficomodel.Transfer):
			model = new(scmmodel.InventJournal)
		case string(ficomodel.ItemRequest):
			model = new(scmmodel.ItemRequest)
		case string(ficomodel.PurchOrder):
			model = new(scmmodel.PurchaseOrderJournal)
		case string(ficomodel.PurchRequest):
			model = new(scmmodel.PurchaseRequestJournal)
		case string(ficomodel.AssetAcquisition):
			model = new(scmmodel.AssetAcquisitionJournal)

		// MFG
		case string(ficomodel.WorkRequest):
			model = new(mfgmodel.WorkRequest)
		case string(ficomodel.WorkOrder):
			model = new(mfgmodel.WorkOrderPlan)
		case string(ficomodel.WorkOrderReportConsumption):
			model = new(mfgmodel.WorkOrderPlanReportConsumption)
		case string(ficomodel.WorkOrderReportResource):
			model = new(mfgmodel.WorkOrderPlanReportResource)
		case string(ficomodel.WorkOrderReportOutput):
			model = new(mfgmodel.WorkOrderPlanReportOutput)
		default:
			continue
		}

		journals := []journal{}
		err = hub.Gets(model, dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", ids...),
		).SetSelect("_id"), &journals)
		if err != nil {
			return nil, fmt.Errorf("error when get journal: %s", err.Error())
		}

		for _, j := range journals {
			result[j.ID] = true
		}
	}

	return result, nil
}
