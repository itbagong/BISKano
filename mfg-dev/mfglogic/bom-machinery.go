package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
)

type BoMMachineryEngine struct{}

type BoMMachinerySaveMultipleRequest struct {
	BoMID        string
	BoMMachinery []mfgmodel.BoMMachinery
}

func (o *BoMMachineryEngine) SaveMultiple(ctx *kaos.Context, payload *BoMMachinerySaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload.BoMID == "" {
		return nil, fmt.Errorf("missing: payload")
	}

	if e := h.DeleteByFilter(new(mfgmodel.BoMMachinery), dbflex.Eq("BoMID", payload.BoMID)); e != nil {
		return nil, fmt.Errorf("error clear bom machinery: " + e.Error())
	}

	for _, dt := range payload.BoMMachinery {
		dt.BoMID = payload.BoMID
		if e := h.Save(&dt); e != nil {
			return nil, fmt.Errorf("error update bom machinery: " + e.Error())
		}
	}

	return payload, nil
}
