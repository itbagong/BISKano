package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type TransferOutBatchEngine struct{}

type TransferOutBatchSaveMultipleRequest struct {
	TransferID        string
	TransferOutBatchs []scmmodel.TransferOutBatch
}

func (o *TransferOutBatchEngine) SaveMultiple(ctx *kaos.Context, payload *TransferOutBatchSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.TransferID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	transferID := payload.TransferID
	e := h.DeleteByFilter(new(scmmodel.TransferOutBatch), dbflex.Eq("TransferID", transferID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.TransferOutBatchs {
		itemBatch.TransferID = payload.TransferID
		if e := h.GetByID(new(scmmodel.TransferOutBatch), itemBatch.ID); e != nil {
			if e := h.Insert(&itemBatch); e != nil {
				return nil, errors.New("error insert transfer out batch: " + e.Error())
			}
		} else {
			if e := h.Save(&itemBatch); e != nil {
				return nil, errors.New("error update Movement Batch: " + e.Error())
			}
		}

		itemBatchIn := new(scmmodel.TransferInBatch)
		itemBatchIn.TransferID = payload.TransferID
		itemBatchIn.ItemID = itemBatch.ItemID
		itemBatchIn.SKU = itemBatch.SKU
		itemBatchIn.Qty = itemBatch.Qty

		if e := h.Insert(itemBatchIn); e != nil {
			return nil, errors.New("error insert transfer in batch: " + e.Error())
		}
	}

	return payload, nil
}
