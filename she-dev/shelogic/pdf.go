package shelogic

import (
	"errors"
	"fmt"
	"net/http"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type PDFLogic struct {
}

func (obj *PDFLogic) GeneratePdfByPreview(ctx *kaos.Context, _ *interface{}) ([]byte, error) {
	sourceJournalID := GetURLQueryParams(ctx)["SourceJournalID"]
	sourceType := GetURLQueryParams(ctx)["SourceType"]
	templateName := GetURLQueryParams(ctx)["TemplateName"]

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
	preview, err := datahub.GetByFilter(h, new(tenantcoremodel.Preview), filter)
	if err != nil {
		return nil, fmt.Errorf("preview report not found")
	}

	pdfData := codekit.M{
		"Data": preview,
	}
	// TemplateName: "JOURNAL_" + sourceType,
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
