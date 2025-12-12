package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
)

type WorkRequestEngine struct{}

type WRApproveRequest struct {
	WRID string
}

func (o *WorkRequestEngine) Approve(ctx *kaos.Context, payload *WRApproveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//Get WO
	workRequest, e := datahub.GetByID(h, new(mfgmodel.WorkRequest), payload.WRID)
	if e != nil {
		return nil, fmt.Errorf("Work Request not found")
	}

	wo, _ := reflector.CopyAttributes(workRequest, new(mfgmodel.WorkOrder))
	wo.WorkRequestID = workRequest.ID
	wo.Source = mfgmodel.FromWorkRequest

	e = h.Save(wo)
	return workRequest, nil
}
