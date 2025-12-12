package hcmlogic

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/kasset"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/xuri/excelize/v2"
)

type GenerateCertificateRequest struct {
	ParticipantID string
}

func (m *TrainingDevelopmentHandler) GenerateCertificate(ctx *kaos.Context, payload *GenerateCertificateRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err := h.GetByID(participant, payload.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	employeeDetail := new(bagongmodel.EmployeeDetail)
	err = h.GetByFilter(employeeDetail, dbflex.Eq("EmployeeID", participant.EmployeeID))
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	employee := new(tenantcoremodel.Employee)
	err = h.GetByID(employee, participant.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	tdc := new(hcmmodel.TrainingDevelopment)
	err = h.GetByID(tdc, participant.TrainingCenterID)
	if err != nil {
		return nil, fmt.Errorf("error when get training: %s", err.Error())
	}

	tdcDetail := new(hcmmodel.TrainingDevelopmentDetail)
	err = h.GetByFilter(tdcDetail, dbflex.Eq("TrainingCenterID", participant.TrainingCenterID))
	if err != nil {
		return nil, fmt.Errorf("error when get training detail: %s", err.Error())
	}

	managerDetail := new(bagongmodel.EmployeeDetail)
	err = h.GetByFilter(managerDetail, dbflex.And(
		dbflex.Eq("Position", "PTE001"),
		dbflex.Eq("WorkerStatus", "Aktif"),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get manager detail: %s", err.Error())
	}

	manager := new(tenantcoremodel.Employee)
	err = h.GetByID(manager, managerDetail.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("error when get manager: %s", err.Error())
	}

	managerSignature := new(kasset.Asset)
	err = h.GetByFilter(managerSignature, dbflex.And(
		dbflex.Eq("Kind", "Employee Signature"),
		dbflex.Eq("RefID", manager.ID),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get manager signature: %s", err.Error())
	}

	id := codekit.M{"_id": "", "Dimension": tenantcoremodel.Dimension{}}
	now := time.Now()
	sequenceKind := fmt.Sprintf("TDC%d%02d", now.Year(), int(now.Month()))
	err = h.Get(new(tenantcoremodel.NumberSequence))
	if err == io.EOF {
		seq := new(tenantcoremodel.NumberSequence)
		seq.ID = sequenceKind
		seq.Name = sequenceKind
		seq.OutFormat = "BDM/TDC/${dt:2006}/${roman_month}/${counter:%03d}"
		seq.LastNo = 0
		h.Save(seq)

		coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
		if coID == "DEMO" || coID == "" {
			return nil, errors.New("missing: Company, please relogin")
		}
		seqSetup := new(tenantcoremodel.NumberSequenceSetup)
		seqSetup.ID = sequenceKind
		seqSetup.CompanyID = coID
		seqSetup.Kind = sequenceKind
		seqSetup.Label = sequenceKind
		seqSetup.NumSeqID = sequenceKind
		h.Save(seqSetup)
	}
	tenantcorelogic.MWPreAssignCustomSequenceNo(sequenceKind)(ctx, id)

	postTestScore := 0.0
	count := 0.0
	for _, p := range participant.TestDetails {
		if p.Stage == hcmmodel.TDCTestTypePost {
			postTestScore += float64(p.Score)
		}
		count++
	}

	// practice test
	practiceScore := 0.0
	switch tdcDetail.AssessmentType {
	case "Assessment Staff":
		practice := new(hcmmodel.TrainingDevelopmentPracticeTestStaff)
		err = h.GetByFilter(practice, dbflex.Eq("ParticipantID", participant.ID))
		if err != nil {
			return nil, fmt.Errorf("error when get staff practice test: %s", err.Error())
		}

		practiceScore = practice.TotalScore
	case "Assessment Mechanic":
		practice := new(hcmmodel.TrainingDevelopmentPracticeScore)
		err = h.GetByFilter(practice, dbflex.And(
			dbflex.Eq("TrainingCenterID", participant.TrainingCenterID),
			dbflex.Eq("EmployeeID", participant.EmployeeID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get mechanic practice test: %s", err.Error())
		}

		practiceScore = practice.FinalScore
	case "Assessment Driver":
		practices := []hcmmodel.TrainingDevelopmentPracticeScore{}
		err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeScore), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("TrainingCenterID", participant.TrainingCenterID),
				dbflex.Eq("EmployeeID", participant.EmployeeID),
			),
		), &practices)
		if err != nil {
			return nil, fmt.Errorf("error when get driver practice test: %s", err.Error())
		}

		score := 0.0
		for _, p := range practices {
			score += p.FinalScore
		}

		practiceScore = lo.Ternary(len(practices) == 0, 0, score/float64(len(practices)))
	}

	classes := strings.Split(tdcDetail.MaterialClass, "\n")
	practices := strings.Split(tdcDetail.PracticeClass, "\n")
	data := codekit.M{
		"No":                     id["_id"],
		"Title":                  tdc.TrainingTitle,
		"EmployeeName":           employee.Name,
		"EmployeeNIK":            employeeDetail.EmployeeNo,
		"TrainingDate":           fmt.Sprintf("%s - %s", tdcDetail.TrainingDateFrom.Format("02 Jan 06"), tdcDetail.TrainingDateTo.Format("02 Jan 06")),
		"ClassMaterials":         classes,
		"PracticeMaterials":      practices,
		"ManagerName":            manager.Name,
		"PostTestScore":          lo.Ternary(count == 0, 0, postTestScore/count),
		"PracticeScore":          fmt.Sprintf("%.2f", practiceScore),
		"AvgScore":               fmt.Sprintf("%.2f", (practiceScore+postTestScore)/2),
		"ManagerSignature":       managerSignature.URI,
		"IsShowTrainerSignature": false,
	}

	if tdcDetail.TrainerType == "External" {
		data["TrainerName"] = tdcDetail.TrainerName
	} else if tdcDetail.TrainerType == "Internal" {
		trainer := new(tenantcoremodel.Employee)
		err = h.GetByFilter(trainer, dbflex.Eq("_id", tdcDetail.TrainerName))
		if err != nil {
			return nil, fmt.Errorf("error when get trainer: %s", err.Error())
		}

		data["TrainerName"] = trainer.Name

		trainerSignature := new(kasset.Asset)
		err = h.GetByFilter(trainerSignature, dbflex.And(
			dbflex.Eq("Kind", "Employee Signature"),
			dbflex.Eq("RefID", trainer.ID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get trainer signature: %s", err.Error())
		}
		data["TrainerSignature"] = trainerSignature.URI
		data["IsShowTrainerSignature"] = true
	}

	pdfData := &tenantcorelogic.GenerateFromTemplateRequest{
		PDFFromTemplateRequest: tenantcorelogic.PDFFromTemplateRequest{
			TemplateName: "training-certificate/tdc_training_certificate",
			Data:         data,
		},
		Asset: kasset.Asset{
			ID:               payload.ParticipantID,
			OriginalFileName: fmt.Sprintf("%s.pdf", payload.ParticipantID),
			NewFileName:      fmt.Sprintf("%s.pdf", payload.ParticipantID),
			URI:              fmt.Sprintf("%s.pdf", payload.ParticipantID),
			ContentType:      "application/pdf",
		},
	}

	url := "/v1/tenant/pdf/generate-from-template"

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("nil: EventHub")
	}

	asset := kasset.Asset{}
	err = ev.Publish(
		url,
		pdfData,
		&asset,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return asset, nil
}

type GenerateTraining struct {
	TrainingCenterID string
}

func (m *TrainingDevelopmentHandler) GenerateTraining(ctx *kaos.Context, payload *GenerateTraining) (interface{}, error) {
	w, wOK := ctx.Data().Get("http_writer", nil).(http.ResponseWriter)

	if !wOK {
		return nil, errors.New("not a http compliant writer")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	tdc := new(hcmmodel.TrainingDevelopment)
	err := h.GetByID(tdc, payload.TrainingCenterID)
	if err != nil {
		return nil, fmt.Errorf("error when get training: %s", err.Error())
	}

	tdcDetail := new(hcmmodel.TrainingDevelopmentDetail)
	err = h.GetByFilter(tdcDetail, dbflex.Eq("TrainingCenterID", payload.TrainingCenterID))
	if err != nil {
		return nil, fmt.Errorf("error when get training detail: %s", err.Error())
	}

	participants := []hcmmodel.TrainingDevelopmentParticipant{}
	err = h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
	), &participants)
	if err != nil {
		return nil, fmt.Errorf("error when get training detail: %s", err.Error())
	}

	participantIds := make([]string, len(participants))
	employeeIds := make([]string, len(participants))
	jobIDs := make([]interface{}, len(participants))
	for i, p := range participants {
		participantIds[i] = p.ID
		employeeIds[i] = p.EmployeeID
		jobIDs[i] = p.ManpowerRequestID
	}

	employeeDetails := []bagongmodel.EmployeeDetail{}
	err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeID", employeeIds...),
	), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}
	mapEmployeeDetail := map[string]bagongmodel.EmployeeDetail{}
	positionIDs := make([]interface{}, len(employeeDetails))
	for i, emp := range employeeDetails {
		mapEmployeeDetail[emp.EmployeeID] = emp
		positionIDs[i] = emp.Position
	}

	employees := []tenantcoremodel.Employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", employeeIds...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}
	mapEmployee := map[string]tenantcoremodel.Employee{}
	siteIds := make([]string, len(employees)+1)
	siteIds[0] = tdcDetail.Site
	for i, emp := range employees {
		mapEmployee[emp.ID] = emp
		siteIds[i+1] = emp.Dimension.Get("Site")
	}

	sites := []bagongmodel.Site{}
	err = h.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", siteIds...),
	), &sites)
	if err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}
	mapSite := lo.Associate(sites, func(site bagongmodel.Site) (string, string) {
		return site.ID, site.Name
	})

	mapPosition := map[string]string{}
	if tdc.TrainingType == "General" {
		// get position
		positions := []tenantcoremodel.MasterData{}
		err = h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", positionIDs...),
			),
		), &positions)
		if err != nil {
			return nil, fmt.Errorf("error when get position: %s", err.Error())
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
			return nil, fmt.Errorf("error when get manpower request: %s", err.Error())
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
			return nil, fmt.Errorf("error when get position: %s", err.Error())
		}

		mapPosition = lo.Associate(positions, func(source tenantcoremodel.MasterData) (string, string) {
			return source.ID, source.Name
		})
	}

	mapPraticeScore := map[string]float64{}
	switch tdcDetail.AssessmentType {
	case "Assessment Staff":
		practices := []hcmmodel.TrainingDevelopmentPracticeTestStaff{}
		err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeTestStaff), dbflex.NewQueryParam().SetWhere(
			dbflex.In("ParticipantID", participantIds...),
		), &practices)
		if err != nil {
			return nil, fmt.Errorf("error when get staff practice test: %s", err.Error())
		}

		for _, prac := range practices {
			mapPraticeScore[prac.ParticipantID] = prac.TotalScore
		}
	case "Assessment Driver", "Assessment Mechanic":
		practices := []hcmmodel.TrainingDevelopmentPracticeScore{}
		err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeScore), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
			),
		), &practices)
		if err != nil {
			return nil, fmt.Errorf("error when get practice test: %s", err.Error())
		}

		for _, p := range practices {
			if tdcDetail.AssessmentType == "Assessment Driver" {
				// because there are 2 tests
				mapPraticeScore[p.EmployeeID] += p.FinalScore / 2
			} else {
				mapPraticeScore[p.EmployeeID] += p.FinalScore
			}
		}
	}

	practices := []hcmmodel.TrainingDevelopmentPracticeDuration{}
	err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeDuration), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	), &practices)
	if err != nil {
		return nil, fmt.Errorf("error when get practice duration: %s", err.Error())
	}

	mapPracticeDuration := map[string]float64{}
	for _, p := range practices {
		mapPracticeDuration[p.ParticipantID] += float64(p.Duration)
	}

	f := excelize.NewFile()
	defer func() {
		f.Close()
	}()
	sheetName := "Training Development"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{
		"NO", "PROVIDER", "NAMA TRAINING", "TIPE TRAINING", "NAMA PESERTA", "NIK",
		"Jabatan", "Site Peserta", "BATCH", "TANGGAL DIMULAI", "TANGGAL SELESAI",
		"TRAINER", "NILAI PRE TEST", "NILAI POST TEST", "NILAI PRAKTIKUM",
		"TOTAL JAM PELATIHAN", "NILAI TOTAL", "KETERANGAN", "DATE", "WEEK", "SITE", "LOKASI",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	provider := "Training Internal"
	if tdcDetail.ExternalTraining {
		provider = "Training External"
	}
	data := make([][]string, len(participantIds))
	for i, p := range participants {
		emp := mapEmployee[p.EmployeeID]
		empDetail := mapEmployeeDetail[p.EmployeeID]
		d := make([]string, 22)
		d[0] = codekit.ToString(i + 1)
		d[1] = provider
		d[2] = tdc.TrainingTitle
		d[3] = lo.Ternary(tdcDetail.OnlineTraining, "Online", "Offline")
		d[4] = emp.Name
		d[5] = empDetail.EmployeeNo
		d[6] = mapPosition[empDetail.Position]
		d[7] = mapSite[emp.Dimension.Get("Site")]
		d[8] = fmt.Sprintf("BATCH %d", tdcDetail.Batch)
		d[9] = tdcDetail.TrainingDateFrom.Format("01/02/2006")
		d[10] = tdcDetail.TrainingDateTo.Format("01/02/2006")
		d[11] = tdcDetail.TrainerName

		postScore := 0.0
		countPost := 0.0
		preScore := 0.0
		countPre := 0.0
		for _, d := range p.TestDetails {
			if d.Stage == hcmmodel.TDCTestTypePost {
				postScore += float64(d.Score)
				countPost++
			} else {
				preScore += float64(d.Score)
				countPre++
			}
		}

		finalPostScore := lo.Ternary(countPost == 0, 0, postScore/countPost)
		finalPreScore := lo.Ternary(countPre == 0, 0, preScore/countPre)
		// pre & post test
		d[12] = fmt.Sprintf("%.0f", finalPreScore)
		d[13] = fmt.Sprintf("%.0f", finalPostScore)

		practiceScore := 0.0
		if tdcDetail.AssessmentType == "Assessment Staff" {
			practiceScore = mapPraticeScore[p.ID]
		} else {
			practiceScore = mapPraticeScore[p.EmployeeID]
		}
		// practice score
		d[14] = fmt.Sprintf("%.0f", practiceScore)

		// practice duration
		d[15] = fmt.Sprintf("%.0f", mapPracticeDuration[p.ID])

		// final score
		finalScore := (finalPostScore + practiceScore) / 2
		d[16] = fmt.Sprintf("%.0f", finalScore)

		if finalScore > 74.9 {
			d[17] = "LULUS"
		} else {
			d[17] = "GAGAL"
		}

		d[18] = tdcDetail.TrainingDateTo.Format("01")

		_, week := tdcDetail.TrainingDateTo.ISOWeek()
		d[19] = fmt.Sprintf("%d", week)

		d[20] = mapSite[tdcDetail.Site]
		d[21] = tdcDetail.Location

		data[i] = d
	}

	for i, row := range data {
		for j, r := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			f.SetCellValue(sheetName, cell, r)
		}
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="example.xlsx"`)

	if err := f.Write(w); err != nil {
		return nil, fmt.Errorf("error when create file: %s", err.Error())
	}

	return nil, nil
}
