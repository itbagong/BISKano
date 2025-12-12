package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/kanoteknologi/hd"
)

type FiscalYearHandler struct {
}

func (obj *FiscalYearHandler) Activate(ctx *kaos.Context, fy *ficomodel.FiscalYear) (*ficomodel.FiscalYear, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// ndak boleh ada period yang masih buka
	per, err := datahub.GetByFilter(h, new(ficomodel.FiscalPeriod), dbflex.And(dbflex.Eq("CompanyID", fy.CompanyID), dbflex.Ne("Modules.Status", ficomodel.PeriodOpen)))
	if err == nil {
		return nil, fmt.Errorf("no module allowed to be opened: %s", per.Name)
	}

	// deactivate all active year
	h.UpdateField(&ficomodel.FiscalYear{IsActive: false}, dbflex.Eqs("CompanyID", fy.CompanyID, "IsActive", true), "IsActive")

	// activate
	fy.IsActive = true
	h.Save(fy)

	return fy, nil
}

func createFiscalYear(h *datahub.Hub, fiscalYr *ficomodel.FiscalYear, moduleNames ...string) (*ficomodel.FiscalYear, error) {
	co, err := datahub.GetByID(h, new(tenantcoremodel.Company), fiscalYr.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("invalid: CompanyID: %s", fiscalYr.CompanyID)
	}

	err = h.Save(fiscalYr)
	if err != nil {
		return fiscalYr, fmt.Errorf("create fiscal year: %s", err.Error())
	}

	periods := []ficomodel.FiscalPeriod{}
	dtFrom := fiscalYr.From
	for {
		if dtFrom.After(fiscalYr.To) {
			break
		}
		dtEnd := EoMoDateOnly(dtFrom)

		period := ficomodel.FiscalPeriod{
			Name:         fmt.Sprintf("%s %s %s", co.Name, fiscalYr.Name, dtFrom.Format("Jan-2006")),
			FromDate:     dtFrom,
			ToDate:       dtEnd,
			FiscalYearID: fiscalYr.ID,
			CompanyID:    fiscalYr.CompanyID,
			Modules: map[string]ficomodel.PeriodStatus{
				"Finance":   ficomodel.PeriodBlocked,
				"Inventory": ficomodel.PeriodBlocked,
			},
		}
		period.Modules = make(map[string]ficomodel.PeriodStatus, len(moduleNames))
		for _, name := range moduleNames {
			period.Modules[name] = ficomodel.PeriodBlocked
		}
		periods = append(periods, period)

		dtFrom = dtEnd.AddDate(0, 0, 1)
	}

	for _, period := range periods {
		h.Insert(&period)
	}

	return fiscalYr, nil
}

func GetPeriodStatus(h *datahub.Hub, coID string, fiscalYearID string, date time.Time, module string) ficomodel.PeriodStatus {
	if fiscalYearID == "" {
		yr, err := datahub.GetByFilter(h, new(ficomodel.FiscalYear), dbflex.Eqs("CompanyID", coID, "IsActive", true))
		if err != nil {
			return ficomodel.PeriodBlocked
		}
		fiscalYearID = yr.ID
	}

	filters := []*dbflex.Filter{dbflex.Eq("FiscalYearID", fiscalYearID)}
	filters = append(filters, dbflex.Lte("FromDate", date))
	filters = append(filters, dbflex.Gte("ToDate", date))

	period, err := datahub.GetByFilter(h, new(ficomodel.FiscalPeriod), dbflex.And(filters...))
	if err != nil {
		return ficomodel.PeriodBlocked
	}

	modStat, ok := period.Modules[module]
	if !ok {
		return ficomodel.PeriodBlocked
	}

	return modStat
}

func RegisterFiscalYear(s *kaos.Service, modDB, modUI kaos.Mod) {
	s.Group().SetMod(modDB, modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(ficomodel.FiscalPeriod), "fiscalperiod").DisableRoute("insert", "update", "save"),
		s.RegisterModel(new(ficomodel.FiscalPeriod), "fiscalperiod").AllowOnlyRoute("insert", "update", "save").RegisterMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				parm, ok := payload.(*ficomodel.FiscalPeriod)
				if ok {
					parm.CompanyID = tenantcorelogic.GetCompanyIDFromContext(ctx)
					if parm.FiscalYearID == "" {
						return false, errors.New("invalid fiscal year ID")
					}

					if parm.Modules["Finance"] != ficomodel.PeriodBlocked {
						hub := sebar.GetTenantDBFromContext(ctx)
						if hub == nil {
							return false, errors.New("missing: connection")
						}

						fiscalYear := new(ficomodel.FiscalYear)
						if e := hub.GetByID(fiscalYear, parm.FiscalYearID); e != nil {
							return false, fmt.Errorf("missing: fiscal year: %s", parm.FiscalYearID)
						}

						if fiscalYear.IsActive {
							if parm.Modules["Finance"] == ficomodel.PeriodClose {
								if e := hub.GetByFilter(new(ficomodel.FiscalPeriod), dbflex.And(
									dbflex.Eq("FiscalYearID", parm.FiscalYearID),
									dbflex.Ne("_id", parm.ID),
									dbflex.Or(
										dbflex.Eq("Modules.Finance", ficomodel.PeriodOpen),
										dbflex.Eq("Modules.Finance", ficomodel.PeriodBlocked),
									),
								)); e != nil {
									return false, errors.New("fiscal year active must have one or more periods with OPEN or BLOCKED status")
								}
							}
						} else {
							return false, errors.New("period status can only be OPEN or CLOSED, during active Fiscal Year")
						}
					}
				}
				return true, nil
			},
		),

		s.RegisterModel(new(ficomodel.FiscalYear), "fiscalyear").DisableRoute("insert", "update", "save"),
		s.RegisterModel(new(ficomodel.FiscalYear), "fiscalyear").AllowOnlyRoute("insert", "update", "save").RegisterMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				parm, ok := payload.(*ficomodel.FiscalYear)
				if ok {
					parm.CompanyID = tenantcorelogic.GetCompanyIDFromContext(ctx)
					if parm.ID == "" {
						parm.IsActive = false
					}

					if parm.IsActive {
						hub := sebar.GetTenantDBFromContext(ctx)
						if hub == nil {
							return false, errors.New("missing: connection")
						}

						if e := hub.GetByFilter(new(ficomodel.FiscalYear), dbflex.And(
							dbflex.Ne("_id", parm.ID),
							dbflex.Eq("CompanyID", parm.CompanyID),
							dbflex.Eq("IsActive", true),
						)); e == nil {
							return false, errors.New("previous fiscal year still active")
						}

						if e := hub.GetByFilter(new(ficomodel.FiscalPeriod), dbflex.And(
							dbflex.Eq("FiscalYearID", parm.ID),
							dbflex.Or(
								dbflex.Eq("Modules.Finance", ficomodel.PeriodOpen),
								dbflex.Eq("Modules.Finance", ficomodel.PeriodBlocked),
							),
						)); e != nil {
							return false, errors.New("active fiscal year must have one or more periods with OPEN or BLOCKED status")
						}
					}
				}
				return true, nil
			},
		),
	)
}
