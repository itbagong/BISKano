package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type GoodReceiptSerialEngine struct{}

type GoodReceiptSerialMultipleSaveRequest struct {
	GoodReceiptID string
	Serials       []scmmodel.GoodReceiptItemSerial
}

func (o *GoodReceiptSerialEngine) SaveMultiple(ctx *kaos.Context, payload *GoodReceiptSerialMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.GoodReceiptID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	goodReceiptID := payload.GoodReceiptID
	e := h.DeleteByFilter(new(scmmodel.GoodReceiptItemBatch), dbflex.Eq("GoodReceiptID", goodReceiptID))
	if e != nil {
		return nil, errors.New("error clear item serials: " + e.Error())
	}

	for _, itemSerial := range payload.Serials {
		itemSerial.GoodReceiptID = payload.GoodReceiptID
		if e := h.GetByID(new(scmmodel.GoodReceiptItemSerial), itemSerial.ID); e != nil {
			if e := h.Insert(&itemSerial); e != nil {
				return nil, errors.New("error insert Good Receipt Serial: " + e.Error())
			}
		} else {
			if e := h.Save(&itemSerial); e != nil {
				return nil, errors.New("error update Good Receipt Serial: " + e.Error())
			}
		}
	}

	return payload, nil
}
