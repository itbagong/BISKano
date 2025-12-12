package hcmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostGetsTalentDevelopment() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		tdIds := make([]interface{}, len(ms))
		for i, d := range ms {
			ids[i] = d["EmployeeID"]
			tdIds[i] = d["_id"]
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

		// get assesment
		assessments := []hcmmodel.TalentDevelopmentAssesment{}
		err = h.Gets(new(hcmmodel.TalentDevelopmentAssesment), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("TalentDevelopmentID", tdIds...),
			),
		), &assessments)
		if err != nil {
			return true, nil
		}
		mapAssessment := lo.Associate(assessments, func(m hcmmodel.TalentDevelopmentAssesment) (string, hcmmodel.TalentDevelopmentAssesment) {
			return m.TalentDevelopmentID, m
		})

		// get sk
		sks := []hcmmodel.TalentDevelopmentSK{}
		err = h.Gets(new(hcmmodel.TalentDevelopmentSK), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("TalentDevelopmentID", tdIds...),
			),
		), &sks)
		if err != nil {
			return true, nil
		}

		type sk struct {
			ID     string
			Type   hcmmodel.TalentDevelopmentSKType
			Status ficomodel.JournalStatus
		}

		mapSK := map[string][]sk{}
		for _, d := range sks {
			mapSK[d.TalentDevelopmentID] = append(mapSK[d.TalentDevelopmentID], sk{
				ID:     d.ID,
				Type:   d.Type,
				Status: d.Status,
			})
		}

		for _, m := range ms {
			empID := m.GetString("EmployeeID")
			empDetail := mapEmployeeDetail[empID]
			m["EmployeePOH"] = empDetail.POH
			m["EmployeeNIK"] = empDetail.EmployeeNo

			if v, ok := mapEmployee[empID]; ok {
				m["EmployeeName"] = v.Name
				m["EmployeeSite"] = v.Dimension.Get("Site")
				m["EmployeeJoinedDate"] = v.JoinDate
			}

			id := m.GetString("_id")
			if v, ok := mapAssessment[id]; ok {
				m["AssessmentStatus"] = v.Status
				m["AssessmentID"] = v.ID
			}

			if sks, ok := mapSK[id]; ok {
				for _, d := range sks {
					if d.Type == hcmmodel.TalentDevelopmentSKTypeActing {
						m["ActingSKStatus"] = d.Status
						m["ActingSKID"] = d.ID
					} else {
						m["PermanentSKStatus"] = d.Status
						m["PermanentSKID"] = d.Status
					}
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsOvertime() kaos.MWFunc {
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
			ids[i] = d["RequestorID"]
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, string) {
			return source.EmployeeID, source.Department
		})

		for _, m := range ms {
			empID := m.GetString("RequestorID")
			m["RequestorDepartment"] = mapEmployeeDetail[empID]
			m["RequestorName"] = mapEmployee[empID]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetOvertime() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(*hcmmodel.Overtime)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := codekit.M{}
		serde.Serde(res, &ms)

		ids := make([]string, len(res.Lines))
		for i, v := range res.Lines {
			ids[i] = v.EmployeeID
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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

		lines := make([]codekit.M, len(res.Lines))
		for i, line := range res.Lines {
			ln := codekit.M{}
			serde.Serde(line, &ln)

			empDetail := mapEmployeeDetail[line.EmployeeID]
			ln["EmployeeNIK"] = empDetail.EmployeeNo
			ln["EmployeePosition"] = empDetail.Position
			ln["EmployeeDepartment"] = empDetail.Department
			ln["EmployeeName"] = mapEmployee[line.EmployeeID]

			lines[i] = ln
		}
		ms["Lines"] = lines

		ctx.Data().Set("FnResult", ms)
		return true, nil
	}
}

func MWPostGetsBusinessTrip() kaos.MWFunc {
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
			ids[i] = d["RequestorID"]
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, string) {
			return source.EmployeeID, source.EmployeeNo
		})

		for _, m := range ms {
			empID := m.GetString("RequestorID")
			m["RequestorName"] = mapEmployee[empID]
			m["RequestorNIK"] = mapEmployeeDetail[empID]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetBusinessTrip() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(*hcmmodel.BusinessTrip)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := codekit.M{}
		serde.Serde(res, &ms)

		ids := make([]string, len(res.Lines))
		for i, v := range res.Lines {
			ids[i] = v.EmployeeID
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
		})
		// Dimension
		mapEmployeeDim := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, tenantcoremodel.Dimension) {
			return source.ID, source.Dimension
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

		lines := make([]codekit.M, len(res.Lines))
		for i, line := range res.Lines {
			ln := codekit.M{}
			serde.Serde(line, &ln)

			empDetail := mapEmployeeDetail[line.EmployeeID]
			ln["EmployeeNIK"] = empDetail.EmployeeNo
			ln["EmployeePosition"] = empDetail.Position
			ln["EmployeeDepartment"] = mapEmployeeDim[line.EmployeeID].Get("CC")
			ln["EmployeeLevel"] = empDetail.Level
			ln["EmployeeSite"] = mapEmployeeDim[line.EmployeeID].Get("Site")
			ln["EmployeeName"] = mapEmployee[line.EmployeeID]

			lines[i] = ln
		}
		ms["Lines"] = lines

		ctx.Data().Set("FnResult", ms)
		return true, nil
	}
}

func MWPostGetsLeaveCompensation() kaos.MWFunc {
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
			ids[i] = d["RequestorID"]
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, string) {
			return source.EmployeeID, source.EmployeeNo
		})

		for _, m := range ms {
			empID := m.GetString("RequestorID")
			m["RequestorID"] = mapEmployee[empID]
			m["RequestorNIK"] = mapEmployeeDetail[empID]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsLoan() kaos.MWFunc {
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
			m["EmployeeSite"] = mapEmployee[empID].Dimension.Get("Site")
			m["EmployeeNIK"] = mapEmployeeDetail[empID].EmployeeNo
			m["EmployeePosition"] = mapEmployeeDetail[empID].Position
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsWorkTermination() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms)*2)
		i := 0
		for _, d := range ms {
			ids[i] = d["EmployeeID"]
			i++
			ids[i] = d["Requestor"]
			i++
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, string) {
			return source.EmployeeID, source.EmployeeNo
		})

		for _, m := range ms {
			empID := m.GetString("EmployeeID")
			m["EmployeeID"] = mapEmployee[empID]
			m["EmployeeNIK"] = mapEmployeeDetail[empID]
			req := m.GetString("Requestor")
			m["Requestor"] = mapEmployee[req]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsCoachingViolation() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms)*2)
		i := 0
		for _, d := range ms {
			ids[i] = d["EmployeeID"]
			i++
			ids[i] = d["RequestorID"]
			i++
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
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
		mapEmployeeDetail := lo.Associate(employeeDetails, func(source bagongmodel.EmployeeDetail) (string, string) {
			return source.EmployeeID, source.EmployeeNo
		})

		for _, m := range ms {
			empID := m.GetString("EmployeeID")
			m["EmployeeID"] = mapEmployee[empID]
			m["EmployeeNIK"] = mapEmployeeDetail[empID]
			m["RequestorID"] = mapEmployee[m.GetString("RequestorID")]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsTraining() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		i := 0
		for _, d := range ms {
			ids[i] = d["CandidateID"]
			i++
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
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
		})

		for _, m := range ms {
			m["Name"] = mapEmployee[m.GetString("CandidateID")]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsTrainingDevelopmentParticipant() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		var trainingID interface{}
		empIDs := make([]interface{}, len(ms))
		jobIDs := make([]interface{}, len(ms))
		for i := range ms {
			trainingID = ms[i]["TrainingCenterID"]
			empIDs[i] = ms[i]["EmployeeID"]
			jobIDs[i] = ms[i]["ManpowerRequestID"]
		}

		if len(ms) == 0 {
			return true, nil
		}

		// get employee
		employees := []tenantcoremodel.Employee{}
		err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", empIDs...),
			),
		), &employees)
		if err != nil {
			return true, nil
		}
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, tenantcoremodel.Employee) {
			return source.ID, source
		})

		training := new(hcmmodel.TrainingDevelopment)
		h.GetByID(training, trainingID)

		// get employee
		employeeDetails := []bagongmodel.EmployeeDetail{}
		err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("EmployeeID", empIDs...),
			),
		), &employeeDetails)
		if err != nil {
			return true, nil
		}

		mapEmployeeDetail := map[string]bagongmodel.EmployeeDetail{}
		positionIDs := make([]interface{}, len(employeeDetails))
		for i, emp := range employeeDetails {
			mapEmployeeDetail[emp.EmployeeID] = emp
			positionIDs[i] = emp.Position
		}

		mapPosition := map[string]string{}
		if training.TrainingType == "General" {
			// get position
			positions := []tenantcoremodel.MasterData{}
			err = h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.In("_id", positionIDs...),
				),
			), &positions)
			if err != nil {
				return true, nil
			}

			mapPosition = lo.Associate(positions, func(source tenantcoremodel.MasterData) (string, string) {
				return source.ID, source.Name
			})
		} else {
			// get manpower request
			jobs := []hcmmodel.ManpowerRequest{}
			err = h.Gets(new(hcmmodel.ManpowerRequest), dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.In("_id", jobIDs...),
				),
			), &jobs)
			if err != nil {
				return true, nil
			}

			positionIDs := make([]interface{}, len(jobs))
			for i, job := range jobs {
				positionIDs[i] = job.Position
			}

			// get position
			positions := []tenantcoremodel.MasterData{}
			err = h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.In("_id", positionIDs...),
				),
			), &positions)
			if err != nil {
				return true, nil
			}

			mapPosition = lo.Associate(positions, func(source tenantcoremodel.MasterData) (string, string) {
				return source.ID, source.Name
			})
		}

		for i := range ms {
			emp := mapEmployee[ms[i].GetString("EmployeeID")]
			empDet := mapEmployeeDetail[emp.ID]
			ms[i]["Name"] = emp.Name
			ms[i]["Site"] = emp.Dimension.Get("Site")
			ms[i]["Position"] = mapPosition[empDet.Position]
			ms[i]["Department"] = emp.Dimension.Get("CC")
			ms[i]["NIK"] = empDet.EmployeeNo
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsTrainingDevelopmentAttendance() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		i := 0
		for _, d := range ms {
			ids[i] = d["LocationID"]
			i++
		}

		// get master location
		locations := []tenantcoremodel.MasterData{}
		err := h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", ids...),
			),
		), &locations)
		if err != nil {
			return true, nil
		}
		mapLocation := lo.Associate(locations, func(source tenantcoremodel.MasterData) (string, string) {
			return source.ID, source.Name
		})

		for _, m := range ms {
			m["LocationName"] = mapLocation[m.GetString("LocationID")]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

// MWPreAssignSequenceNo
func MWPreAssignSequenceNo() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		id, _ := m["_id"].(string)
		if id != "" {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		if h == nil {
			return false, errors.New("missing: db connection")
		}

		journalID := m.GetString("JournalTypeID")
		journalType := new(hcmmodel.JournalType)
		err := h.GetByID(journalType, journalID)
		if err != nil {
			return false, fmt.Errorf("error when get journal type: %s", err.Error())
		}

		kind := ""
		switch journalType.TransactionType {
		case "Work Termination - Resign":
			kind = "WorkTerminationResign"
		case "Work Termination - PHK":
			kind = "WorkTerminationPHK"
		case "Work Termination - Sakit Berkepanjangan":
			kind = "WorkTerminationSakit"
		case "Work Termination - Meninggal":
			kind = "WorkTerminationMeninggal"
		case "Work Termination - Pensiun":
			kind = "WorkTerminationPensiun"
		case "Coaching & Violation - SP1":
			kind = "CoachingViolationSP1"
		case "Coaching & Violation - SP2":
			kind = "CoachingViolationSP2"
		case "Coaching & Violation - SP3":
			kind = "CoachingViolationSP3"
		case "Coaching & Violation - Surat Teguran Tertulis":
			kind = "CoachingViolationSuratTeguranTertulis"
		case "Coaching & Violation - Surat Panggilan Masuk Kerja 1":
			kind = "CoachingViolationPanggilanMasukKerja1"
		case "Coaching & Violation - Surat Panggilan Masuk Kerja 2":
			kind = "CoachingViolationPanggilanMasukKerja2"
		case "Coaching & Violation - Form Coaching":
			kind = "CoachingViolationFormCoaching"
		case "Talent Development - Promotion - General & Benefit":
			kind = "TalentDevelopmentPromotionGeneral"
		case "Talent Development - Promotion - Tracking Assessment":
			kind = "TalentDevelopmentPromotionAssessment"
		case "Talent Development - Promotion - Tracking SK Acting":
			kind = "TalentDevelopmentPromotionSKActing"
		case "Talent Development - Promotion - Tracking SK Tetap":
			kind = "TalentDevelopmentPromotionSKTetap"
		case "Talent Development - Rotation":
			kind = "TalentDevelopmentRotation"
		case "Talent Development - Mutation":
			kind = "TalentDevelopmentDemotion"
		case "Talent Development - Salary Change":
			kind = "TalentDevelopmentSalaryChange"
		case "Talent Development - POH Change":
			kind = "TalentDevelopmentPOH"
		}

		tenantcorelogic.MWPreAssignSequenceNo(kind, false, "")(ctx, payload)

		return true, nil
	}
}

func MWPostGetsManpower() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		jobIDs := make([]interface{}, len(ms))
		i := 0
		for _, d := range ms {
			jobIDs[i] = d["JobVacancyTitle"]
			ids[i] = d["RequestorID"]
			i++
		}

		// get master employee
		employees := []tenantcoremodel.Employee{}
		err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", ids...),
			),
		), &employees)
		if err != nil {
			return true, nil
		}
		mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
			return source.ID, source.Name
		})

		// get master data
		masters := []tenantcoremodel.MasterData{}
		err = h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", jobIDs...),
			),
		), &masters)
		if err != nil {
			return true, nil
		}
		mapMaster := lo.Associate(masters, func(source tenantcoremodel.MasterData) (string, string) {
			return source.ID, source.Name
		})

		for _, m := range ms {
			m["RequestorID"] = mapEmployee[m.GetString("RequestorID")]
			m["JobVacancyTitle"] = mapMaster[m.GetString("JobVacancyTitle")]
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetsTrainingDevelopment() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		i := 0
		for _, d := range ms {
			ids[i] = d["_id"]
			i++
		}

		// get detail
		details := []hcmmodel.TrainingDevelopmentDetail{}
		err := h.Gets(new(hcmmodel.TrainingDevelopmentDetail), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("TrainingCenterID", ids...),
			),
		), &details)
		if err != nil {
			return true, nil
		}
		mapDetail := lo.Associate(details, func(source hcmmodel.TrainingDevelopmentDetail) (string, hcmmodel.TrainingDevelopmentDetail) {
			return source.TrainingCenterID, source
		})

		for _, m := range ms {
			m["StatusDetail"] = mapDetail[m.GetString("_id")].Status
			m["TrainingDevelopmentDetailID"] = mapDetail[m.GetString("_id")].ID
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
