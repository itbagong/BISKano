package scmlogic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type ItemRequestDetailEngine struct{}

type ItemRequestDetailSaveMultipleRequest struct {
	ItemRequestID      string
	ItemRequestDetails []scmmodel.ItemRequestDetail
}

func (o *ItemRequestDetailEngine) SaveMultiple(ctx *kaos.Context, payload *ItemRequestDetailSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ItemRequestID == "" {
		return nil, errors.New("missing: payload")
	}

	p := new(scmmodel.ItemRequest)
	if e := h.GetByID(p, payload.ItemRequestID); e != nil {
		return nil, errors.New("no item request data found: " + e.Error())
	}

	if e := h.DeleteByFilter(new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", payload.ItemRequestID)); e != nil {
		return nil, errors.New("error clear pr details: " + e.Error())
	}

	for _, dt := range payload.ItemRequestDetails {
		dt.ItemRequestID = payload.ItemRequestID
		if e := h.Save(&dt); e != nil {
			return nil, errors.New("error update Movement  Detail: " + e.Error())
		}
	}

	return payload, nil
}

type FulfillmentRequest struct {
	ItemRequestID string
	UserID        string
}

func (o *ItemRequestDetailEngine) Fulfillment(ctx *kaos.Context, h *datahub.Hub, payload *FulfillmentRequest) (interface{}, error) {
	var e error
	var referenceID string

	if payload == nil && payload.ItemRequestID == "" {
		return nil, errors.New("missing: payload")
	}

	//get Item Request By ID
	itemRequest, err := datahub.GetByID(h, new(scmmodel.ItemRequest), payload.ItemRequestID)
	if err != nil {
		return nil, errors.New("missing: item request")
	}

	//get item request detail
	itemDetails := []scmmodel.ItemRequestDetail{}
	e = h.GetsByFilter(new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", itemRequest.ID), &itemDetails)
	if e != nil {
		return nil, errors.New("missing: item details")
	}

	type TransferGroup struct {
		InventDimID   string
		InventDimFrom scmmodel.InventDimension
		Lines         []scmmodel.InventJournalLine
	}

	prLines := []scmmodel.PurchaseJournalLine{}
	movInLines := []scmmodel.InventJournalLine{}
	movOutLines := []scmmodel.InventJournalLine{}
	transGroup := []TransferGroup{}

	lo.ForEach(itemDetails, func(detail scmmodel.ItemRequestDetail, index int) {
		lo.ForEach(detail.DetailLines, func(detailLine scmmodel.ItemRequestDetailLine, indexLine int) {
			switch detailLine.FulfillmentType {
			case "Movement In", "Movement Out":
				line := scmmodel.InventJournalLine{
					LineNo:    1,
					ItemID:    detail.ItemID,
					SKU:       detail.SKU,
					Text:      detail.Description,
					Qty:       detailLine.QtyFulfilled,
					UnitID:    detailLine.UoM,
					Dimension: itemRequest.Dimension,
					InventDim: itemRequest.InventDimTo,
				}

				if detailLine.FulfillmentType == "Movement In" {
					movInLines = append(movInLines, line)
				} else if detailLine.FulfillmentType == "Movement Out" {
					movOutLines = append(movOutLines, line)
				}
			case "Item Transfer":
				line := scmmodel.InventJournalLine{
					LineNo:    1,
					ItemID:    detail.ItemID,
					SKU:       detail.SKU,
					Text:      detail.Description,
					Qty:       detailLine.QtyFulfilled,
					UnitID:    detailLine.UoM,
					Dimension: itemRequest.Dimension,
					InventDim: itemRequest.InventDimTo,
				}

				inventDimFrom := detailLine.InventDimFrom.Calc()
				inventDimID := inventDimFrom.InventDimID

				_, idx, found := lo.FindIndexOf(transGroup, func(d TransferGroup) bool { return d.InventDimID == inventDimID })
				if found {
					transGroup[idx].Lines = append(transGroup[idx].Lines, line)
				} else {
					transGroup = append(transGroup, TransferGroup{
						InventDimID:   inventDimID,
						InventDimFrom: *inventDimFrom,
						Lines:         []scmmodel.InventJournalLine{line},
					})
				}

			case "Purchase Request":
				line := scmmodel.PurchaseJournalLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo:       1,
						ItemID:       detail.ItemID,
						SKU:          detail.SKU,
						Text:         detail.Description,
						Qty:          detailLine.QtyFulfilled,
						UnitID:       detailLine.UoM,
						Dimension:    itemRequest.Dimension,
						InventDim:    itemRequest.InventDimTo,
						RemainingQty: detailLine.QtyFulfilled,
					},
				}

				prLines = append(prLines, line)
			case "Assembly":
				//not implemented yet
				// wo := struct{
				// 	Name string
				// Dimension tenantcoremodel.Dimension
				// InventDim scmmodel.InventDimension
				// Lines:     []scmmodel.InventReceiveIssueLine,
				// }{}
				// 	scmconfig.Config.EventHub().Publish("/v1/mfg/workorder/save", mfgmodel.Model{})
			}
		})
	})

	now := time.Now()

	if len(movInLines) > 0 {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		movement := &scmmodel.InventJournal{
			ID:        id,
			TrxType:   scmmodel.JournalMovementIn,
			Status:    ficomodel.JournalStatusDraft,
			Text:      itemRequest.Name,
			Dimension: itemRequest.Dimension,
			InventDim: itemRequest.InventDimTo,
			Lines:     movInLines,
			ReffNo:    []string{itemRequest.ID},
			TrxDate:   now,
		}

		referenceID = id
		if e = h.Save(movement); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	if len(movOutLines) > 0 {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		movement := &scmmodel.InventJournal{
			ID:        id,
			TrxType:   scmmodel.JournalMovementOut,
			Status:    ficomodel.JournalStatusDraft,
			Text:      itemRequest.Name,
			Dimension: itemRequest.Dimension,
			InventDim: itemRequest.InventDimTo,
			Lines:     movOutLines,
			TrxDate:   now,
			ReffNo:    []string{itemRequest.ID},
		}

		referenceID = id
		if e = h.Save(movement); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	for _, tg := range transGroup {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		transfer := &scmmodel.InventJournal{
			ID:          id,
			Text:        itemRequest.Name,
			Status:      ficomodel.JournalStatusDraft,
			Dimension:   itemRequest.Dimension,
			InventDim:   tg.InventDimFrom,
			InventDimTo: itemRequest.InventDimTo,
			Lines:       tg.Lines,
			TrxType:     scmmodel.JournalTransfer,
			ReffNo:      []string{itemRequest.ID},
			TrxDate:     now,
		}

		referenceID = id
		if e = h.Save(transfer); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	if len(prLines) > 0 {
		// Note: entah kenapa nats tidak bisa di trigger disini, ajax jadi loading lama tidak selesai2
		// purchase := &PurchaseRequestInsertParam{
		// 	Name:      itemRequest.Name,
		// 	Dimension: itemRequest.Dimension,
		// 	Location:  itemRequest.InventDimTo,
		// 	ReffNo:    []string{itemRequest.ID},
		// 	CompanyID: itemRequest.CompanyID,
		// 	Lines:     prLines,
		// }
		// PurchaseRequestInsert(purchase, itemRequest.CompanyID, payload.UserID)

		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "PurchaseRequest") // TODO: seharusnya pake MWPreAssignSequenceNo

		purchase := &scmmodel.PurchaseRequestJournal{
			ID:           id,
			Name:         itemRequest.Name,
			Status:       ficomodel.JournalStatusDraft,
			Dimension:    itemRequest.Dimension,
			Location:     itemRequest.InventDimTo,
			TrxDate:      now,
			DocumentDate: &now,
			PRDate:       &now,
			ExpectedDate: &now,
			ReffNo:       []string{itemRequest.ID},
			CompanyID:    itemRequest.CompanyID,
			Priority:     itemRequest.Priority,
			Lines:        prLines,
		}

		referenceID = id
		if e = h.Save(purchase); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	return referenceID, nil
}
