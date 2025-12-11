package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
)

type ItemSpecEngine struct{}

type ItemSpecUsageCheckParam struct {
	SpecIDs []string
}

func (o *ItemSpecEngine) UsageCheck(ctx *kaos.Context, p *ItemSpecUsageCheckParam) (*UsageCheckResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	res := &UsageCheckResponse{IsUsed: false, UsedIn: []UsageCheckResUsedIn{}}

	inventJs := []scmmodel.InventJournal{}
	h.GetsByFilter(new(scmmodel.InventJournal), dbflex.In("Lines.SKU", p.SpecIDs...), &inventJs)

	if len(inventJs) > 0 {
		res.IsUsed = true
		for _, j := range inventJs {
			res.UsedIn = append(res.UsedIn, UsageCheckResUsedIn{
				Type:      string(j.TrxType),
				JournalID: j.ID,
			})
		}
	}

	poJs := []scmmodel.PurchaseOrderJournal{}
	h.GetsByFilter(new(scmmodel.PurchaseOrderJournal), dbflex.In("Lines.InventJournalLine.SKU", p.SpecIDs...), &poJs)

	if len(poJs) > 0 {
		res.IsUsed = true
		for _, j := range poJs {
			res.UsedIn = append(res.UsedIn, UsageCheckResUsedIn{
				Type:      string(scmmodel.PurchOrder),
				JournalID: j.ID,
			})
		}
	}

	prJs := []scmmodel.PurchaseRequestJournal{}
	h.GetsByFilter(new(scmmodel.PurchaseRequestJournal), dbflex.In("Lines.InventJournalLine.SKU", p.SpecIDs...), &prJs)

	if len(prJs) > 0 {
		res.IsUsed = true
		for _, j := range prJs {
			res.UsedIn = append(res.UsedIn, UsageCheckResUsedIn{
				Type:      string(scmmodel.PurchRequest),
				JournalID: j.ID,
			})
		}
	}

	irDetails := []scmmodel.ItemRequestDetail{}
	h.GetsByFilter(new(scmmodel.ItemRequestDetail), dbflex.In("SKU", p.SpecIDs...), &irDetails)

	if len(irDetails) > 0 {
		irMap := lo.GroupBy(irDetails, func(d scmmodel.ItemRequestDetail) string {
			return d.ItemRequestID
		})

		res.IsUsed = true
		for irID := range irMap {
			res.UsedIn = append(res.UsedIn, UsageCheckResUsedIn{
				Type:      string(scmmodel.ItemRequestType),
				JournalID: irID,
			})
		}
	}

	return res, nil
}
