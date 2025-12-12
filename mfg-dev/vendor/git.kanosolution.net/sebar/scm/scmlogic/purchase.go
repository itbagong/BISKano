package scmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/reflector"
)

type PurchaseEngine struct{}

type UpdatePurchaseRequest struct {
	SourceJournalID string
	SourceType      string
}

func (o *PurchaseEngine) UpdatePrint(ctx *kaos.Context, payload *UpdatePurchaseRequest) (interface{}, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}
	var m orm.DataModel

	switch payload.SourceType {
	case string(scmmodel.PurchOrder):
		m = new(scmmodel.PurchaseOrderJournal)
	case string(scmmodel.PurchRequest):
		m = new(scmmodel.PurchaseRequestJournal)
	default:
		return nil, fmt.Errorf("invalid module: %s", payload.SourceType)
	}

	if e := db.GetByID(m, payload.SourceJournalID); e != nil {
		return nil, e
	}

	jf := reflector.From(m)
	totalPrint, _ := jf.Get("TotalPrint")

	jf.Set("TotalPrint", totalPrint.(int)+1)
	if e := db.Save(m); e != nil {
		return nil, errors.New("error update purchase: " + e.Error())
	}

	return "OK", nil
}
