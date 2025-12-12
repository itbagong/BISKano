package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
)

type BoMMaterialEngine struct{}

type BoMMaterialSaveMultipleRequest struct {
	BoMID        string
	BoMMaterials []mfgmodel.BoMMaterial
}

func (o *BoMMaterialEngine) SaveMultiple(ctx *kaos.Context, payload *BoMMaterialSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload.BoMID == "" {
		return nil, fmt.Errorf("missing: payload")
	}

	if e := h.DeleteByFilter(new(mfgmodel.BoMMaterial), dbflex.Eq("BoMID", payload.BoMID)); e != nil {
		return nil, fmt.Errorf("error clear bom materials: " + e.Error())
	}

	for _, dt := range payload.BoMMaterials {
		dt.BoMID = payload.BoMID
		if e := h.Save(&dt); e != nil {
			return nil, fmt.Errorf("error update bom material: " + e.Error())
		}
	}

	return payload, nil
}
