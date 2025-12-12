package scmlogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
)

type PostingProfileHandlerV2 struct {
}

func (obj *PostingProfileHandlerV2) Post(ctx *kaos.Context, payloads []ficologic.PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
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
		userID = "SYSTEM"
	}

	coID, err := GetCompanyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var (
		ph ficologic.PostRequestHandler
	)

	for idx, req := range payloads {
		reqOpt := ficologic.PostingHubCreateOpt{
			Db:        db,
			UserID:    userID,
			CompanyID: coID,
			ModuleID:  string(req.JournalType),
			JournalID: req.JournalID,
			Op:        req.Op}

		switch req.JournalType {
		case tenantcoremodel.TrxModule(scmmodel.PurchOrder):
			ph = NewPurchaseOrderPosting(ctx, reqOpt) // Purchase Order
		case tenantcoremodel.TrxModule(scmmodel.PurchRequest):
			ph = NewPurchaseRequestPosting(ctx, reqOpt) // Purchase Request
		case scmmodel.ModuleInventory:
			ph = NewInventoryPosting(ctx, reqOpt) // Movement In, Movement Out
		case tenantcoremodel.TrxModule(scmmodel.InventReceive):
			ph = NewInventoryReceivePosting(ctx, reqOpt) // Good Receive
		case tenantcoremodel.TrxModule(scmmodel.InventIssuance):
			ph = NewInventoryIssuancePosting(ctx, reqOpt) // Good Issuance
		case tenantcoremodel.TrxModule(scmmodel.JournalTransfer):
			ph = NewInventoryTransferPosting(ctx, reqOpt) // Item Transfer
		case tenantcoremodel.TrxModule(scmmodel.ItemRequestType):
			ph = NewItemRequestPosting(ctx, reqOpt) // Item Request
		case tenantcoremodel.TrxModule(scmmodel.AssetAcquisitionTrxType):
			ph = NewAssetAcquisitionPosting(ctx, reqOpt) // Asset Acquisition

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

func (pph *PostingProfileHandlerV2) FindPostingProfile(ctx *kaos.Context, nothing string) ([]*ficomodel.PostingProfile, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	return ficologic.FindPostingApprovalByUserID(db, userID)
}

func (pph *PostingProfileHandlerV2) GetApprovalBySource(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceRequest) (*ficomodel.PostingApproval, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	companyID, err := GetCompanyIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pa, err := ficologic.GetPostingApprovalBySource(db, companyID, string(req.JournalType), req.JournalID, false)
	return pa, err
}

func (pph *PostingProfileHandlerV2) GetApprovalBySourceUser(ctx *kaos.Context, req *ficologic.GetPostingApprovalBySourceUserRequest) (ficologic.PostingApprovalBySourceUserResponse, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	db := sebar.GetTenantDBFromContext(ctx)
	companyID, err := GetCompanyIDFromContext(ctx)
	if err != nil {
		return ficologic.PostingApprovalBySourceUserResponse{}, err
	}
	// companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)
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

func (pph *PostingProfileHandlerV2) MapSourceDataToUrl(ctx *kaos.Context, req *MapSourceDataToURLRequest) (*MapSourceDataToURLResponse, error) {
	res := &MapSourceDataToURLResponse{
		Menu: req.SourceType,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: db")
	}

	url := strings.Replace(scmconfig.Config.AddrWebTenant, "/app", "/bagong", 1) // http://localhost:37000/app -> http://localhost:37000/bagong
	urlPath := scmmodel.SourceTypeURLMap[req.SourceType]

	if req.SourceType == string(scmmodel.ModuleInventory) {
		invORM := sebar.NewMapRecordWithORM(h, new(scmmodel.InventJournal))
		j, _ := invORM.Get(req.SourceID)
		if j.ID != "" {
			res.Menu = string(j.TrxType)
			urlPath = scmmodel.SourceTypeURLMap[string(j.TrxType)]
		}
	}

	res.URL = fmt.Sprintf("%s/%s", url, urlPath)
	return res, nil
}

// BreakdownInventTrxDataModelMap will convert back map[string][]orm.DataModel -> []*scmmodel.InventTrx
func BreakdownInventTrxDataModelMap(trxs map[string][]orm.DataModel, tableName ...string) []*scmmodel.InventTrx {
	resTrxs := []*scmmodel.InventTrx{}

	for tbName, trxLines := range trxs {
		if len(tableName) > 0 && tableName[0] != "" && tbName != tableName[0] {
			continue
		}

		fTrxs := lo.FilterMap(trxLines, func(d orm.DataModel, i int) (*scmmodel.InventTrx, bool) {
			if dtrx, ok := d.(*scmmodel.InventTrx); ok {
				return dtrx, true
			}

			return nil, false
		})

		resTrxs = append(resTrxs, fTrxs...)
	}

	return resTrxs
}
