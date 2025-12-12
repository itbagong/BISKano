package scmlogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type MovementInEngine struct{}

func (o *MovementInEngine) Draft(ctx *kaos.Context, payload *scmmodel.MovementIn) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Status = scmmodel.MovementStatusDraft
	e := o.upsert(h, payload)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

func (o *MovementInEngine) Submit(ctx *kaos.Context, payload *scmmodel.MovementIn) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Status = scmmodel.MovementStatusSubmitted
	e := o.upsert(h, payload)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

type MovementInApproveRequest struct {
	MovementInID string `json:"MovementInID"`
}

func (o *MovementInEngine) Approve(ctx *kaos.Context, payload *MovementInApproveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	//get movement in
	movementIn := new(scmmodel.MovementIn)
	e := h.GetByID(movementIn, payload.MovementInID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing movement in by ID: %s", payload.MovementInID))
	}

	if movementIn.Status != scmmodel.MovementStatusSubmitted {
		return nil, fmt.Errorf("status is not submitted (already '%s')", movementIn.Status)
	}

	if movementIn.Status == scmmodel.MovementStatusApproved {
		return payload, nil
	}

	//get item details
	movementItemDetails := []scmmodel.MovementInDetail{}
	e = h.GetsByFilter(new(scmmodel.MovementInDetail), dbflex.Eq("MovementInID", movementIn.ID), &movementItemDetails)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing item details with movement in ID: %s", movementIn.ID))
	}

	//prepare for item balances
	detailTotalQty := map[string]float64{}
	sep := "||"
	for _, detail := range movementItemDetails {
		idsku := fmt.Sprintf("%s%s%s", detail.ItemID, sep, detail.SKU)
		detailTotalQty[idsku] = detailTotalQty[idsku] + float64(detail.Qty) // TODO: belum ada proses konversi unit (UoM)
	}

	for idsku, itembalanceQty := range detailTotalQty {
		idskus := strings.Split(idsku, sep)
		if len(idskus) != 2 {
			return nil, fmt.Errorf("data missmatch item ID + sku: %s", idsku)
		}
		itemID := idskus[0]
		sku := idskus[1]

		param := ItemBalanceUniqueParam{
			ReferenceID:        movementIn.ID,
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: movementIn.InventoryDimension,
		}

		if e := new(ItemBalanceLogic).SetPlan(h, param, MovementTypeIn, itembalanceQty); e != nil {
			return nil, e
		}
	}

	//update status movement in
	movementIn.Status = scmmodel.MovementStatusApproved
	e = h.Save(movementIn)
	if e != nil {
		return nil, errors.New("Error update status movement in")
	}

	// TODO: proses copy data ke good-receipt
	goodReceipt := new(scmmodel.GoodReceipt)
	goodReceipt.GoodReceiptDate = time.Now()
	goodReceipt.GoodReceiptFrom = scmmodel.GoodReceiptFromMovementIn
	goodReceipt.ReffNo = movementIn.ID
	goodReceipt.JournalType = movementIn.JournalType
	goodReceipt.Status = scmmodel.GoodReceiptStatusOpen
	goodReceipt.FinancialDimension = movementIn.FinancialDimension
	goodReceipt.InventoryDimension = movementIn.InventoryDimension
	e = h.Save(goodReceipt)
	if e != nil {
		return nil, errors.New("Error insert good receipt")
	}

	for _, detail := range movementItemDetails {
		goodReceiptDetail := new(scmmodel.GoodReceiptDetail)
		goodReceiptDetail.GoodReceiptID = goodReceipt.ID
		goodReceiptDetail.ItemID = detail.ItemID
		goodReceiptDetail.SKU = detail.SKU
		goodReceiptDetail.Description = detail.Description
		goodReceiptDetail.Qty = detail.Qty
		goodReceiptDetail.UoM = detail.UoM
		goodReceiptDetail.Remarks = detail.Remarks
		goodReceiptDetail.InventoryDimension = detail.InventoryDimension

		e := h.Save(goodReceiptDetail)
		if e != nil {
			return nil, errors.New("Error insert good receipt detail")
		}
	}
	return payload, nil
}

type MovementInRejectRequest struct {
	MovementInID string `json:"MovementInID"`
	Reason       string `json:"Reason"`
}

func (o *MovementInEngine) Reject(ctx *kaos.Context, payload *MovementInRejectRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	//get movement in
	movementIn := new(scmmodel.MovementIn)
	e := h.GetByID(movementIn, payload.MovementInID)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Missing movement in by ID: %s", payload.MovementInID))
	}

	movementIn.Status = scmmodel.MovementStatusRejected
	movementIn.ReasonReject = payload.Reason
	e = o.upsert(h, movementIn)
	if e != nil {
		return nil, e
	}

	return payload, nil
}

type MovementDetailGRRequest struct {
	MovementIn scmmodel.MovementIn
	Items      []scmmodel.MovementInDetail
}

func (o *MovementInEngine) GoodReceipt(ctx *kaos.Context, payload *MovementDetailGRRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	detailTotalQty := map[string]int{}
	for _, detail := range payload.Items {
		keyItem := fmt.Sprintf("%s|%s", detail.ItemID, detail.SKU)
		if _, ok := detailTotalQty[keyItem]; !ok {
			detailTotalQty[keyItem] = 0
		}

		detailTotalQty[keyItem] += detail.Qty // TODO: belum ada proses konversi unit (UoM)
	}

	isReceivedValid := true
	movementInStatus := scmmodel.MovementStatusClosed
	var checkErrorMessage error

	for itemID, totalQty := range detailTotalQty {
		//get item balance by list of item_id and sku
		keyItem := strings.Split(itemID, "|")
		if len(keyItem) != 2 {
			return nil, errors.New("Failed key of item " + itemID)
		}

		itemID := keyItem[0]
		sku := keyItem[1]

		//get item detail
		itemDetails := []scmmodel.MovementInDetail{}
		e := h.GetsByFilter(new(scmmodel.MovementInDetail), dbflex.And(
			dbflex.Eq("MovementInID", payload.MovementIn.ID),
			dbflex.Eq("ItemID", itemID),
			dbflex.Eq("SKU", sku),
		), &itemDetails)

		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item detail not found %s %s: %s", itemID, sku, e.Error()))
		}

		totalExistingQty := 0
		for _, item := range itemDetails {
			totalExistingQty += item.Qty
		}

		if totalQty > totalExistingQty {
			return nil, errors.New(fmt.Sprintf("item with ID %s SKU %s are only allowed to enter the maximum quantity %d", itemID, sku, totalExistingQty))
		} else if totalQty < totalExistingQty {
			// movementInStatus = scmmodel.MovementStatusPartialReceived
		}
	}

	if checkErrorMessage != nil {
		return nil, checkErrorMessage
	} else if isReceivedValid == false {
		return nil, errors.New("Received logic not valid because " + checkErrorMessage.Error())
	}

	for itemID, totalReceived := range detailTotalQty {
		keyItem := strings.Split(itemID, "|")
		if len(keyItem) != 2 {
			return nil, errors.New("Failed key of item " + itemID)
		}

		itemID := keyItem[0]
		sku := keyItem[1]

		e := new(ItemBalanceLogic).SetConfirm(h, ItemBalanceUniqueParam{
			ReferenceID:        string(payload.MovementIn.ID),
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: payload.MovementIn.InventoryDimension,
		}, MovementTypeIn, float64(totalReceived))

		if e != nil {
			return nil, errors.New("Error when calculate item balance " + e.Error())
		}

		//update qty item details
		itemDetails := []scmmodel.MovementInDetail{}
		e = h.GetsByFilter(new(scmmodel.MovementInDetail), dbflex.And(
			dbflex.Eq("MovementInID", payload.MovementIn.ID),
			dbflex.Eq("ItemID", itemID),
			dbflex.Eq("SKU", sku),
		), &itemDetails)

		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item detail not found %s %s: %s", itemID, sku, e.Error()))
		}

		for _, item := range itemDetails {
			item.Qty -= totalReceived
			e = h.Save(&item)
			if e != nil {
				return nil, fmt.Errorf("Failed update quantity item detail" + e.Error())
			}
		}
	}

	payload.MovementIn.Status = movementInStatus
	e := h.Save(&payload.MovementIn)
	if e != nil {
		return nil, errors.New("Error update movement in " + e.Error())
	}

	return payload, nil
}

func (o *MovementInEngine) upsert(h *datahub.Hub, model *scmmodel.MovementIn) error {
	if e := h.GetByID(new(scmmodel.MovementIn), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Movement In: " + e.Error())
		}
	} else {
		if e := h.Save(model); e != nil {
			return errors.New("error update Movement In: " + e.Error())
		}
	}

	return nil
}
