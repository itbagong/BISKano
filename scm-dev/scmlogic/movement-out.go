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
)

type MovementOutEngine struct{}

func (o *MovementOutEngine) Submit(ctx *kaos.Context, payload *scmmodel.MovementOut) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Status = scmmodel.MovementStatusSubmitted
	if e := h.Save(payload); e != nil {
		return nil, e
	}

	return payload, nil
}

type MovementOutApproveRequest struct {
	MovementOutID string `json:"MovementOutID"`
}

func (o *MovementOutEngine) Approve(ctx *kaos.Context, payload *MovementOutApproveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	movementID := payload.MovementOutID
	if movementID == "" {
		return nil, errors.New("missing: movement ID")
	}

	mov := new(scmmodel.MovementOut)
	if e := h.GetByID(mov, movementID); e != nil {
		return nil, e
	}

	if mov.Status != scmmodel.MovementStatusSubmitted {
		return nil, fmt.Errorf("status is not submitted (already '%s')", mov.Status)
	}

	mDetails := []scmmodel.MovementOutDetail{}
	if e := h.GetsByFilter(new(scmmodel.MovementOutDetail), dbflex.Eq("MovementOutID", movementID), &mDetails); e != nil {
		return nil, e
	}

	itemM := map[string]float64{}
	sep := "||"
	for _, d := range mDetails {
		idsku := fmt.Sprintf("%s%s%s", d.ItemID, sep, d.SKU)
		itemM[idsku] = itemM[idsku] + float64(d.Qty) // TODO: belum ada proses konversi unit (UoM)
	}

	for idsku, qty := range itemM {
		idskus := strings.Split(idsku, sep)
		if len(idskus) != 2 {
			return nil, fmt.Errorf("data missmatch item ID + sku: %s", idsku)
		}
		itemID := idskus[0]
		sku := idskus[1]

		param := ItemBalanceUniqueParam{
			ReferenceID:        mov.ID,
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: mov.InventoryDimension,
		}

		if e := new(ItemBalanceLogic).SetPlan(h, param, MovementTypeOut, qty); e != nil {
			return nil, e
		}
	}

	mov.Status = scmmodel.MovementStatusApproved
	e := h.Save(mov)
	if e != nil {
		return nil, errors.New("Failed update movement out " + e.Error())
	}

	// TODO: proses copy data ke good-issue
	// TODO: proses copy data ke good-receipt
	GoodIssue := new(scmmodel.GoodIssue)
	GoodIssue.GoodIssueDate = time.Now()
	GoodIssue.GoodIssueFrom = scmmodel.GoodIssueFromMovementOut
	GoodIssue.ReffNo = mov.ID
	GoodIssue.JournalType = mov.JournalType
	GoodIssue.Status = scmmodel.GoodIssueStatusOpen
	GoodIssue.FinancialDimension = mov.FinancialDimension
	GoodIssue.InventoryDimension = mov.InventoryDimension
	e = h.Save(GoodIssue)
	if e != nil {
		return nil, errors.New("Error insert good issue")
	}

	for _, detail := range mDetails {
		goodIssueDetail := new(scmmodel.GoodIssueDetail)
		goodIssueDetail.GoodIssueID = GoodIssue.ID
		goodIssueDetail.ItemID = detail.ItemID
		goodIssueDetail.SKU = detail.SKU
		goodIssueDetail.Description = detail.Description
		goodIssueDetail.Qty = detail.Qty
		goodIssueDetail.UoM = detail.UoM
		goodIssueDetail.Remarks = detail.Remarks
		goodIssueDetail.InventoryDimension = detail.InventoryDimension

		e := h.Save(goodIssueDetail)
		if e != nil {
			return nil, errors.New("Error insert good receipt detail")
		}
	}

	return "movement out approved", nil
}

type MovementOutRejectRequest struct {
	MovementOutID string `json:"MovementOutID"`
	Reason        string `json:"Reason"`
}

func (o *MovementOutEngine) Reject(ctx *kaos.Context, payload *MovementOutRejectRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	mov := new(scmmodel.MovementOut)
	if e := h.GetByID(mov, payload.MovementOutID); e != nil {
		return nil, e
	}

	mov.Status = scmmodel.MovementStatusRejected
	mov.ReasonReject = payload.Reason
	if e := h.Save(mov); e != nil {
		return nil, e
	}

	return "movement out rejected", nil
}

type MovementDetailGIRequest struct {
	MovementOut scmmodel.MovementOut
	Items       []scmmodel.MovementOutDetail
}

func (o *MovementInEngine) GoodIssued(ctx *kaos.Context, payload *MovementDetailGIRequest) (interface{}, error) {
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

	isReservedValid := true
	movementOutStatus := scmmodel.MovementStatusClosed
	var checkErrorMessage error

	for itemID, totalReserved := range detailTotalQty {
		//get item balance by list of item_id and sku
		keyItem := strings.Split(itemID, "|")
		if len(keyItem) != 2 {
			return nil, errors.New("Failed key of item " + itemID)
		}

		itemID := keyItem[0]
		sku := keyItem[1]

		//get item detail
		itemDetails := []scmmodel.MovementOutDetail{}
		e := h.GetsByFilter(new(scmmodel.MovementOutDetail), dbflex.And(
			dbflex.Eq("MovementOutID", payload.MovementOut.ID),
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

		if totalReserved > totalExistingQty {
			checkErrorMessage = errors.New("Total Received must be less than or equal to Reserved Balance")
			isReservedValid = false
			break
		}

		if totalReserved < totalExistingQty {
			// movementOutStatus = scmmodel.MovementStatusPartialIssued
		}
	}

	if checkErrorMessage != nil {
		return nil, checkErrorMessage
	} else if isReservedValid == false {
		return nil, errors.New("Reserved logic not valid because " + checkErrorMessage.Error())
	}

	for itemID, totalReserved := range detailTotalQty {
		keyItem := strings.Split(itemID, "|")
		if len(keyItem) != 2 {
			return nil, errors.New("Failed key of item " + itemID)
		}

		itemID := keyItem[0]
		sku := keyItem[1]

		e := new(ItemBalanceLogic).SetConfirm(h, ItemBalanceUniqueParam{
			ReferenceID:        string(payload.MovementOut.ID),
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: payload.MovementOut.InventoryDimension,
		}, MovementTypeOut, float64(totalReserved))

		if e != nil {
			return nil, errors.New("Error when calculate item balance " + e.Error())
		}

		//update qty item details
		itemDetails := []scmmodel.MovementOutDetail{}
		e = h.GetsByFilter(new(scmmodel.MovementOutDetail), dbflex.And(
			dbflex.Eq("MovementOutID", payload.MovementOut.ID),
			dbflex.Eq("ItemID", itemID),
			dbflex.Eq("SKU", sku),
		), &itemDetails)

		if e != nil {
			return nil, errors.New(fmt.Sprintf("Item detail not found %s %s: %s", itemID, sku, e.Error()))
		}

		for _, item := range itemDetails {
			item.Qty -= totalReserved
			e = h.Save(&item)
			if e != nil {
				return nil, fmt.Errorf("Failed update quantity item detail" + e.Error())
			}
		}
	}

	payload.MovementOut.Status = movementOutStatus
	e := h.Save(&payload.MovementOut)
	if e != nil {
		return nil, errors.New("Error update movement out " + e.Error())
	}

	return payload, nil
}
