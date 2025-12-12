package mfglogic

import (
	"errors"
	"net/http"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type PreviewLogic struct {
}

type PreviewPayload struct {
	SourceType      string
	SourceJournalID string
	Name            string
	VoucherNo       string
	Reload          bool
}

func (pv *PreviewLogic) Get(ctx *kaos.Context, payload *PreviewPayload) (*tenantcoremodel.PreviewReport, error) {
	if payload == nil {
		qr, ok := ctx.Data().Get("http_request", nil).(*http.Request)
		if !ok {
			return nil, errors.New("payload is nil")
		}

		payload = new(PreviewPayload)
		query := qr.URL.Query()
		payload.SourceType = query.Get("type")
		payload.SourceJournalID = query.Get("id")
		payload.Name = query.Get("name")
		payload.VoucherNo = query.Get("voucher")
		payload.Reload = query.Get("reload") == "1"
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}
	if payload.Name == "" {
		payload.Name = "Default"
	}

	preview, err := datahub.GetByFilter(db, new(tenantcoremodel.Preview),
		dbflex.Eqs("SourceType", payload.SourceType, "SourceJournalID", payload.SourceJournalID, "Name", payload.Name, "VoucherNo", payload.VoucherNo))
	if err == nil && preview.PreviewReport.Sections != nil && !payload.Reload {
		return preview.PreviewReport, nil
	}

	if payload.Name != "Default" {
		return nil, errors.New("missing: preview")
	}

	ph := new(PostingProfileHandler)
	previews, err := ph.Post(ctx, []ficologic.PostRequest{{JournalType: tenantcoremodel.TrxModule(payload.SourceType), JournalID: payload.SourceJournalID, Op: ficologic.PostOpPreview}})
	if err != nil {
		return nil, err
	}
	if len(previews) == 0 {
		return nil, errors.New("preview len is 0")
	}
	return previews[0], nil
}
