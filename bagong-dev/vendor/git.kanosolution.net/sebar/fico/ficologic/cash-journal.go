package ficologic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type CashJournalHandler struct {
}

type CashJournalPayload struct {
	IDs        []string
	References tenantcoremodel.References
}

func (obj *CashJournalHandler) UpdateReferences(ctx *kaos.Context, cj *CashJournalPayload) ([]ficomodel.CashJournal, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if cj == nil {
		return nil, errors.New("missing: payload")
	} else if len(cj.IDs) == 0 {
		return nil, errors.New("id is required")
	}

	// gets data cash journal by id
	cashJournals := []ficomodel.CashJournal{}
	err := h.Gets(new(ficomodel.CashJournal), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", cj.IDs...),
		),
	), &cashJournals)
	if err != nil {
		return nil, fmt.Errorf("err when get cash journal: %s", err.Error())
	}

	if len(cashJournals) > 0 {
		for _, val := range cashJournals {
			if len(cj.References) > 0 {
				val.References = append(val.References, cj.References[0])
			} else {
				if len(val.References) > 0 {
					for iref, valref := range val.References {
						if valref.Key == "SubmissionJournalID" {
							val.References = append(val.References[:iref], val.References[iref+1:]...)
							break
						}
					}
				}
			}
			err = h.Save(&val)
			if err != nil {
				return nil, errors.New("Failed update data cash journal: " + err.Error())
			}
		}
	}

	return cashJournals, nil
}

func RegisterCashJournal(s *kaos.Service, dbMod, uiMod kaos.Mod) {
	routeName := "cashjournal"
	s.Group().SetMod(dbMod, uiMod).
		RegisterMWs(MWPreFilterCompanyID, rbaclogic.MWRbacFilterDim("", "jwt")).
		Apply(
			s.RegisterModel(new(ficomodel.CashJournal), routeName).AllowOnlyRoute(uiEndPoints...),
			s.RegisterModel(new(ficomodel.CashJournal), routeName).AllowOnlyRoute(deleteEndPoints...),
			s.RegisterModel(new(ficomodel.CashJournal), routeName).AllowOnlyRoute(readEndPoints...),
			s.RegisterModel(new(ficomodel.CashJournal), routeName).AllowOnlyRoute("insert", "update", "save").
				RegisterMWs(
					tenantcorelogic.MWPreAssignCustomSequenceNo("CashJournal"),
					MWPreCashJournalAssignDefault,
					// set created by
					func(ctx *kaos.Context, i interface{}) (bool, error) {
						record, ok := i.(*ficomodel.CashJournal)
						if !ok {
							return false, errors.New("invalid payload type")
						}
						if record.CreatedBy == "" {
							user := sebar.GetUserIDFromCtx(ctx)
							record.CreatedBy = user
						}
						return true, nil
					},
				),
		)
	s.RegisterModel(new(CashJournalHandler), routeName)
}

func MWPreCashJournalAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*ficomodel.CashJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	if record.Status == "" {
		record.Status = ficomodel.JournalStatusDraft
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))
	return true, nil
}
