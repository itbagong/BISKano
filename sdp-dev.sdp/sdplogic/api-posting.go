package sdplogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpconfig"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

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

type PostingProfileHandler struct {
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payloads []ficologic.PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))

	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	eventhub, err := ctx.DefaultEvent()
	if err != nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)

	var (
		ph ficologic.PostRequestHandler
	)

	for idx, req := range payloads {
		reqOpt := ficologic.PostingHubCreateOpt{
			Db:        db,
			UserID:    userID,
			CompanyID: coID,
			ModuleID:  string(req.JournalType),
			JournalID: req.JournalID}

		switch req.JournalType {
		case tenantcoremodel.TrxModule(sdpmodel.SalsOrder):
			ph = NewSalesOrderPosting(reqOpt, eventhub)
		case tenantcoremodel.TrxModule(sdpmodel.SalsQuotation):
			ph = NewSalesQuotationPosting(reqOpt)

		default:
			errorTxts = append(errorTxts, fmt.Sprintf("invalid module: %s", req.JournalType))
			continue
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

func (pph *PostingProfileHandler) FindPostingProfile(ctx *kaos.Context, nothing string) ([]*ficomodel.PostingProfile, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	return ficologic.FindPostingApprovalByUserID(db, userID)
}

func (pph *PostingProfileHandler) GetApprovalBySource(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceRequest) (*ficomodel.PostingApproval, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	pa, err := ficologic.GetPostingApprovalBySource(db, companyID, string(req.JournalType), req.JournalID, false)
	return pa, err
}

func (pph *PostingProfileHandler) GetApprovalBySourceUser(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceUserRequest) (ficologic.PostingApprovalBySourceUserResponse, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	return ficologic.FindPostingApprovalBySourceUserID(db, userID, companyID, req, false)
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

	fmt.Println("tett")

	res := &MapSourceDataToURLResponse{
		Menu: req.SourceType,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: db")
	}

	url := strings.Replace(sdpconfig.Config.AddrWebTenant, "/app", "/bagong", 1) // http://localhost:37000/app -> http://localhost:37000/bagong
	urlPath := sdpmodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(sdpmodel.SalsOrder) {
		soORM := sebar.NewMapRecordWithORM(h, new(sdpmodel.SalesOrder))
		j, _ := soORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string("Sales Order")
			urlPath = sdpmodel.SourceTypeURLMap[string("Sales Order")]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
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
		case sdpmodel.SalsQuotation:
			source := new(sdpmodel.SalesQuotation)
			err := db.GetByID(source, req.JournalID)
			if err != nil {
				return nil, errors.New("missing: journal not found")
			}
			source.Status = ficomodel.JournalStatusDraft
			if err := db.Save(source); err != nil {
				return nil, errors.New("error: save journal")
			}
		case sdpmodel.SalsOrder:
			source := new(sdpmodel.SalesOrder)
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
