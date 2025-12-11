package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type MovementOutBatchEngine struct{}

type MovementOutBatchMultipleSaveRequest struct {
	MovementID     string
	MovementBatchs []scmmodel.MovementOutItemBatch
}

func (o *MovementOutBatchEngine) SaveMultiple(ctx *kaos.Context, payload *MovementOutBatchMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.MovementID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	movementID := payload.MovementID
	e := h.DeleteByFilter(new(scmmodel.MovementOutItemBatch), dbflex.Eq("MovementOutID", movementID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.MovementBatchs {
		itemBatch.MovementOutID = payload.MovementID
		if e := h.GetByID(new(scmmodel.MovementOutItemBatch), itemBatch.ID); e != nil {
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
