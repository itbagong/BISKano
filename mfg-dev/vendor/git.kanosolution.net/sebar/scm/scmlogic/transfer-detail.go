package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type TransferDetailEngine struct{}

type TransferDetailMultipleSaveRequest struct {
	TransferID    string
	TransferItems []scmmodel.TransferDetail
}

func (o *TransferDetailEngine) SaveMultiple(ctx *kaos.Context, payload *TransferDetailMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.TransferID == "" {
		return nil, errors.New("missing: payload")
	}

	e := h.DeleteByFilter(new(scmmodel.Transfer), dbflex.Eq("TransferID", payload.TransferID))
	if e != nil {
		return nil, errors.New("error clear item details: " + e.Error())
	}

	for _, itemTransfer := range payload.TransferItems {
		itemTransfer.TransferID = payload.TransferID
		if e := h.GetByID(new(scmmodel.TransferDetail), itemTransfer.ID); e != nil {
			if e := h.Insert(&itemTransfer); e != nil {
				return nil, errors.New("error insert Transfer  Detail: " + e.Error())
			}
		} else {
			if e := h.Save(&itemTransfer); e != nil {
				return nil, errors.New("error update Transfer  Detail: " + e.Error())
			}
		}
	}

	return payload, nil
}
