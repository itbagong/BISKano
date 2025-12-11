package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type InventoryAdjustmentEngine struct{}

type InventoryAdjustmentProcessRequest struct {
	InventoryAdjustment scmmodel.InventoryAdjustment
	Lines               []scmmodel.InventoryAdjustmentDetail
}

type InventoryAdjustmentProcessParam struct {
	ID string
}

func (o *InventoryAdjustmentEngine) Process(ctx *kaos.Context, payload *InventoryAdjustmentProcessParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	ia := new(scmmodel.InventoryAdjustment)
	if e := h.GetByID(ia, payload.ID); e != nil {
		return nil, e
	}

	iaLines := []scmmodel.InventoryAdjustmentDetail{}
	if e := h.GetsByFilter(new(scmmodel.InventoryAdjustmentDetail), dbflex.Eq("InventoryAdjustmentID", payload.ID), &iaLines); e != nil {
		return nil, e
	}

	// for _, d := range payload.Lines {
	// 	balance := new(scmmodel.ItemBalance)
	// 	if e := h.GetByFilter(balance, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
	// 		ItemID:             d.ItemID,
	// 		SKU:                d.SKU,
	// 		InventoryDimension: d.InventoryDimension,
	// 	})); e != nil {
	// 		return nil, e
	// 	}

	// 	gap := float64(d.QtyActual) - balance.Qty // recalculate, just to be sure
	// 	balance.Qty = float64(d.QtyActual)
	// 	balance.QtyAvail = balance.QtyAvail + gap // bcoz QtyAvail is not always same with Qty, Move Out process that can make the difference
	// 	if e := h.Save(balance); e != nil {
	// 		return nil, e
	// 	}

	// 	// TODO: need to add item transaction as well? [TBC: Mas Eky]
	// }

	// // TODO: update journal in FICO [PENDING: TBD]

	return nil, o.process(h, ia, iaLines)
}

func (o *InventoryAdjustmentEngine) process(h *datahub.Hub, ia *scmmodel.InventoryAdjustment, iaLines []scmmodel.InventoryAdjustmentDetail) error {
	err := sebar.Tx(h, true, func(tx *datahub.Hub) error {
		var err error
		itemORM := sebar.NewMapRecordWithORM(tx, new(tenantcoremodel.Item))
		journalID := ia.ID

		// build line trx
		inventTrxs := []*scmmodel.InventTrx{}
		for _, iaLine := range iaLines {
			item, _ := itemORM.Get(iaLine.ItemID)
			qty := (iaLine.QtyActual - iaLine.QtyInSystem)

			// convert to default unit
			unitRatio, err := ConvertUnit(tx, 1, iaLine.UoM, item.DefaultUnitID)
			if err != nil {
				return fmt.Errorf("invalid: unit for item %s: %s", item.ID, err.Error())
			}

			inventQty := unitRatio * qty

			trx := &scmmodel.InventTrxLine{
				InventJournalLine: scmmodel.InventJournalLine{
					ItemID:    iaLine.ItemID,
					LineNo:    iaLine.LineNo,
					SKU:       iaLine.SKU,
					Qty:       inventQty,
					ReffNo:    journalID,
					UnitID:    iaLine.UoM,
					Text:      tenantcorelogic.ItemVariantName(tx, iaLine.ItemID, iaLine.SKU),
					UnitCost:  0, // fill in below
					Dimension: ia.Dimension,
					InventDim: iaLine.InventoryDimension,
					Remarks:   "Adjusted from Stock Opname",
				},
				InventQty:   inventQty,
				JournalID:   journalID,
				TrxType:     scmmodel.JournalOpname,
				TrxDate:     ia.AdjustmentDate,
				Item:        item,
				CostPerUnit: 0, // fill in below
			}

			// calc unit cost
			trx.CostPerUnit = GetCostPerUnit(tx, *trx.Item, trx.InventDim, &ia.AdjustmentDate)
			trx.UnitCost = trx.CostPerUnit * unitRatio

			inventTrx, err := trxLineToInventTrx(trx, tx)
			if err != nil {
				return fmt.Errorf("create inventory transaction: %s", err.Error())
			}
			inventTrxs = append(inventTrxs, inventTrx)
		}

		ia.Status = scmmodel.InventoryAdjustmentStatusDone
		if e := tx.Save(ia); e != nil {
			return e
		}

		// track the inventory and posting it as RESVD or PLANNED
		tx.DeleteByFilter(new(scmmodel.InventTrx), dbflex.Eqs(
			"CompanyID", ia.CompanyID,
			"SourceType", scmmodel.JournalOpname,
			"SourceJournalID", journalID,
		))

		trxs := map[string][]orm.DataModel{}
		trxs[inventTrxs[0].TableName()] = ficologic.ToDataModels(inventTrxs)

		inventTrxModels := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
		for _, trx := range inventTrxModels {
			trx.CompanyID = ia.CompanyID
			trx.Status = scmmodel.ItemConfirmed
			tx.Save(trx)
		}

		// sync item balance
		balanceOpt := ItemBalanceOpt{}
		balanceOpt.CompanyID = ia.CompanyID
		balanceOpt.DisableGrouping = true
		balanceOpt.ConsiderSKU = true
		balanceOpt.ItemIDs = lo.Map(inventTrxs, func(t *scmmodel.InventTrx, index int) string {
			return t.Item.ID
		})

		_, err = NewItemBalanceHub(tx).Sync(nil, balanceOpt)

		return err
	})
	if err != nil {
		return err
	}

	return nil
}
