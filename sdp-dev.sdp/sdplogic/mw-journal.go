package sdplogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

func MWPreJournal(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*scmmodel.InventJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	if record.Status == "" {
		record.Status = ficomodel.JournalStatus(scmmodel.JournalDraft)
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))
	return true, nil
}

func MWPreReceieveIssueJournal(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*scmmodel.InventReceiveIssueJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	if record.Status == "" {
		record.Status = ficomodel.JournalStatus(scmmodel.JournalDraft)
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))
	return true, nil
}
