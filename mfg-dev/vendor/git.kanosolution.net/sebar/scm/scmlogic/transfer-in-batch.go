package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type TransferInBatchEngine struct{}

type TransferInBatchSaveMultipleRequest struct {
	TransferID       string
	TransferInBatchs []scmmodel.TransferInBatch
}

func (o *TransferInBatchEngine) SaveMultiple(ctx *kaos.Context, payload *TransferInBatchSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.TransferID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	transferID := payload.TransferID
	e := h.DeleteByFilter(new(scmmodel.TransferInBatch), dbflex.Eq("TransferID", transferID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.TransferInBatchs {
		itemBatch.TransferID = payload.TransferID
		if e := h.GetByID(new(scmmodel.TransferInBatch), itemBatch.ID); e != nil {
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
