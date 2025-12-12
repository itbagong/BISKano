package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type TransferInSerialEngine struct{}

type TransferInSerialSaveMultipleRequest struct {
	TransferId        string
	TransferInSerials []scmmodel.TransferInSerial
}

func (o *TransferInSerialEngine) SaveMultiple(ctx *kaos.Context, payload *TransferInSerialSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.TransferId == "" {
		return nil, errors.New("missing: payload")
	}

	trf := new(scmmodel.TransferIn)
	if e := h.GetByID(trf, payload.TransferId); e != nil {
		return nil, e
	}

	// clear data by transfer in id before save
	transferInID := payload.TransferId
	if e := h.DeleteByFilter(new(scmmodel.TransferInSerial), dbflex.Eq("TransferID", transferInID)); e != nil {
		return nil, errors.New("error clear item serials: " + e.Error())
	}

	for _, itemSerial := range payload.TransferInSerials {
		itemSerial.TransferID = payload.TransferId
		itemSerial.InventoryDimension = trf.InventoryDimension
		if e := h.Save(&itemSerial); e != nil {
			return nil, errors.New("error saving Transfer Serial: " + e.Error())
		}
	}

	return payload, nil
}
