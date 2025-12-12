package ficologic

import (
	"errors"
	"strings"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPreLedgerJournalReadFromJournalType(ctx *kaos.Context, payload interface{}) (bool, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return false, errors.New("missing: dbconn")
	}

	o, ok := payload.(*ficomodel.LedgerJournal)
	if !ok {
		return false, errors.New("invalid payload: LedgerJournal")
	}

	if o.JournalTypeID == "" {
		return false, errors.New("missing: JournalType")
	}

	journalType, err := datahub.GetByID(h, new(ficomodel.LedgerJournalType), o.JournalTypeID)
	if err != nil {
		return false, ctx.Log().Error2("get journal type", "get journal type error: %s", err.Error())
	}

	// set default value if null
	o.PostingProfileID = lo.Ternary(o.PostingProfileID == "", journalType.PostingProfileID, o.PostingProfileID)
	o.ReferenceTemplateID = lo.Ternary(o.ReferenceTemplateID == "", journalType.ReferenceTemplateID, o.ReferenceTemplateID)
	o.ChecklistTemplateID = lo.Ternary(o.ChecklistTemplateID == "", journalType.ChecklistTemplateID, o.ChecklistTemplateID)

	return true, nil
}

func MWPreLedgerJournalAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*ficomodel.LedgerJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	if record.Status == "" {
		record.Status = ficomodel.JournalStatusDraft
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))
	return true, nil
}

func RegisterLedgerJournal(s *kaos.Service, dbMod, uiMod kaos.Mod) {
	s.Group().SetMod(uiMod).Apply(
		s.RegisterModel(new(ficomodel.LedgerJournalLineGrid), "ledgerjournal/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.LedgerJournalLineForm), "ledgerjournal/line").AllowOnlyRoute("formconfig"),
	)

	s.Group().SetMod(dbMod, uiMod).
		RegisterMWs(MWPreFilterCompanyID, rbaclogic.MWRbacFilterDim("", "jwt")).
		Apply(
			s.RegisterModel(new(ficomodel.LedgerJournal), "ledgerjournal").AllowOnlyRoute(uiEndPoints...),
			s.RegisterModel(new(ficomodel.LedgerJournal), "ledgerjournal").AllowOnlyRoute(deleteEndPoints...),
			s.RegisterModel(new(ficomodel.LedgerJournal), "ledgerjournal").
				AllowOnlyRoute(readEndPoints...).
				RegisterPostMWs(func(ctx *kaos.Context, payload interface{}) (bool, error) {
					routePath := ctx.Data().Get("RoutePath", "").(string)
					if strings.HasSuffix(routePath, "/get") {
						res, ok := ctx.Data().Get("FnResult", nil).(*ficomodel.LedgerJournal)
						if ok && len(res.Lines) > 0 {
							lineMs := make([]codekit.M, len(res.Lines))
							for idx, line := range res.Lines {
								lineM := codekit.M{}
								serde.Serde(line, &lineM)
								lineM["Debit"] = lo.Ternary(line.Amount > 0, line.Amount, 0)
								lineM["Credit"] = lo.Ternary(line.Amount < 0, -line.Amount, 0)
								lineMs[idx] = lineM
							}
							resM, _ := codekit.ToM(res)
							resM.Set("Lines", lineMs)
							ctx.Data().Set("FnResult", resM)
							return true, nil
						}
					}
					return true, nil
				}),
			s.RegisterModel(new(ficomodel.LedgerJournal), "ledgerjournal").
				AllowOnlyRoute("insert", "update", "save").
				RegisterMWs(
					tenantcorelogic.MWPreAssignCustomSequenceNo("LedgerJournal"),
					MWPreLedgerJournalAssignDefault,
					MWPreLedgerJournalReadFromJournalType),
		)
}
