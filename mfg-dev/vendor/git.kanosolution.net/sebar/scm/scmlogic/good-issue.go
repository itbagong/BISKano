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

type GoodIssueEngine struct{}

type GoodIssueRequest struct {
	GoodIssue scmmodel.GoodIssue
	Items     []scmmodel.GoodIssueDetail
}

func (o *GoodIssueEngine) Process(ctx *kaos.Context, payload *GoodIssueRequest) (interface{}, error) {
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

	isReservedValid := true
	giStatus := scmmodel.GoodIssueStatusClosed
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
		itemDetails := []scmmodel.GoodIssueDetail{}
		e := h.GetsByFilter(new(scmmodel.GoodIssueDetail), dbflex.And(
			dbflex.Eq("GoodIssueID", payload.GoodIssue.ID),
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
			giStatus = scmmodel.GoodIssueStatusPartialIssued
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
			ReferenceID:        string(payload.GoodIssue.ID),
			ItemID:             itemID,
			SKU:                sku,
			InventoryDimension: payload.GoodIssue.InventoryDimension,
		}, MovementTypeOut, float64(totalReserved))

		if e != nil {
			return nil, errors.New("Error when calculate item balance " + e.Error())
		}

		//update qty item details
		itemDetails := []scmmodel.GoodIssueDetail{}
		e = h.GetsByFilter(new(scmmodel.GoodIssueDetail), dbflex.And(
			dbflex.Eq("GoodIssueID", payload.GoodIssue.ID),
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

			h.Insert(&scmmodel.GoodIssueItemHistory{
				GoodIssueID:        item.GoodIssueID,
				ItemID:             item.ItemID,
				SKU:                item.SKU,
				Qty:                totalReserved,
				UoM:                item.UoM,
				InventoryDimension: item.InventoryDimension,
			})
		}
	}

	payload.GoodIssue.Status = giStatus
	e := h.Save(&payload.GoodIssue)
	if e != nil {
		return nil, errors.New("Error update good issue " + e.Error())
	}

	return payload, nil
}
