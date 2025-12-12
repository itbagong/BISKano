package mfglogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type PostingProfileHandler struct {
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payloads []ficologic.PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))
	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	coID, userID, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var (
		ph ficologic.PostRequestHandler
	)

	for index, payload := range payloads {
		reqOpt := ficologic.PostingHubCreateOpt{
			Db:        h,
			UserID:    userID,
			CompanyID: coID,
			ModuleID:  string(payload.JournalType),
			JournalID: payload.JournalID,
			Op:        payload.Op,
		}

		switch payload.JournalType {
		case tenantcoremodel.TrxModule(mfgmodel.JournalWorkRequest):
			ph = NewWorkRequestPosting(ctx, reqOpt) // Work Request
		// case scmmodel.ModuleWorkorder:
		// 	ph = NewJournalPostingWorkOrder(reqOpt) // Work Order - deprecated
		case tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderPlan):
			ph = NewWorkOrderPlanPosting(ctx, reqOpt) // Work Order Plan
		case tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportConsumption):
			ph = NewWorkOrderPlanReportConsumptionPosting(ctx, reqOpt) // Work Order Plan Report Consumption
		case tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportResource):
			ph = NewWorkOrderPlanReportResourcePosting(reqOpt) // Work Order Plan Report Resource
		case tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportOutput):
			ph = NewWorkOrderPlanReportOutputPosting(reqOpt) // Work Order Plan Report Output
		default:
			return nil, fmt.Errorf("missing: journal posting engine for type %s", payload.JournalType)
		}

		if pv, err := ph.HandlePostRequest(ctx, &payload); err != nil {
			errorTxts = append(errorTxts, err.Error())
		} else {
			res[index] = pv
		}
	}
	if len(errorTxts) > 0 {
		return res, errors.New(strings.Join(errorTxts, "\n"))
	}
	return res, nil
}

type PostOp string

const (
	PostOpPreview PostOp = "Preview"
	PostOpSubmit  PostOp = "Submit"
	PostOpApprove PostOp = "Approve"
	PostOpReject  PostOp = "Reject"
	PostOpPost    PostOp = "Post"
)

type PostRequest struct {
	JournalType tenantcoremodel.TrxModule
	JournalID   string
	Op          PostOp
	Text        string
}

func (obj *PostingProfileHandler) PostSingle(ctx *kaos.Context, request *PostRequest, postingEngine ficologic.JournalPosting) (*tenantcoremodel.PreviewReport, error) {
	var res *tenantcoremodel.PreviewReport
	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}
	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}
	if pr, err := scmlogic.PostJournal(postingEngine, userID, string(request.Op), request.Text); err != nil {
		return nil, err
	} else {
		res = pr
	}
	return res, nil
}

func (pph *PostingProfileHandler) GetApprovalBySource(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceRequest) (*ficomodel.PostingApproval, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	companyID, _, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pa, err := ficologic.GetPostingApprovalBySource(db, companyID, string(req.JournalType), req.JournalID, false)
	return pa, err
}

type MapSourceDataToURLRequest struct {
	SourceType string
	SourceID   string
}

type MapSourceDataToURLResponse struct {
	Menu      string
	URL       string
	JournalID string
}

func (pph *PostingProfileHandler) MapSourceDataToUrl(ctx *kaos.Context, req *MapSourceDataToURLRequest) (*MapSourceDataToURLResponse, error) {
	res := &MapSourceDataToURLResponse{
		Menu: req.SourceType,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: db")
	}

	url := strings.Replace(Config.AddrWebTenant, "/app", "/bagong", 1) // http://localhost:37000/app -> http://localhost:37000/bagong
	urlPath := mfgmodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(mfgmodel.JournalWorkOrderReportConsumption) {
		jORM := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderPlanReportConsumption))
		j, _ := jORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string(j.TrxType)
			res.JournalID = j.WorkOrderPlanID
			urlPath = mfgmodel.SourceTypeURLMap[string(mfgmodel.JournalWorkOrderPlan)]
		}
	}

	if req.SourceType == string(mfgmodel.JournalWorkOrderReportResource) {
		jORM := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderPlanReportResource))
		j, _ := jORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string(j.TrxType)
			res.JournalID = j.WorkOrderPlanID
			urlPath = mfgmodel.SourceTypeURLMap[string(mfgmodel.JournalWorkOrderPlan)]
		}
	}

	if req.SourceType == string(mfgmodel.JournalWorkOrderReportOutput) {
		jORM := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderPlanReportOutput))
		j, _ := jORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string(j.TrxType)
			res.JournalID = j.WorkOrderPlanID
			urlPath = mfgmodel.SourceTypeURLMap[string(mfgmodel.JournalWorkOrderPlan)]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
	return res, nil
}
