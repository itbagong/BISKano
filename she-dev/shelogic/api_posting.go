package shelogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/sheconfig"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type PostingProfileHandler struct {
}

type ResponseSHEPosting struct {
	CustomerJournal []*tenantcoremodel.PreviewReport
	LedgerJournal   []*tenantcoremodel.PreviewReport
}

func (obj *PostingProfileHandler) Post(ctx *kaos.Context, payload PostRequest) (interface{}, error) {
	// res := ResponseSHEPosting{}
	var res interface{}
	if ctx == nil {
		return res, errors.New("missing: ctx")
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

	// switch payload.JournalType {
	// case shemodel.PostTypeCoaching:
	ev, _ := ctx.DefaultEvent()
	var postinger ficologic.JournalPosting
	postinger, err := NewSHEPosting(ctx, h, ev, payload.JournalID, userID, coID, string(payload.JournalType))
	if err != nil {
		return res, fmt.Errorf("error new posting: %s", err.Error())
	}
	res, err = obj.PostSingle(ctx, &payload, postinger)
	if err != nil {
		return res, fmt.Errorf("error posting single: %s", err.Error())
	}

	return res, nil
	// default:
	// 	return res, fmt.Errorf("invalid module: %s", payload.JournalType)
	// }
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

type PostRequest struct {
	JournalType shemodel.PostType
	JournalID   string
	Op          shemodel.PostOp
	Text        string
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

	url := strings.Replace(sheconfig.Config.AddrWebTenant, "/app", "/she", 1) // http://localhost:37000/app -> http://localhost:37000/she
	urlPath := shemodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(shemodel.ModuleSidak) {
		soORM := sebar.NewMapRecordWithORM(h, new(shemodel.Sidak))
		j, _ := soORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string("Sidak")
			urlPath = shemodel.SourceTypeURLMap[string(shemodel.ModuleSidak)]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
	return res, nil
}
