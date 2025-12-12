package hcmlogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/reflector"
)

type PostingProfileHandler struct {
}

type ResponseTrayekPosting struct {
	CustomerJournal []*tenantcoremodel.PreviewReport
	LedgerJournal   []*tenantcoremodel.PreviewReport
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payloads []ficologic.PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	res := make([]*tenantcoremodel.PreviewReport, len(payloads))
	if ctx == nil {
		return res, errors.New("missing: ctx")
	}

	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if coID == "DEMO" || coID == "" {
		return res, errors.New("missing: Company, please relogin")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		return res, errors.New("missing: User, please relogin")
	}

	for idx, req := range payloads {
		reqOpt := ficologic.PostingHubCreateOpt{
			Db:        h,
			UserID:    userID,
			CompanyID: coID,
			ModuleID:  string(req.JournalType),
			JournalID: req.JournalID,
			Op:        req.Op,
		}

		ph, err := obj.createPostingEngine(ctx, req, reqOpt)
		if err != nil {
			errorTxts = append(errorTxts, err.Error())
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

func (obj *PostingProfileHandler) createPostingEngine(ctx *kaos.Context, req ficologic.PostRequest, reqOpt ficologic.PostingHubCreateOpt) (PostRequestHandler, error) {
	var ph PostRequestHandler
	switch req.JournalType {
	case "CONTRACT":
		ph = NewContractPosting(reqOpt)
	case "OVERTIME":
		ph = NewOvertimePosting(reqOpt)
	case "LOAN":
		ph = NewLoanPosting(reqOpt)
	case "WORKTERMINATION":
		ph = NewWorkTerminationPosting(reqOpt)
	case "MANPOWER":
		ph = NewManpowerPosting(reqOpt)
	case "COACHING":
		ph = NewCoachingPosting(reqOpt)
	case "PLOTTING":
		ph = NewPlottingPosting(reqOpt)
	case "BUSINESSTRIP":
		ph = NewBusinessTripPosting(reqOpt)
	case "LEAVECOMPENSATION":
		ph = NewLeaveCompensationPosting(reqOpt)
	case "SK":
		ph = NewTalentDevelopmentSKPosting(reqOpt, ctx)
	case "ASSESSMENT":
		ph = NewTalentDevelopmentAssesmentPosting(reqOpt)
	case "TALENTDEVELOPMENT":
		ph = NewTalentDevelopmentPosting(reqOpt)
	case "TRAININGDEVELOPMENT":
		ph = NewTrainingDevelopmentPosting(reqOpt)
	case "TRAININGDEVELOPMENTDETAIL":
		ph = NewTrainingDevelopmentDetailPosting(reqOpt)

	default:
		return nil, fmt.Errorf("invalid module: %s", req.JournalType)
	}
	return ph, nil
}

func (obj *PostingProfileHandler) Reopen(ctx *kaos.Context, payloads []ficologic.PostRequest) (interface{}, error) {
	errorTxts := []string{}
	if ctx == nil {
		return nil, errors.New("missing: ctx")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	for _, req := range payloads {
		var model orm.DataModel
		switch req.JournalType {
		case "CONTRACT":
			model = new(hcmmodel.Contract)
		case "OVERTIME":
			model = new(hcmmodel.Overtime)
		case "LOAN":
			model = new(hcmmodel.Loan)
		case "WORKTERMINATION":
			model = new(hcmmodel.WorkTermination)
		case "MANPOWER":
			model = new(hcmmodel.ManpowerRequest)
		case "COACHING":
			model = new(hcmmodel.CoachingViolation)
		case "PLOTTING":
			model = new(hcmmodel.OLPlotting)
		case "BUSINESSTRIP":
			model = new(hcmmodel.BusinessTrip)
		case "LEAVECOMPENSATION":
			model = new(hcmmodel.LeaveCompensation)
		case "SK":
			model = new(hcmmodel.TalentDevelopmentSK)
		case "ASSESSMENT":
			model = new(hcmmodel.TalentDevelopmentAssesment)
		case "TALENTDEVELOPMENT":
			model = new(hcmmodel.TalentDevelopment)
		case "TRAININGDEVELOPMENT":
			model = new(hcmmodel.TrainingDevelopment)
		case "TRAININGDEVELOPMENTDETAIL":
			model = new(hcmmodel.TrainingDevelopmentDetail)
		default:
			errorTxts = append(errorTxts, fmt.Sprintf("invalid module: %s", req.JournalType))
			continue
		}

		err := db.GetByID(model, req.JournalID)
		if err != nil {
			errorTxts = append(errorTxts, fmt.Sprintf("error when get journal: %s - %s - %s", req.JournalID, req.JournalType, err.Error()))
			continue
		}

		reflector.From(model).Set("Status", ficomodel.JournalStatusDraft)
		err = db.Update(model, "Status")
		if err != nil {
			errorTxts = append(errorTxts, fmt.Sprintf("error when update journal: %s - %s - %s", req.JournalID, req.JournalType, err.Error()))
			continue
		}
	}

	if len(errorTxts) > 0 {
		return nil, errors.New(strings.Join(errorTxts, "\n"))
	}
	return "success", nil
}
