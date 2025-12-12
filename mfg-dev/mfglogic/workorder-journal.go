package mfglogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type WorkOrderJournalEngine struct{}

type WOJournalSaveRequest struct {
	WOID     string
	WOStatus mfgmodel.WOStatus
	Lines    []scmmodel.InventReceiveIssueLine
}

func (o *WorkOrderJournalEngine) Save(ctx *kaos.Context, payload *WOJournalSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//Get WO
	workOrders, e := datahub.GetByID(h, new(mfgmodel.WorkOrder), payload.WOID)
	if e != nil {
		return nil, fmt.Errorf("Work Order not found")
	}

	workOrders.Status = payload.WOStatus
	if e = h.Save(workOrders); e != nil {
		return nil, e
	}

	now := time.Now()
	//transform line
	woJournal := new(mfgmodel.WorkOrderJournal)
	woJournal.CompanyID = workOrders.CompanyID
	woJournal.Name = workOrders.Name
	woJournal.JournalTypeID = workOrders.JournalTypeID
	woJournal.PostingProfileID = workOrders.PostingProfileID
	woJournal.ItemUsage = payload.Lines
	woJournal.TrxDate = &now
	woJournal.TrxType = scmmodel.JournalWorkOrder
	woJournal.Status = ficomodel.JournalStatusDraft
	woJournal.InventDim = workOrders.InventDim
	woJournal.Dimension = workOrders.Dimension

	if e = h.Save(woJournal); e != nil {
		return nil, e
	}

	return woJournal, nil
}
