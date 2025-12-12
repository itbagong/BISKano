package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type MovementInDetailEngine struct{}

type MovementDetailMultipleSaveRequest struct {
	MovementID    string
	MovementItems []scmmodel.MovementInDetail
}

func (o *MovementInDetailEngine) SaveMultiple(ctx *kaos.Context, payload *MovementDetailMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.MovementID == "" {
		return nil, errors.New("missing: payload")
	}

	mov := new(scmmodel.MovementIn)
	if e := h.GetByID(mov, payload.MovementID); e != nil {
		return nil, errors.New("no movement in data found: " + e.Error())
	}

	if e := h.DeleteByFilter(new(scmmodel.MovementInDetail), dbflex.Eq("MovementInID", payload.MovementID)); e != nil {
		return nil, errors.New("error clear item details: " + e.Error())
	}

	for _, itemMovement := range payload.MovementItems {
		itemMovement.MovementInID = payload.MovementID
		itemMovement.InventoryDimension = mov.InventoryDimension
		if e := h.GetByID(new(scmmodel.MovementInDetail), itemMovement.ID); e != nil {
			if e := h.Insert(&itemMovement); e != nil {
				return nil, errors.New("error insert Movement  Detail: " + e.Error())
			}
		} else {
			if e := h.Save(&itemMovement); e != nil {
				return nil, errors.New("error update Movement  Detail: " + e.Error())
			}
		}
	}

	return payload, nil
}
