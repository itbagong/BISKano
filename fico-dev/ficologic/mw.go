package ficologic

import (
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostPostingProfileID(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"PostingProfileID"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		postingProfileIDs := []interface{}{}
		for _, field := range fields {
			postingProfileIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})
			postingProfileIDs = append(postingProfileIDs, postingProfileIDFields...)
		}

		profiles := []ficomodel.PostingProfile{}
		h.GetsByFilter(new(ficomodel.PostingProfile), dbflex.In("_id", postingProfileIDs...), &profiles)
		mapProfiles := lo.Associate(profiles, func(profile ficomodel.PostingProfile) (string, ficomodel.PostingProfile) {
			return profile.ID, profile
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapProfiles[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostTaxCodeIDs(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"TaxCodes"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		taxCodeIDs := []interface{}{}
		for _, field := range fields {
			for _, m := range ms {
				if val, ok := m.Get(field).([]string); ok {
					for _, v := range val {
						taxCodeIDs = append(taxCodeIDs, v)
					}
				}
			}
		}

		taxCodes := []ficomodel.TaxSetup{}
		h.GetsByFilter(new(ficomodel.TaxSetup), dbflex.In("_id", taxCodeIDs...), &taxCodes)
		mapTaxCodes := lo.Associate(taxCodes, func(taxcode ficomodel.TaxSetup) (string, ficomodel.TaxSetup) {
			return taxcode.ID, taxcode
		})

		for _, m := range ms {
			for _, field := range fields {
				taxIDs, ok := m.Get(field).([]string)
				if ok {
					for keyIDOfTax, id := range taxIDs {
						if v, ok := mapTaxCodes[id]; ok {
							taxIDs[keyIDOfTax] = v.Name
						}
					}

					if len(taxIDs) > 0 {
						m.Set(field, taxIDs)
					}
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostChargeCodeIDs(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"ChargeCodes"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		chargeCodeIDs := []interface{}{}
		for _, field := range fields {
			for _, m := range ms {
				if val, ok := m.Get(field).([]string); ok {
					for _, v := range val {
						chargeCodeIDs = append(chargeCodeIDs, v)
					}
				}
			}
		}

		chargeCodes := []ficomodel.ChargeSetup{}
		h.GetsByFilter(new(ficomodel.ChargeSetup), dbflex.In("_id", chargeCodeIDs...), &chargeCodes)
		mapChargeCodes := lo.Associate(chargeCodes, func(chargecode ficomodel.ChargeSetup) (string, ficomodel.ChargeSetup) {
			return chargecode.ID, chargecode
		})

		for _, m := range ms {
			for _, field := range fields {
				chargeIDs, ok := m.Get(field).([]string)
				if ok {
					for keyIDOfCharge, id := range chargeIDs {
						if v, ok := mapChargeCodes[id]; ok {
							chargeIDs[keyIDOfCharge] = v.Name
						}
					}

					if len(chargeIDs) > 0 {
						m.Set(field, chargeIDs)
					}
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreFilterCompanyID(ctx *kaos.Context, i interface{}) (bool, error) {
	ctx.Data().Set(tenantcoremodel.DBModFilter, []*dbflex.Filter{dbflex.Eq("CompanyID", tenantcorelogic.GetCompanyIDFromContext(ctx))})
	return true, nil
}

func MWPostSavePostingProfilePIC() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(*ficomodel.PostingProfilePIC)
		if !ok {
			return true, nil
		}

		userID := sebar.GetUserIDFromCtx(ctx)
		if userID == "" {
			userID = "SYSTEM"
			if ctx.Data().Get("UserID", "").(string) != "" {
				userID = ctx.Data().Get("UserID", "").(string)
			}
		}

		h := sebar.GetTenantDBFromContext(ctx)
		log := ficomodel.PostingProfilePICLog{
			UpdatedBy:         userID,
			PostingProfilePIC: res,
			Action:            "update",
		}

		pp := new(ficomodel.PostingProfile)
		err := h.GetByParm(pp, dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("_id", res.PostingProfileID),
		))
		if err != nil && err != io.EOF {
			return true, nil
		}

		log.PostingProfileName = pp.Name
		h.Save(&log)

		return true, nil
	}
}

func MWPostSavePostingProfile() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(*ficomodel.PostingProfile)
		if !ok {
			return true, nil
		}

		userID := sebar.GetUserIDFromCtx(ctx)
		if userID == "" {
			userID = "SYSTEM"
			if ctx.Data().Get("UserID", "").(string) != "" {
				userID = ctx.Data().Get("UserID", "").(string)
			}
		}

		h := sebar.GetTenantDBFromContext(ctx)
		log := ficomodel.PostingProfileLog{
			UpdatedBy:      userID,
			PostingProfile: res,
		}

		h.Save(&log)

		return true, nil
	}
}
func MWPostGetsLoanSetup() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		for i, d := range ms {
			ids[i] = d["EmployeeID"]
		}
		// get employee
		employees := []tenantcoremodel.Employee{}
		err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", ids...),
			),
		), &employees)
		if err != nil {
			return true, nil
		}
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, tenantcoremodel.Employee) {
			return source.ID, source
		})

		// get employee detail
		employeeDetails := []bagongmodel.EmployeeDetail{}
		err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("EmployeeID", ids...),
			),
		), &employeeDetails)
		if err != nil {
			return true, nil
		}
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, bagongmodel.EmployeeDetail) {
			return source.EmployeeID, source
		})

		for _, m := range ms {
			empID := m.GetString("EmployeeID")
			m["EmployeeName"] = mapEmployee[empID].Name
			m["EmployeeNIK"] = mapEmployeeDetail[empID].EmployeeNo
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
