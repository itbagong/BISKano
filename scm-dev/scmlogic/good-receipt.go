package scmlogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type GoodReceiptEngine struct{}

type GoodReceiptRequest struct {
	GoodReceipt scmmodel.GoodReceipt
	Items       []scmmodel.GoodReceiptDetail
}

func (o *GoodReceiptEngine) Process(ctx *kaos.Context, payload *GoodReceiptRequest) (interface{}, error) {
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

		detailTotalQty[keyItem] += detail.Qty
	}

	isReceivedValid := true
	grStatus := scmmodel.GoodReceiptStatusClosed
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
		itemDetails := []scmmodel.GoodReceiptDetail{}
		e := h.GetsByFilter(new(scmmodel.GoodReceiptDetail), dbflex.And(
			dbflex.Eq("GoodReceiptID", payload.GoodReceipt.ID),
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
			grStatus = scmmodel.GoodReceiptStatusPartialReceived
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
			ReferenceID:        string(payload.GoodReceipt.ID),
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: payload.GoodReceipt.InventoryDimension,
		}, MovementTypeIn, float64(totalReceived))

		if e != nil {
			return nil, errors.New("Error when calculate item balance " + e.Error())
		}

		//update qty item details
		itemDetails := []scmmodel.GoodReceiptDetail{}
		e = h.GetsByFilter(new(scmmodel.GoodReceiptDetail), dbflex.And(
			dbflex.Eq("GoodReceiptID", payload.GoodReceipt.ID),
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

			h.Insert(&scmmodel.GoodReceiptItemHistory{
				GoodReceiptID:      item.GoodReceiptID,
				ItemID:             item.ItemID,
				SKU:                item.SKU,
				Qty:                totalReceived,
				UoM:                item.UoM,
				InventoryDimension: item.InventoryDimension,
			})
		}
	}

	payload.GoodReceipt.Status = grStatus
	e := h.Save(&payload.GoodReceipt)
	if e != nil {
		return nil, errors.New("Error update good receipt in " + e.Error())
	}

	return payload, nil
}
