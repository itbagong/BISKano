package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
)

type RoutineDetailEngine struct{}

type RoutineDetailSaveMultipleRequest struct {
	RoutineID      string
	RoutineDetails []mfgmodel.RoutineDetail
}

func (o *RoutineDetailEngine) SaveMultiple(ctx *kaos.Context, payload *RoutineDetailSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload.RoutineID == "" {
		return nil, fmt.Errorf("missing: payload")
	}

	if e := h.DeleteByFilter(new(mfgmodel.RoutineDetail), dbflex.Eq("RoutineID", payload.RoutineID)); e != nil {
		return nil, fmt.Errorf("error clear item serials: " + e.Error())
	}

	for _, dt := range payload.RoutineDetails {
		dt.RoutineID = payload.RoutineID
		if e := h.Save(&dt); e != nil {
			return nil, fmt.Errorf("error update Movement Serial: " + e.Error())
		}
	}

	return payload, nil
}
