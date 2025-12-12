package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type InventoryAdjustmentEngine struct{}

type InventoryAdjustmentProcessRequest struct {
	InventoryAdjustment scmmodel.InventoryAdjustment
	Lines               []scmmodel.InventoryAdjustmentDetail
}

func (o *InventoryAdjustmentEngine) Process(ctx *kaos.Context, payload *InventoryAdjustmentProcessRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	payload.InventoryAdjustment.Status = scmmodel.InventoryAdjustmentStatusDone
	if e := h.Save(&payload.InventoryAdjustment); e != nil {
		return nil, e
	}

	// TODO: is it possible to use trx, transactional mechanism with rollback?

	for _, d := range payload.Lines {
		balance := new(scmmodel.ItemBalance)
		if e := h.GetByFilter(balance, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
			ItemID:             d.ItemID,
			SKU:                d.SKU,
			InventoryDimension: d.InventoryDimension,
		})); e != nil {
			return nil, e
		}

		gap := float64(d.QtyActual) - balance.Qty // recalculate, just to be sure
		balance.Qty = float64(d.QtyActual)
		balance.QtyAvail = balance.QtyAvail + gap // bcoz QtyAvail is not always same with Qty, Move Out process that can make the difference
		if e := h.Save(balance); e != nil {
			return nil, e
		}

		// TODO: need to add item transaction as well? [TBC: Mas Eky]
	}

	// TODO: update journal in FICO [PENDING: TBD]

	return nil, nil
}
