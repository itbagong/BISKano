package scmlogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type PostingProfileHandler struct {
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payloads []PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))

	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	ev, _ := ctx.DefaultEvent()

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	for index, payload := range payloads {
		var postinger ficologic.JournalPosting
		switch payload.JournalType {
		case scmmodel.ModuleInventory:
			//postinger = NewInventJournalPosting(h, ev, payload.JournalID, userID)
		case tenantcoremodel.TrxModule(scmmodel.InventReceive):
			postinger = NewInventReceivePosting(h, ev, payload.JournalID, userID)
		case tenantcoremodel.TrxModule(scmmodel.InventIssuance):
			postinger = NewInventIssuancePosting(h, ev, payload.JournalID, userID)
		case tenantcoremodel.TrxModule(scmmodel.JournalTransfer):
			postinger = NewInventTransferJournalPosting(h, ev, payload.JournalID, userID)
		case tenantcoremodel.TrxModule(scmmodel.JournalOpname):
			postinger = NewStockOpnameJournalPosting(h, ev, payload.JournalID, userID)
		// case scmmodel.ModulePurchase:
		// 	postinger = NewPurchaseOrderPosting(h, payload.JournalID, userID)
		default:
			return nil, fmt.Errorf("missing: journal posting engine for type %s", payload.JournalType)
		}
		m, err := obj.PostSingle(ctx, &payload, postinger)
		if err != nil {
			res[index] = nil
			errorTxts = append(errorTxts, err.Error())
		} else {
			res[index] = m
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
	if pr, err := PostJournal(postingEngine, userID, string(request.Op), request.Text); err != nil {
		return nil, err
	} else {
		res = pr
	}
	return res, nil
}

func (pph *PostingProfileHandler) GetApprovalBySource(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceRequest) (*ficomodel.PostingApproval, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	companyID, err := GetCompanyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	pa, err := ficologic.GetPostingApprovalBySource(db, companyID, string(req.JournalType), req.JournalID, false)
	return pa, err
}
