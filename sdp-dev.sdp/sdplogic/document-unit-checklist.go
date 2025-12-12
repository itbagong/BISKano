package sdplogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type DocumentUnitChecklistEngine struct{}

type DocumentUnitChecklistGetsSearch struct {
	SUNID    string
	ChasisNo string
	Status   string
}

type ResponseDocumentUnitChecklist struct {
	Count int         `json:"count"`
	Data  interface{} `json:"data"`
}

func (o *DocumentUnitChecklistEngine) GetsFilter(ctx *kaos.Context, payload *DocumentUnitChecklistGetsSearch) (ResponseDocumentUnitChecklist, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return ResponseDocumentUnitChecklist{}, errors.New("missing: connection")
	}
	documentUnitChecklist := []sdpmodel.DocumentUnitChecklist{}

	var queryParam *dbflex.QueryParam = dbflex.NewQueryParam()
	if payload.SUNID != "" {
		queryParam = queryParam.SetWhere(
			dbflex.Contains("SUNID", payload.SUNID),
		)
	}

	if payload.ChasisNo != "" {
		queryParam = queryParam.SetWhere(
			dbflex.Contains("ChasisNo", payload.ChasisNo),
		)
	}

	if payload.Status != "" {
		queryParam = queryParam.SetWhere(dbflex.Or(
			dbflex.Contains("StatusSamsat", payload.Status),
			dbflex.Contains("StatusSRUT", payload.Status),
			dbflex.Contains("StatusUjiKIR", payload.Status),
			dbflex.Contains("StatusFinal", payload.Status),
			dbflex.Contains("StatusRekomPeruntukan", payload.Status),
		))
	}

	err := h.Gets(new(sdpmodel.DocumentUnitChecklist), queryParam, &documentUnitChecklist)
	if err != nil {
		return ResponseDocumentUnitChecklist{}, fmt.Errorf("Search Document Unit Checklist: %s", err.Error())
	}

	return ResponseDocumentUnitChecklist{
		Data:  documentUnitChecklist,
		Count: len(documentUnitChecklist),
	}, nil
}

type SaveFromWORequest struct {
	WONo    string
	SunID   string
	HullNo  string
	AssetID string
}

func (o *DocumentUnitChecklistEngine) SaveFromWO(ctx *kaos.Context, payload *SaveFromWORequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return ResponseDocumentUnitChecklist{}, errors.New("missing: connection")
	}

	if payload == nil && payload.SunID == "" {
		return nil, errors.New("Missing payload")
	}

	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	//get by sun id
	document, _ := datahub.GetByFilter(h, new(sdpmodel.DocumentUnitChecklist), dbflex.Eq("SUNID", payload.SunID))
	document.WONo = payload.WONo
	document.SUNID = payload.SunID
	document.HullNo = payload.HullNo
	document.AssetID = payload.AssetID

	bagongasset := map[string]any{}
	send := []string{document.AssetID}

	err = event.Publish("/v1/bagong/asset/get", &send, &bagongasset, &kaos.PublishOpts{})
	if err != nil {
		return nil, fmt.Errorf("Error get data bagong: %v", err)
	}
	var DetailUnit map[string]any = bagongasset["DetailUnit"].(map[string]any)

	document.AssetName = payload.AssetID + " | " + bagongasset["Name"].(string)
	document.ChasisNo = DetailUnit["ChassisNum"].(string)
	document.EngineNo = DetailUnit["MachineNum"].(string)

	err = h.Save(document)
	if err != nil {
		return nil, err
	}
	return document, err
}
