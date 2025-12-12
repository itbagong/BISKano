package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type InventoryAdjustmentDetailEngine struct{}

type InventoryAdjustmentDetailSaveMultipleRequest struct {
	InventoryAdjustmentID      string
	InventoryAdjustmentDetails []scmmodel.InventoryAdjustmentDetail
}

func (o *InventoryAdjustmentDetailEngine) SaveMultiple(ctx *kaos.Context, payload *InventoryAdjustmentDetailSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.InventoryAdjustmentID == "" {
		return nil, errors.New("missing: payload")
	}

	err := sebar.Tx(h, true, func(tx *datahub.Hub) error {
		p := new(scmmodel.InventoryAdjustment)
		if e := tx.GetByID(p, payload.InventoryAdjustmentID); e != nil {
			return errors.New("no item request data found: " + e.Error())
		}

		if e := tx.DeleteByFilter(new(scmmodel.InventoryAdjustmentDetail), dbflex.Eq("InventoryAdjustmentID", payload.InventoryAdjustmentID)); e != nil {
			return errors.New("error clear ia details: " + e.Error())
		}

		for _, dt := range payload.InventoryAdjustmentDetails {
			dt.InventoryAdjustmentID = payload.InventoryAdjustmentID
			if e := tx.Save(&dt); e != nil {
				return errors.New("error update Movement  Detail: " + e.Error())
			}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return payload, nil
}
