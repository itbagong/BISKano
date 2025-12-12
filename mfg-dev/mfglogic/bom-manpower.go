package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
)

type BoMManpowerEngine struct{}

type BoMManpowerSaveMultipleRequest struct {
	BoMID        string
	BoMManpowers []mfgmodel.BoMManpower
}

func (o *BoMManpowerEngine) SaveMultiple(ctx *kaos.Context, payload *BoMManpowerSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload.BoMID == "" {
		return nil, fmt.Errorf("missing: payload")
	}

	if e := h.DeleteByFilter(new(mfgmodel.BoMManpower), dbflex.Eq("BoMID", payload.BoMID)); e != nil {
		return nil, fmt.Errorf("error clear bom man powers: " + e.Error())
	}

	for _, dt := range payload.BoMManpowers {
		dt.BoMID = payload.BoMID
		if e := h.Save(&dt); e != nil {
			return nil, fmt.Errorf("error update bom man power: " + e.Error())
		}
	}

	return payload, nil
}
