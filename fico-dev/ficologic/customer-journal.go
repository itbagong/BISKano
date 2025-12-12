package ficologic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

func RegisterCustomerJournal(s *kaos.Service, dbMod, uiMod kaos.Mod) {
	routeName := "customerjournal"
	s.Group().SetMod(dbMod, uiMod).
		RegisterMWs(MWPreFilterCompanyID, rbaclogic.MWRbacFilterDim("", "jwt")).
		Apply(
			s.RegisterModel(new(ficomodel.CustomerJournal), routeName).AllowOnlyRoute(uiEndPoints...).DisableRoute("gridconfig", "formconfig"),
			s.RegisterModel(new(ficomodel.CustomerJournal), routeName).AllowOnlyRoute(deleteEndPoints...),
			s.RegisterModel(new(ficomodel.CustomerJournal), routeName).AllowOnlyRoute(readEndPoints...),
			s.RegisterModel(new(ficomodel.CustomerJournal), routeName).AllowOnlyRoute("insert", "update", "save").
				RegisterMWs(
					tenantcorelogic.MWPreAssignCustomSequenceNo("CustomerJournal"),
					MWPreCustomerJournalAssignDefault,
				),
		)

	s.Group().SetMod(uiMod).Apply(
		s.RegisterModel(new(ficomodel.CustomerJournalGrid), "customerjournal").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.CustomerJournalForm), "customerjournal").AllowOnlyRoute("formconfig"),
	)

}

func MWPreCustomerJournalAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*ficomodel.CustomerJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	if record.Status == "" {
		record.Status = ficomodel.JournalStatusDraft
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))
	return true, nil
}
