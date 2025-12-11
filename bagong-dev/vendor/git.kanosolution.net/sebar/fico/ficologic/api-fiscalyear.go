package ficologic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
)

func (obj *FiscalYearHandler) Create(ctx *kaos.Context, payload *ficomodel.FiscalYear) (*ficomodel.FiscalYear, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.IsActive {
		// check if there is other active, active is exclusive
		other, err := datahub.GetByFilter(h, new(ficomodel.FiscalYear), dbflex.Eqs("CompanyID", payload.CompanyID, "IsActive", true))
		if err == nil {
			return nil, fmt.Errorf("invalid: other fiscal year %s, is active", other.ID)
		}
	}

	payload.CompanyID = tenantcorelogic.GetCompanyIDFromContext(ctx)
	payload, err := createFiscalYear(h, payload, ficoconfig.Config.FinancialPeriodModules...)
	if err != nil {
		return nil, ctx.Log().Error2("unable to create fiscal year", "create fiscal year: %s", err.Error())
	}

	return payload, nil
}
