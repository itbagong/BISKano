package shelogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

func MWPreCoachingAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*shemodel.Coaching)
	if !ok {
		return false, errors.New("invalid payload type")
	}

	if record.Feedback != "" {
		record.Status = string(shemodel.SHEStatusCompleted)
	}

	return true, nil
}

func MWPostCoaching(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		employees := []tenantcoremodel.Employee{}
		h.GetsByFilter(new(tenantcoremodel.Employee), nil, &employees)
		mapEmployees := lo.Associate(employees, func(emp tenantcoremodel.Employee) (string, string) {
			return emp.ID, emp.Name
		})

		employeePositions := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "PTE"), &employeePositions)
		mapEmployeePositions := lo.Associate(employeePositions, func(emp tenantcoremodel.MasterData) (string, string) {
			return emp.ID, emp.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if coachValue, ok := m["Coach"]; ok {
				if coachStr, ok := coachValue.(string); ok {
					if val, ok := mapEmployees[coachStr]; ok {
						m["Coach"] = val
					}
				}
			}
			if coacheeValue, ok := m["Coachee"]; ok {
				if coacheeStr, ok := coacheeValue.(string); ok {
					if val, ok := mapEmployees[coacheeStr]; ok {
						m["Coachee"] = val
					}
				}
			}

			if coachTitleValue, ok := m["CoachTitle"]; ok {
				if coachTitleStr, ok := coachTitleValue.(string); ok {
					if val, ok := mapEmployeePositions[coachTitleStr]; ok {
						m["CoachTitle"] = val
					}
				}
			}
			if coacheeTitleValue, ok := m["CoacheeTitle"]; ok {
				if coacheeTitleStr, ok := coacheeTitleValue.(string); ok {
					if val, ok := mapEmployeePositions[coacheeTitleStr]; ok {
						m["CoacheeTitle"] = val
					}
				}
			}

			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostCSMS(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		customers := []tenantcoremodel.Customer{}
		h.GetsByFilter(new(tenantcoremodel.Customer), nil, &customers)
		mapcustomers := lo.Associate(customers, func(cust tenantcoremodel.Customer) (string, string) {
			return cust.ID, cust.Name
		})

		itemTemplates := []shemodel.MCUItemTemplate{}
		h.GetsByFilter(new(shemodel.MCUItemTemplate), dbflex.Eq("Menu", "SHE-0010"), &itemTemplates)
		mapItemTemplates := lo.Associate(itemTemplates, func(item shemodel.MCUItemTemplate) (string, string) {
			return item.ID, item.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if custValue, ok := m["Customer"]; ok {
				if custStr, ok := custValue.(string); ok {
					if val, ok := mapcustomers[custStr]; ok {
						m["Customer"] = val
					}
				}
			}
			if templateValue, ok := m["TemplateID"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapItemTemplates[templateStr]; ok {
						m["TemplateID"] = val
					}
				}
			}

			var endDateRaw time.Time
			if m["EndDate"] != nil {
				endDateRaw = m["EndDate"].(time.Time)
			}
			now := time.Now()
			status := "EXPIRED"
			if now.Before(endDateRaw) {
				status = "ACTIVE"
			}
			m["CSMSStatus"] = status
			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostJSA(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		customers := []tenantcoremodel.Customer{}
		h.GetsByFilter(new(tenantcoremodel.Customer), nil, &customers)
		mapcustomers := lo.Associate(customers, func(cust tenantcoremodel.Customer) (string, string) {
			return cust.ID, cust.Name
		})

		masterDatas := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), nil, &masterDatas)
		mapmasterDatas := lo.Associate(masterDatas, func(emp tenantcoremodel.MasterData) (string, string) {
			return emp.ID, emp.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if custValue, ok := m["Customer"]; ok {
				if custStr, ok := custValue.(string); ok {
					if val, ok := mapcustomers[custStr]; ok {
						m["Customer"] = val
					}
				}
			}
			if templateValue, ok := m["PositionInvolved"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapmasterDatas[templateStr]; ok {
						m["PositionInvolved"] = val
					}
				}
			}
			if templateValue, ok := m["Location"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapmasterDatas[templateStr]; ok {
						m["Location"] = val
					}
				}
			}
			if templateValue, ok := m["EquipmentInvolved"]; ok {
				if templateStr, ok := templateValue.([]string); ok {
					if len(templateStr) > 0 {
						temp := []string{}
						for _, xval := range templateStr {
							if val, ok := mapmasterDatas[xval]; ok {
								temp = append(temp, val)
							}
						}
						m["EquipmentInvolved"] = temp
					}
				}
			}
			if templateValue, ok := m["Apd"]; ok {
				if templateStr, ok := templateValue.([]string); ok {
					if len(templateStr) > 0 {
						temp := []string{}
						for _, xval := range templateStr {
							if val, ok := mapmasterDatas[xval]; ok {
								temp = append(temp, val)
							}
						}
						m["Apd"] = temp
					}
				}
			}

			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreInspectionAssignDefault(ctx *kaos.Context, i interface{}) (bool, error) {
	record, ok := i.(*shemodel.Inspection)
	if !ok {
		return false, errors.New("invalid payload type")
	}

	if record.Inspectors == "" {
		userID := sebar.GetUserIDFromCtx(ctx)
		record.Inspectors = userID
	}

	return true, nil
}

func MWPostLegalCompliance(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		siteIds := make([]interface{}, len(ms))
		for i, m := range ms {
			if val, ok := m.Get("_id").(string); ok {
				lcids[i] = val
			}

			if val, ok := m.Get("SiteID").(string); ok {
				siteIds[i] = val
			}
		}

		Sites, err := GetSites(ctx, dbflex.In("_id", siteIds...))
		if err != nil {
			return false, errors.New("error GetSites: " + err.Error())
		}

		legalCompliances := []shemodel.LegalCompliance{}
		h.GetsByFilter(new(shemodel.LegalCompliance), dbflex.In("_id", lcids...), &legalCompliances)
		maplegalCompliances := lo.Associate(legalCompliances, func(legalCompliance shemodel.LegalCompliance) (string, shemodel.LegalCompliance) {
			return legalCompliance.SiteID, legalCompliance
		})

		pipe := []bson.M{
			{
				"$match": bson.M{
					"RelatedSite": bson.M{"$in": siteIds},
				},
			},
			{
				"$group": bson.M{
					"_id":        "$RelatedSite",
					"Compliance": bson.M{"$avg": "$Achievement"},
				},
			},
		}

		type register struct {
			SiteID     string `bson:"_id"`
			Compliance float64
		}

		legalRegisters := []register{}
		cmd := dbflex.From(new(shemodel.LegalRegister).TableName()).Command("pipe", pipe)
		if _, err := h.Populate(cmd, &legalRegisters); err != nil {
			return false, fmt.Errorf("err when get legal register: %s", err.Error())
		}
		maplegalRegisters := lo.Associate(legalRegisters, func(l register) (string, float64) {
			return l.SiteID, l.Compliance
		})

		datas := make([]codekit.M, 0)
		for _, site := range Sites {
			data := codekit.M{}

			data["_id"] = ""
			data["SiteID"] = site.ID
			data["SiteName"] = site.Name
			data["Compliance"] = 0.0

			if val, ok := maplegalCompliances[site.ID]; ok {
				data["_id"] = val.ID
				data["Compliance"] = maplegalRegisters[site.ID]
			}

			datas = append(datas, data)
		}

		res.Set("data", datas)
		res.Set("count", len(datas))
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostSafetyCard(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		MasterDataType := []tenantcoremodel.MasterDataType{}
		h.GetsByFilter(new(tenantcoremodel.MasterDataType), nil, &MasterDataType)
		mapmasterDatasType := lo.Associate(MasterDataType, func(emp tenantcoremodel.MasterDataType) (string, string) {
			return emp.ID, emp.Name
		})

		MasterLOC := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "LOC"), &MasterLOC)
		mapMasterLOC := lo.Associate(MasterLOC, func(loc tenantcoremodel.MasterData) (string, string) {
			return loc.ID, loc.Name
		})

		MasterPTE := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "PTE"), &MasterPTE)
		mapMasterPTE := lo.Associate(MasterPTE, func(loc tenantcoremodel.MasterData) (string, string) {
			return loc.ID, loc.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if templateValue, ok := m["CategoryID"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapmasterDatasType[templateStr]; ok {
						m["CategoryID"] = val
					}
				}
			}

			if templateValue, ok := m["LocationID"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterLOC[templateStr]; ok {
						m["LocationID"] = val
					}
				}
			}
			if templateValue, ok := m["PositionID"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterPTE[templateStr]; ok {
						m["PositionID"] = val
					}
				}
			}
			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostMeeting(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		MasterLOC := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "LOC"), &MasterLOC)
		mapMasterLOC := lo.Associate(MasterLOC, func(loc tenantcoremodel.MasterData) (string, string) {
			return loc.ID, loc.Name
		})

		MasterMTY := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "MTY"), &MasterMTY)
		mapMasterMTY := lo.Associate(MasterMTY, func(loc tenantcoremodel.MasterData) (string, string) {
			return loc.ID, loc.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if templateValue, ok := m["Location"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterLOC[templateStr]; ok {
						m["Location"] = val
					}
				}
			}
			if templateValue, ok := m["MeetingType"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterMTY[templateStr]; ok {
						m["MeetingType"] = val
					}
				}
			}
			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
func MWPostLegalRegister(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		MasterLTY := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "LTY"), &MasterLTY)
		mapMasterLTY := lo.Associate(MasterLTY, func(mtr tenantcoremodel.MasterData) (string, string) {
			return mtr.ID, mtr.Name
		})

		MasterLFI := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "LFI"), &MasterLFI)
		mapMasterLFI := lo.Associate(MasterLFI, func(mtr tenantcoremodel.MasterData) (string, string) {
			return mtr.ID, mtr.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if templateValue, ok := m["Type"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterLTY[templateStr]; ok {
						m["Type"] = val
					}
				}
			}

			if templateValue, ok := m["Fields"]; ok {
				if templateAry, ok := templateValue.([]string); ok {
					Fields := []string{}
					for _, LFI := range templateAry {
						if _, ok := mapMasterLFI[LFI]; ok {
							Fields = append(Fields, mapMasterLFI[LFI])
						}
					}
					m["Fields"] = Fields
				}
			}

			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
func MWPostInvestigasi(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lcids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					lcids[i] = val
				}
			}
		}

		MasterLOC := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "LOC"), &MasterLOC)
		mapMasterLOC := lo.Associate(MasterLOC, func(loc tenantcoremodel.MasterData) (string, string) {
			return loc.ID, loc.Name
		})

		MasterAC := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.Eq("MasterDataTypeID", "AccidentClassification"), &MasterAC)
		mapMasterAC := lo.Associate(MasterAC, func(ac tenantcoremodel.MasterData) (string, string) {
			return ac.ID, ac.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if templateValue, ok := m["Location"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterLOC[templateStr]; ok {
						m["Location"] = val
					}
				}
			}

			if templateValue, ok := m["Classification"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterAC[templateStr]; ok {
						m["Classification"] = val
					}
				}
			}
			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostMCUTransaction() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		employeeIds := make([]interface{}, len(ms))
		for i, m := range ms {
			if val, ok := m.Get("Customer").(string); ok {
				employeeIds[i] = val
			}
		}

		masterData := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.In("MasterDataTypeID", []string{"GEME", "PTE", "MPR", "MPU"}...), &masterData)
		mapMasterData := lo.Associate(masterData, func(master tenantcoremodel.MasterData) (string, string) {
			return master.ID, master.Name
		})

		employees := []tenantcoremodel.Employee{}
		h.GetsByFilter(new(tenantcoremodel.Employee), dbflex.In("_id", employeeIds...), &employees)
		mapEmployee := lo.Associate(employees, func(ac tenantcoremodel.Employee) (string, string) {
			return ac.ID, ac.Name
		})

		datas := make([]codekit.M, 0)
		for _, m := range ms {
			if templateValue, ok := m["Gender"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterData[templateStr]; ok {
						m["Gender"] = val
					}
				}
			}

			if templateValue, ok := m["Position"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterData[templateStr]; ok {
						m["Position"] = val
					}
				}
			}

			if templateValue, ok := m["Purpose"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterData[templateStr]; ok {
						m["Purpose"] = val
					}
				}
			}

			if templateValue, ok := m["Provider"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapMasterData[templateStr]; ok {
						m["Provider"] = val
					}
				}
			}

			if templateValue, ok := m["Customer"]; ok {
				if templateStr, ok := templateValue.(string); ok {
					if val, ok := mapEmployee[templateStr]; ok {
						m["Customer"] = val
					}
				}
			}

			datas = append(datas, m)
		}

		res.Set("data", datas)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
