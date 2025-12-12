package ficologic

import (
	"errors"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

func RegisterVendorJournal(s *kaos.Service, dbMod, uiMod kaos.Mod) {
	sequenceNo := "VendorJournal"
	routeName := strings.ToLower(sequenceNo)
	s.Group().SetMod(dbMod, uiMod).
		RegisterMWs(MWPreFilterCompanyID, rbaclogic.MWRbacFilterDim("", "jwt")).
		Apply(
			// s.RegisterModel(new(ficomodel.VendorJournal), routeName).AllowOnlyRoute(uiEndPoints...),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).AllowOnlyRoute(uiEndPoints...).DisableRoute("gridconfig", "formconfig"),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).AllowOnlyRoute(deleteEndPoints...),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).
				AllowOnlyRoute("gets").
				RegisterPostMWs(tenantcorelogic.MWPostVendorName()),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).
				AllowOnlyRoute("get").
				RegisterPostMWs(MWPostVendorJournalVoucherNo()),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).AllowOnlyRoute(readEndPoints...).DisableRoute("gets", "get"),
			s.RegisterModel(new(ficomodel.VendorJournal), routeName).AllowOnlyRoute("insert", "update", "save").
				RegisterMW(tenantcorelogic.MWPreAssignCustomSequenceNo("VendorJournal"), "preSeq").
				// RegisterMW(tenantcorelogic.MWPreAssignSequenceNo(sequenceNo, true, "_id"), "preSeq").
				RegisterMW(MWPreVendorJournalAssignDefault, "assignDefault"),
		)

	s.Group().SetMod(uiMod).Apply(
		s.RegisterModel(new(ficomodel.VendorJournalGrid), routeName).AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.VendorJournalForm), routeName).AllowOnlyRoute("formconfig"),
	)

}

func MWPreVendorJournalAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*ficomodel.VendorJournal)
	if !ok {
		return false, errors.New("invalid payload type")
	}
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	record.CompanyID = coID
	if record.Status == "" {
		record.Status = ficomodel.JournalStatusDraft
	}
	tenantcorelogic.SetIfEmpty(&record.CompanyID, tenantcorelogic.GetCompanyIDFromContext(ctx))

	// check GR Line already use or not
	// if len(record.ReffGRNumber) > 0 {
	if len(record.References) > 0 {
		listGR := []string{}
		for _, val := range record.References {
			if val.Key == "GR Ref No" {
				grNo := val.Value.(string)
				if grNo != "" {
					listGR = strings.Split(grNo, ",")
				}
			}
		}
		if len(listGR) > 0 {
			payload := struct {
				GRNumber []string
			}{
				GRNumber: listGR,
			}

			ev, _ := ctx.DefaultEvent()
			if ev == nil {
				return false, errors.New("nil: EventHub")
			}
			res := ""
			e := ev.Publish("/v1/scm/inventory/receive/check-gr-used", &payload, &res, nil)
			if e != nil {
				return false, errors.New(e.Error())
			}
		}
	}

	return true, nil
}

func MWPostVendorJournalVoucherNo() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)

		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(*ficomodel.VendorJournal)
		if !ok {
			return true, nil
		}

		cashSchedule := []ficomodel.CashSchedule{}
		e := h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.Eq("SourceJournalID", res.ID), &cashSchedule)
		if e != nil {
			ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
		}

		if len(cashSchedule) > 0 {
			res.LedgerVoucherNo = cashSchedule[0].VoucherNo
		}

		// res.Set("data", ms)
		ctx.Data().Set("FnResult", res)

		return true, nil
	}
}
