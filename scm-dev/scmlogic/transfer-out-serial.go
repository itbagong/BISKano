package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type TransferOutSerialEngine struct{}

type TransferOutSerialSaveMultipleRequest struct {
	TransferId         string
	TransferOutSerials []scmmodel.TransferOutSerial
}

func (o *TransferOutSerialEngine) SaveMultiple(ctx *kaos.Context, payload *TransferOutSerialSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.TransferId == "" {
		return nil, errors.New("missing: payload")
	}

	// trf := new(scmmodel.Transfer)
	// if e := h.GetByID(trf, payload.TransferId); e != nil {
	// 	return nil, e
	// }

	// clear data by transfer in id before save
	transferInID := payload.TransferId
	if e := h.DeleteByFilter(new(scmmodel.TransferOutSerial), dbflex.Eq("TransferID", transferInID)); e != nil {
		return nil, errors.New("error clear item serials: " + e.Error())
	}

	for _, itemSerial := range payload.TransferOutSerials {
		itemSerial.TransferID = payload.TransferId
		// itemSerial.InventoryDimension = trf.InventoryDimension
		if e := h.Save(&itemSerial); e != nil {
			return nil, errors.New("error saving Transfer Serial: " + e.Error())
		}

		itemSerialIn := new(scmmodel.TransferInSerial)
		itemSerialIn.TransferID = payload.TransferId
		itemSerialIn.ItemID = itemSerial.ItemID
		itemSerialIn.BatchID = itemSerial.BatchID
		itemSerialIn.SKU = itemSerial.SKU
		itemSerialIn.SerialNumberID = itemSerial.SKU
		// itemSerialIn.InventoryDimension = trf.InventoryDimension
		if e := h.Save(itemSerialIn); e != nil {
			return nil, errors.New("error saving Transfer In Serial: " + e.Error())
		}
	}

	return payload, nil
}
