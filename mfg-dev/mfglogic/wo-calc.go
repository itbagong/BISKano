package mfglogic

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func WOCalcOutputValuePerQty(db *datahub.Hub, coID, woID string) float64 {
	filter := dbflex.And(
		dbflex.Eq("CompanyID", coID),
		dbflex.Eq("Status", ficomodel.AmountConfirmed),
		dbflex.ElemMatch("References", dbflex.Eq("Key", "WorkOrderID"), dbflex.Eq("Value", woID)),
		dbflex.ElemMatch("References", dbflex.Eq("Key", "WorkOrderElementType"), dbflex.Eq("Value", "Cost")),
	)

	// get total production cost
	ledgerTrxs, _ := datahub.FindByFilter(db, new(ficomodel.LedgerTransaction), filter)
	sumCost := lo.SumBy(ledgerTrxs, func(l *ficomodel.LedgerTransaction) float64 {
		if l.Amount > 0 {
			return l.Amount
		}
		return 0
	})

	// get all production output
	filter = dbflex.And(
		dbflex.Eq("CompanyID", coID),
		dbflex.Eq("Status", scmmodel.ItemConfirmed),
		dbflex.ElemMatch("References", dbflex.Eq("Key", "WorkOrderID"), dbflex.Eq("Value", woID)),
		dbflex.ElemMatch("References", dbflex.Eq("Key", "WorkOrderElementType"), dbflex.Eq("Value", "Output")),
	)
	inventTrxs, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), filter)

	wasteCost := lo.SumBy(inventTrxs, func(item *scmmodel.InventTrx) float64 {
		return lo.Ternary(
			item.References.Get("WorkOrderOutputType", "") == "Waste",
			item.AmountFinancial, 0)
	})

	outputValue := sumCost - wasteCost
	outputQty := lo.SumBy(inventTrxs, func(item *scmmodel.InventTrx) float64 {
		return lo.Ternary(
			item.References.Get("WorkOrderOutputType", "") == "Output",
			item.Qty, 0)
	})
	if outputQty == 0 {
		return 0
	}
	return outputValue / outputQty
}
