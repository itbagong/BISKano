package scmlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type TransferEngine struct{}

func (o *TransferEngine) Draft(ctx *kaos.Context, payload *scmmodel.Transfer) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Status = scmmodel.TransferStatusDraft
	e := o.upsert(h, payload)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

func (o *TransferEngine) Submit(ctx *kaos.Context, payload *scmmodel.Transfer) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Status = scmmodel.TransferStatusSubmitted
	e := o.upsert(h, payload)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

type TransferApproveRequest struct {
	TransferID string `json:"TransferID"`
}

func (o *TransferEngine) Approval(ctx *kaos.Context, payload *TransferApproveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	//get transfer
	transfer := new(scmmodel.Transfer)
	e := h.GetByID(transfer, payload.TransferID)
	if e != nil {
		return nil, errors.New("Transfer not found " + e.Error())
	}

	//get transfer detail
	transferDetails := []scmmodel.TransferDetail{}
	e = h.GetsByFilter(new(scmmodel.TransferDetail), dbflex.In("TransferID", transfer.ID), &transferDetails)
	if e != nil {
		return nil, errors.New("Failed get transfer details " + e.Error())
	}

	//copy to GR and GI
	GoodIssue := new(scmmodel.GoodIssue)
	GoodIssue.GoodIssueDate = time.Now()
	GoodIssue.GoodIssueFrom = scmmodel.GoodIssueFromItemTransfer
	GoodIssue.ReffNo = transfer.ID
	GoodIssue.JournalType = transfer.JournalType
	GoodIssue.Status = scmmodel.GoodIssueStatusOpen
	GoodIssue.FinancialDimension = transfer.FinancialDimensionFrom
	GoodIssue.InventoryDimension = transfer.InventoryDimensionFrom
	e = h.Save(GoodIssue)
	if e != nil {
		return nil, errors.New("Error insert good issue")
	}

	GoodReceipt := new(scmmodel.GoodReceipt)
	GoodReceipt.GoodReceiptDate = time.Now()
	GoodReceipt.GoodReceiptFrom = scmmodel.GoodReceiptFromItemTransfer
	GoodReceipt.ReffNo = transfer.ID
	GoodReceipt.JournalType = transfer.JournalType
	GoodReceipt.Status = scmmodel.GoodReceiptStatusOpen
	GoodReceipt.FinancialDimension = transfer.FinancialDimensionTo
	GoodReceipt.InventoryDimension = transfer.InventoryDimensionTo
	e = h.Save(GoodReceipt)
	if e != nil {
		return nil, errors.New("Error insert good receipt")
	}

	for _, item := range transferDetails {
		e = new(ItemBalanceLogic).SetPlan(h, ItemBalanceUniqueParam{
			ReferenceID:        item.TransferID,
			ItemID:             item.ItemID,
			SKU:                item.SKU,
			InventoryDimension: transfer.InventoryDimensionFrom,
		}, MovementTypeOut, float64(item.Qty))
		if e != nil {
			return nil, e
		}

		//move item to good issue detail
		goodIssueDetail := new(scmmodel.GoodIssueDetail)
		goodIssueDetail.GoodIssueID = GoodIssue.ID
		goodIssueDetail.ItemID = item.ItemID
		goodIssueDetail.SKU = item.SKU
		goodIssueDetail.Description = item.Description
		goodIssueDetail.Qty = item.Qty
		goodIssueDetail.UoM = item.UoM
		goodIssueDetail.Remarks = item.Remarks
		goodIssueDetail.InventoryDimension = transfer.InventoryDimensionFrom
		e = h.Save(goodIssueDetail)
		if e != nil {
			return nil, errors.New("Failed save transfer out " + e.Error())
		}

		//move item to good issue detail
		goodReceiptDetail := new(scmmodel.GoodReceiptDetail)
		goodReceiptDetail.GoodReceiptID = GoodReceipt.ID
		goodReceiptDetail.ItemID = item.ItemID
		goodReceiptDetail.SKU = item.SKU
		goodReceiptDetail.Description = item.Description
		goodReceiptDetail.Qty = item.Qty
		goodReceiptDetail.UoM = item.UoM
		goodReceiptDetail.Remarks = item.Remarks
		goodReceiptDetail.InventoryDimension = transfer.InventoryDimensionFrom
		e = h.Save(goodReceiptDetail)
		if e != nil {
			return nil, errors.New("Failed save transfer out " + e.Error())
		}
	}

	transfer.Status = scmmodel.TransferStatusApproved
	e = h.Save(transfer)
	if e != nil {
		return nil, errors.New("Failed update transfer " + e.Error())
	}

	return payload, nil
}

type TransferRejectRequest struct {
	TransferID string `json:"TransferID"`
	Reason     string `json:"Reason"`
}

func (o *TransferEngine) Reject(ctx *kaos.Context, payload *TransferRejectRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	//get movement in
	transfer := new(scmmodel.Transfer)
	e := h.GetByID(transfer, payload.TransferID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing transfer in by ID: %s", payload.TransferID))
	}

	transfer.Status = scmmodel.TransferStatusRejected
	transfer.ReasonReject = payload.Reason
	e = o.upsert(h, transfer)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

type TransferGIRequest struct {
	TransferID string
	Items      []scmmodel.TransferOut
}

func (o TransferEngine) GoodIssued(ctx *kaos.Context, payload *TransferGIRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	isReservedValid := true
	transferOutStatus := scmmodel.TransferStatusIssued
	var checkErrorMessage error

	transfer := new(scmmodel.Transfer)
	e := h.GetByID(transfer, payload.TransferID)
	if e != nil {
		return nil, errors.New("Transfer not found " + e.Error())
	}

	itemQty := map[string]int{}
	for _, item := range payload.Items {
		//get item detail
		if _, ok := itemQty[item.ID]; !ok {
			itemQty[item.ID] = 0
		}

		itemQty[item.ID] = item.Qty

		itemExisting := new(scmmodel.TransferOut)
		e := h.GetByID(itemExisting, item.ID)
		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item out not found %s %s: %s", item.ID, item.SKU, e.Error()))
		}

		if item.Qty > itemExisting.Qty {
			checkErrorMessage = errors.New("Total Received must be less than or equal to Reserved Balance")
			isReservedValid = false
			break
		}

		if itemExisting.Qty == 0 {
			checkErrorMessage = errors.New("Reserved Balance is empty")
			isReservedValid = false
			break
		}

		if item.Qty < itemExisting.Qty {
			if transfer.Status != scmmodel.TransferStatusPartialReceived {
				transferOutStatus = scmmodel.TransferStatusPartialIssued
			} else {
				transferOutStatus = scmmodel.TransferStatusPartialReceived
			}
		}
	}

	if checkErrorMessage != nil {
		return nil, checkErrorMessage
	} else if isReservedValid == false {
		return nil, errors.New("Reserved logic not valid because " + checkErrorMessage.Error())
	}

	for _, item := range payload.Items {
		e := new(ItemBalanceLogic).SetConfirm(h, ItemBalanceUniqueParam{
			ReferenceID:        string(payload.TransferID),
			ItemID:             item.ItemID,
			SKU:                item.SKU,
			InventoryDimension: transfer.InventoryDimensionFrom,
		}, MovementTypeOut, float64(item.Qty))

		if e != nil {
			return nil, e
		}

		//update qty item details
		itemExisting := new(scmmodel.TransferOut)
		e = h.GetByID(itemExisting, item.ID)
		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item out not found %s %s: %s", item.ID, item.SKU, e.Error()))
		}

		totalReserved := 0
		if qty, ok := itemQty[item.ID]; ok {
			totalReserved = qty
		}

		if totalReserved < itemExisting.Qty {
			itemExisting.Status = scmmodel.TransferStatusPartialIssued
		} else {
			itemExisting.Status = scmmodel.TransferStatusClosed
		}

		itemExisting.Qty -= totalReserved
		e = h.Save(itemExisting)
		if e != nil {
			return nil, fmt.Errorf("Failed update quantity item out" + e.Error())
		}

		//move item to transfer in
		transferIn := new(scmmodel.TransferIn)
		transferIn.TransferID = item.TransferID
		transferIn.TransferOutID = item.ID
		transferIn.ItemID = item.ItemID
		transferIn.SKU = item.SKU
		transferIn.Description = item.Description
		transferIn.Qty = totalReserved
		transferIn.UoM = item.UoM
		transferIn.Remarks = item.Remarks
		transferIn.Status = scmmodel.TransferStatusApproved
		transferIn.FinancialDimension = transfer.FinancialDimensionTo
		transferIn.InventoryDimension = item.InventoryDimension
		e = h.Save(transferIn)
		if e != nil {
			return nil, errors.New("Failed save transfer in " + e.Error())
		}

		e = new(ItemBalanceLogic).SetPlan(h, ItemBalanceUniqueParam{
			ReferenceID:        item.TransferID,
			ItemID:             item.ItemID,
			SKU:                item.SKU,
			InventoryDimension: transfer.InventoryDimensionTo,
		}, MovementTypeIn, float64(totalReserved))
	}

	transfer.Status = transferOutStatus
	e = h.Save(transfer)
	if e != nil {
		return nil, errors.New("Error update transfer out " + e.Error())
	}

	return payload, nil
}

type TransferGRRequest struct {
	TransferID string
	Items      []scmmodel.TransferIn
}

func (o TransferEngine) GoodReceipt(ctx *kaos.Context, payload *TransferGRRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	isReceivedValid := true
	transferInStatus := scmmodel.TransferStatusClosed
	var checkErrorMessage error

	transfer := new(scmmodel.Transfer)
	e := h.GetByID(transfer, payload.TransferID)
	if e != nil {
		return nil, errors.New("Transfer not found " + e.Error())
	}

	itemQty := map[string]int{}
	for _, item := range payload.Items {
		//get item detail
		if _, ok := itemQty[item.ID]; !ok {
			itemQty[item.ID] = 0
		}

		itemQty[item.ID] = item.Qty

		itemExisting := new(scmmodel.TransferIn)
		e := h.GetByID(itemExisting, item.ID)

		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item in not found %s %s: %s", item.ID, item.SKU, e.Error()))
		}

		if item.Qty > itemExisting.Qty {
			checkErrorMessage = errors.New("Total Received must be less than or equal to Reserved Balance")
			isReceivedValid = false
			break
		}

		// if itemExisting.Qty == 0 {
		// 	// TODO: salah kasih handler, kalo disini, ketika ada 1 aja receive yang nol, bakal error padahal lainnya belum nol
		// 	checkErrorMessage = errors.New("This item has been received entirely")
		// 	isReceivedValid = false
		// 	break
		// }

		if item.Qty < itemExisting.Qty {
			transferInStatus = scmmodel.TransferStatusPartialReceived
		}
	}

	if checkErrorMessage != nil {
		return nil, checkErrorMessage
	} else if isReceivedValid == false {
		return nil, errors.New("Received logic not valid because " + checkErrorMessage.Error())
	}

	for _, item := range payload.Items {
		e := new(ItemBalanceLogic).SetConfirm(h, ItemBalanceUniqueParam{
			ReferenceID:        string(payload.TransferID),
			ItemID:             item.ItemID,
			SKU:                item.SKU,
			InventoryDimension: transfer.InventoryDimensionTo,
		}, MovementTypeIn, float64(item.Qty))

		if e != nil {
			return nil, errors.New("Error when calculate item balance " + e.Error())
		}

		//update qty item details
		// itemExisting := new(scmmodel.TransferIn)
		// e = h.GetByID(itemExisting, item.ID)
		// if e != nil {
		// 	return nil, errors.New(fmt.Sprintf("Item in not found %s %s: %s", item.ID, item.SKU, e.Error()))
		// }

		itemExisting := new(scmmodel.TransferIn)
		e = h.GetByID(itemExisting, item.ID)

		totalReserved := 0
		if qty, ok := itemQty[item.ID]; ok {
			totalReserved = qty
		}

		if totalReserved < itemExisting.Qty {
			itemExisting.Status = scmmodel.TransferStatusPartialReceived
		} else {
			itemExisting.Status = scmmodel.TransferStatusClosed
		}

		itemExisting.Qty -= totalReserved
		e = h.Save(itemExisting)
		if e != nil {
			return nil, fmt.Errorf("Failed update quantity item in" + e.Error())
		}
	}

	if transferInStatus == scmmodel.TransferStatusClosed {
		// check if its really closed
		dispOuts := []scmmodel.TransferOut{}
		h.Gets(new(scmmodel.TransferOut), dbflex.NewQueryParam().SetWhere(dbflex.Eq("TransferID", transfer.ID)), &dispOuts)

		for _, out := range dispOuts {
			if out.Qty > 0 {
				transferInStatus = scmmodel.TransferStatusPartialReceived
			}
		}
	}

	transfer.Status = transferInStatus
	e = h.Save(transfer)
	if e != nil {
		return nil, errors.New("Error update transfer in " + e.Error())
	}

	return payload, nil
}

func (o *TransferEngine) upsert(h *datahub.Hub, model *scmmodel.Transfer) error {
	if e := h.GetByID(new(scmmodel.Transfer), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Transfer: " + e.Error())
		}
	} else {
		if e := h.Save(model); e != nil {
			return errors.New("error update Transfer: " + e.Error())
		}
	}

	return nil
}
