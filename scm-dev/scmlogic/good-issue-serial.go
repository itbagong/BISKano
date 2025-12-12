package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type GoodIssueSerialEngine struct{}

type GoodIssueSerialMultipleSaveRequest struct {
	GoodIssueID string
	Serials     []scmmodel.GoodIssueItemSerial
}

func (o *GoodIssueSerialEngine) SaveMultiple(ctx *kaos.Context, payload *GoodIssueSerialMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.GoodIssueID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	goodIssueID := payload.GoodIssueID
	e := h.DeleteByFilter(new(scmmodel.GoodIssueItemBatch), dbflex.Eq("GoodIssueID", goodIssueID))
	if e != nil {
		return nil, errors.New("error clear item serials: " + e.Error())
	}

	for _, itemSerial := range payload.Serials {
		itemSerial.GoodIssueID = payload.GoodIssueID
		if e := h.GetByID(new(scmmodel.GoodIssueItemSerial), itemSerial.ID); e != nil {
			if e := h.Insert(&itemSerial); e != nil {
				return nil, errors.New("error insert Good Issue Serial: " + e.Error())
			}
		} else {
			if e := h.Save(&itemSerial); e != nil {
				return nil, errors.New("error update Good Issue Serial: " + e.Error())
			}
		}
	}

	return payload, nil
}
