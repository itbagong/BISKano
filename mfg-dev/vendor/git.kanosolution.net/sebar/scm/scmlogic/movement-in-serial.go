package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type MovementInSerialEngine struct{}

type MovementSerialMultipleSaveRequest struct {
	MovementID      string
	MovementSerials []scmmodel.MovementInItemSerial
}

func (o *MovementInSerialEngine) SaveMultiple(ctx *kaos.Context, payload *MovementSerialMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.MovementID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	movementInID := payload.MovementID
	e := h.DeleteByFilter(new(scmmodel.MovementInItemSerial), dbflex.Eq("MovementInID", movementInID))
	if e != nil {
		return nil, errors.New("error clear item serials: " + e.Error())
	}

	for _, itemSerial := range payload.MovementSerials {
		itemSerial.MovementInID = payload.MovementID
		if e := h.GetByID(new(scmmodel.MovementInItemSerial), itemSerial.ID); e != nil {
			if e := h.Insert(&itemSerial); e != nil {
				return nil, errors.New("error insert Movement Serial: " + e.Error())
			}
		} else {
			if e := h.Save(&itemSerial); e != nil {
				return nil, errors.New("error update Movement Serial: " + e.Error())
			}
		}
	}

	return payload, nil
}
