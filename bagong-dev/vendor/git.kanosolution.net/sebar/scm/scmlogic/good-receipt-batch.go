package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type GoodReceiptBatchEngine struct{}

type GoodReceiptBatchMultipleSaveRequest struct {
	GoodReceiptID string
	Batchs        []scmmodel.GoodReceiptItemBatch
}

func (o *GoodReceiptBatchEngine) SaveMultiple(ctx *kaos.Context, payload *GoodReceiptBatchMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.GoodReceiptID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	goodReceiptID := payload.GoodReceiptID
	e := h.DeleteByFilter(new(scmmodel.GoodReceiptItemBatch), dbflex.Eq("GoodReceiptID", goodReceiptID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.Batchs {
		itemBatch.GoodReceiptID = payload.GoodReceiptID
		if e := h.GetByID(new(scmmodel.GoodReceiptItemBatch), itemBatch.ID); e != nil {
			if e := h.Insert(&itemBatch); e != nil {
				return nil, errors.New("error insert Good Receipt Batch: " + e.Error())
			}
		} else {
			if e := h.Save(&itemBatch); e != nil {
				return nil, errors.New("error update Good Receipt Batch: " + e.Error())
			}
		}
	}

	return payload, nil
}
