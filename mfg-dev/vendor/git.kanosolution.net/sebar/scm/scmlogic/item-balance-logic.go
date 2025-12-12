package scmlogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
)

type ItemBalanceLogic struct{}

type MovementType string

const (
	MovementTypeIn  MovementType = "MovementIn"
	MovementTypeOut MovementType = "MovementOut"
)

type ItemBalanceUniqueParam struct {
	ReferenceID        string
	ItemID             string
	SKU                string
	InventoryDimension scmmodel.InventDimension
}

func (o *ItemBalanceLogic) SetPlan(h *datahub.Hub, param ItemBalanceUniqueParam, moveType MovementType, qty float64) error {
	// TODO: how to set lock in db collection?

	ib := new(scmmodel.ItemBalance)
	h.GetByFilter(ib, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
		ItemID:             param.ItemID,
		SKU:                param.SKU,
		InventoryDimension: param.InventoryDimension,
	}))

	ib.ItemID = param.ItemID
	ib.SKU = param.SKU
	ib.InventDim = param.InventoryDimension

	switch moveType {
	case MovementTypeIn:
		ib.QtyPlanned = ib.QtyPlanned + qty
	case MovementTypeOut:
		if qty > ib.QtyAvail {
			return fmt.Errorf("not enough qty (itemid: %s, available: %v, reserving: %v)", param.ItemID, ib.QtyAvail, qty)
		}

		ib.QtyAvail = ib.QtyAvail - qty
		ib.QtyReserved = ib.QtyReserved + qty
	}

	if e := h.Save(ib); e != nil {
		return e
	}

	trp := ItemTransactionUniqueParam{
		ReferenceID: param.ReferenceID,
		ItemID:      param.ItemID,
		SKU:         param.SKU,
	}

	if e := new(ItemTransactionLogic).Add(h, trp, moveType, scmmodel.ItemTransactionStatusPlanned, qty); e != nil {
		return e
	}

	return nil
}

func (o *ItemBalanceLogic) SetConfirm(h *datahub.Hub, param ItemBalanceUniqueParam, moveType MovementType, qty float64) error {
	// TODO: how to set lock in db collection?

	ib := new(scmmodel.ItemBalance)
	if e := h.GetByFilter(ib, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
		ItemID:             param.ItemID,
		SKU:                param.SKU,
		InventoryDimension: param.InventoryDimension,
	})); e != nil {
		return e
	}

	switch moveType {
	case MovementTypeIn:
		ib.QtyPlanned = ib.QtyPlanned - qty
		if ib.QtyPlanned < 0 {
			ib.QtyPlanned = 0
		}

		ib.Qty = ib.Qty + qty
		ib.QtyAvail = ib.QtyAvail + qty
	case MovementTypeOut:
		if qty > ib.QtyReserved {
			return fmt.Errorf("goods issuance is more than reserved (itemid: %s, reserved: %v, issuance: %v)", param.ItemID, ib.QtyReserved, qty)
		}

		ib.Qty = ib.Qty - qty
		ib.QtyReserved = ib.QtyReserved - qty
	}

	if e := h.Save(ib); e != nil {
		return e
	}

	trp := ItemTransactionUniqueParam{
		ReferenceID: param.ReferenceID,
		ItemID:      param.ItemID,
		SKU:         param.SKU,
	}

	if e := new(ItemTransactionLogic).Add(h, trp, moveType, scmmodel.ItemTransactionStatusConfirmed, qty); e != nil {
		return e
	}

	return nil
}

func (o *ItemBalanceLogic) GetLastDate(h *datahub.Hub, companyID string, balanceDate *time.Time) (*time.Time, error) {
	dateFilter := dbflex.And(dbflex.Eq("CompanyID", companyID), dbflex.Gte("BalanceDate", balanceDate))

	dateRecord, err := datahub.GetByParm(h, new(scmmodel.ItemBalance), dbflex.NewQueryParam().SetWhere(dateFilter).SetSort("-BalanceDate"))
	if err != nil {
		return nil, nil
	}

	return dateRecord.BalanceDate, nil
}
