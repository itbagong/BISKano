package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type MovementInBatchEngine struct{}

type MovementBatchMultipleSaveRequest struct {
	MovementID     string
	MovementBatchs []scmmodel.MovementInItemBatch
}

func (o *MovementInBatchEngine) SaveMultiple(ctx *kaos.Context, payload *MovementBatchMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.MovementID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	movementInID := payload.MovementID
	e := h.DeleteByFilter(new(scmmodel.MovementInItemBatch), dbflex.Eq("MovementInID", movementInID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.MovementBatchs {
		itemBatch.MovementInID = payload.MovementID
		if e := h.GetByID(new(scmmodel.MovementInItemBatch), itemBatch.ID); e != nil {
			if e := h.Insert(&itemBatch); e != nil {
				return nil, errors.New("error insert Movement Batch: " + e.Error())
			}
		} else {
			if e := h.Save(&itemBatch); e != nil {
				return nil, errors.New("error update Movement Batch: " + e.Error())
			}
		}
	}

	return payload, nil
}
