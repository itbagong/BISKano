package ficologic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type PostingProfileHandler struct {
}

func (obj *PostingProfileHandler) PreviewDownloadAsPdf(ctx *kaos.Context, _ *interface{}) ([]byte, error) {
	sourceJournalID := GetURLQueryParams(ctx)["SourceJournalID"]
	sourceType := GetURLQueryParams(ctx)["SourceType"]

	w, wOK := ctx.Data().Get("http_writer", nil).(http.ResponseWriter)

	if !wOK {
		return nil, errors.New("not a http compliant writer")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("nil: EventHub")
	}

	filter := dbflex.Eqs("SourceType", sourceType, "SourceJournalID", sourceJournalID)
	pv, err := datahub.GetByParm(h, new(tenantcoremodel.Preview), dbflex.NewQueryParam().SetWhere(filter).SetSort("-Created"))

	if err != nil {
		return nil, fmt.Errorf("Preview report not found: %s ", err)
	}

	pdfData := codekit.M{
		"Data": pv,
	}
	// TemplateName: "JOURNAL_" + sourceType,
	templateName := "JOURNAL_VENDOR"
	if sourceType == "MCU" || sourceType == "MCU_FOLLOWUP" {
		templateName = "MCU_REFERRAL"
	} else if sourceType == "CASH OUT" {
		templateName = "multiple-row-signature"
	} else if sourceType == "Work Order" {
		templateName = "JOURNAL_WO"
	} else if sourceType == "Purchase Order" {
		templateName = "JOURNAL_PO"
	} else if sourceType == "CASHBANK" {
		templateName = "multiple-row-signature"
	} else if sourceType == "CUSTOMER" && pv.Name == "Mining Invoice - Rent" {
		templateName = "mining-invoice"
	}
	content, e := templateToPDF(&tenantcorelogic.PDFFromTemplateRequest{
		TemplateName: templateName,
		Data:         pdfData,
	}, ev)
	if e != nil {
		return nil, e
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(content)

	ctx.Data().Set("kaos_command_1", "stop")
	return content, nil
}

func templateToPDF(param *tenantcorelogic.PDFFromTemplateRequest, ev kaos.EventHub) ([]byte, error) {
	url := "/v1/tenant/pdf/from-template"

	res := tenantcorelogic.PDFByteResponse{}
	err := ev.Publish(
		url,
		param,
		&res,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return res.PDFByte, nil
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payloads []PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))

	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		if ctx.Data().Get("UserID", "").(string) != "" {
			userID = ctx.Data().Get("UserID", "").(string)
		}
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	for idx, req := range payloads {
		reqOpt := PostingHubCreateOpt{
			Db:        db,
			UserID:    userID,
			CompanyID: coID,
			ModuleID:  string(req.JournalType),
			JournalID: req.JournalID,
			Op:        req.Op,
		}

		ph, err := createPostingEngine(req, reqOpt)
		if err != nil {
			errorTxts = append(errorTxts, err.Error())
		}

		if pv, err := ph.HandlePostRequest(ctx, &req); err != nil {
			errorTxts = append(errorTxts, err.Error())
		} else {
			res[idx] = pv
		}
	}

	if len(errorTxts) > 0 {
		return res, errors.New(strings.Join(errorTxts, "\n"))
	}
	return res, nil
}

func createPostingEngine(req PostRequest, reqOpt PostingHubCreateOpt) (PostRequestHandler, error) {
	var ph PostRequestHandler
	switch req.JournalType {
	case ficomodel.SubledgerAccounting:
		ph = NewLedgerPosting(reqOpt)

	case ficomodel.SubledgerCustomer:
		ph = NewCustomerPosting(reqOpt)

	case ficomodel.SubledgerVendor:
		ph = NewVendorPosting(reqOpt)

	case ficomodel.SubledgerAsset:
		ph = NewAssetPosting(reqOpt)

	case ficomodel.SubledgerCashBank:
		ph = NewCashPosting(reqOpt)

	default:
		return nil, fmt.Errorf("invalid module: %s", req.JournalType)
	}
	return ph, nil
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

type EventPostRequest struct {
	UserID      string
	CompanyID   string
	PostRequest []PostRequest
}

func (pph *PostingProfileHandler) FindPostingProfile(ctx *kaos.Context, nothing string) ([]*ficomodel.PostingProfile, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	return FindPostingApprovalByUserID(db, userID)
}

type GetPostingApprovalBySourceRequest struct {
	JournalType tenantcoremodel.TrxModule
	JournalID   string
}

type GetPostingApprovalBySourceUserRequest struct {
	ApprovalType string
	JournalType  tenantcoremodel.TrxModule
	JournalID    string
}

type PostingApprovalBySourceUserResponse struct {
	Approvers  bool
	Submitters bool
	Postingers bool
}

func (pph *PostingProfileHandler) GetApprovalBySource(ctx *kaos.Context, req *GetPostingApprovalBySourceRequest) (*ficomodel.PostingApproval, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	pa, err := GetPostingApprovalBySource(db, companyID, string(req.JournalType), req.JournalID, false)
	return pa, err
}

func (pph *PostingProfileHandler) GetApprovalBySourceUser(ctx *kaos.Context, req *GetPostingApprovalBySourceUserRequest) (PostingApprovalBySourceUserResponse, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	return FindPostingApprovalBySourceUserID(db, userID, companyID, req, false)
}

type MapSourceDataToURLRequest struct {
	SourceType string
	SourceID   string
}

type MapSourceDataToURLResponse struct {
	Menu string
	URL  string
}

func (pph *PostingProfileHandler) MapSourceDataToUrl(ctx *kaos.Context, req *MapSourceDataToURLRequest) (*MapSourceDataToURLResponse, error) {
	res := &MapSourceDataToURLResponse{
		Menu: req.SourceType,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: db")
	}

	url := strings.Replace(ficoconfig.Config.AddrWebTenant, "/app", "/bagong", 1) // http://localhost:37000/app -> http://localhost:37000/bagong
	urlPath := ficomodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(ficomodel.SubledgerCashBank) {
		cashORM := sebar.NewMapRecordWithORM(h, new(ficomodel.CashJournal))
		j, _ := cashORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string(j.CashJournalType)
			urlPath = ficomodel.SourceTypeURLMap[string(j.CashJournalType)]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
	return res, nil
}

type IsLatestApproverRequest struct {
	SourceType string
	SourceID   string
}

type IsLatestApproverRespond struct {
	LatestApprover bool
}

func (pph *PostingProfileHandler) IsLatestApprover(ctx *kaos.Context, req *IsLatestApproverRequest) (*IsLatestApproverRespond, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	userID := sebar.GetUserIDFromCtx(ctx)
	if coID == "" || userID == "" {
		return nil, fmt.Errorf("missing: companyID or UserID")
	}
	pa, err := GetPostingApprovalBySource(db, coID, req.SourceType, req.SourceID, true)
	if err != nil {
		return nil, fmt.Errorf("missing: posting approval: %s", err.Error())
	}
	res := &IsLatestApproverRespond{
		LatestApprover: IsLastApprover(pa, userID),
	}
	return res, nil
}

func (obj *PostingProfileHandler) Reopen(ctx *kaos.Context, payloads []PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))

	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	for _, req := range payloads {
		switch req.JournalType {
		case ficomodel.SubledgerAccounting:
			source := new(ficomodel.LedgerJournal)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		case ficomodel.SubledgerCustomer:
			source := new(ficomodel.CustomerJournal)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		case ficomodel.SubledgerVendor:
			source := new(ficomodel.VendorJournal)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		case ficomodel.SubledgerAsset:
			source := new(ficomodel.LedgerJournal)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		case ficomodel.SubledgerCashBank:
			source := new(ficomodel.CashJournal)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		default:
			errorTxts = append(errorTxts, fmt.Sprintf("invalid module: %s", req.JournalType))
			continue
		}
	}

	if len(errorTxts) > 0 {
		return res, errors.New(strings.Join(errorTxts, "\n"))
	}
	return res, nil
}
