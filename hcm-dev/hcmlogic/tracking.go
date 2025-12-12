package hcmlogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/kmsg"
	"github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrackingHandler struct{}

type TrackingApplyApplicantRequest struct {
	EmployeeID string
	JobID      []string
}

func (m *TrackingHandler) ApplyApplicant(ctx *kaos.Context, payload *TrackingApplyApplicantRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	for _, j := range payload.JobID {
		applicant := new(hcmmodel.Screening)
		applicant.CandidateID = userID
		applicant.JobVacancyID = j
		applicant.Status = "Not Selected"

		err := h.Save(applicant)
		if err != nil {
			return nil, fmt.Errorf("error when save applicant: %s", err.Error())
		}
	}

	return "success", nil
}

type TrackingGetApplicantCountRequest struct {
	JobID string
}

type TrackingGetApplicantCountResponse struct {
	Stage string
	Count int
}

func (m *TrackingHandler) GetApplicantCount(ctx *kaos.Context, payload *TrackingGetApplicantCountRequest) ([]TrackingGetApplicantCountResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	mapStage := map[string]int{
		"Screening":          0,
		"PshycologicalTest":  0,
		"Interview":          0,
		"TechnicalInterview": 0,
		"MCU":                0,
		"Training":           0,
		"OLPlotting":         0,
		"PKWTT":              0,
		"OnBoarding":         0,
	}

	for k := range mapStage {
		mdl := m.getStage(k)
		count, err := h.Count(mdl, dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("JobVacancyID", payload.JobID),
				dbflex.Eq("Status", "Passed"),
			),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get applicant count: %s", err.Error())
		}

		mapStage[k] = count
	}

	result := make([]TrackingGetApplicantCountResponse, len(mapStage))
	i := 0
	for k, v := range mapStage {
		result[i] = TrackingGetApplicantCountResponse{
			Stage: k,
			Count: v,
		}
		i++
	}

	return result, nil
}

func (m *TrackingHandler) getStage(stage string) orm.DataModel {
	var mdl orm.DataModel
	switch stage {
	case "Screening":
		mdl = new(hcmmodel.Screening)
	case "PshycologicalTest":
		mdl = new(hcmmodel.PsychologicalTest)
	case "Interview":
		mdl = new(hcmmodel.Interview)
	case "OLPlotting":
		mdl = new(hcmmodel.OLPlotting)
	case "MCU":
		mdl = new(hcmmodel.MCU)
	case "TechnicalInterview":
		mdl = new(hcmmodel.TechnicalInterview)
	case "PKWTT":
		mdl = new(hcmmodel.PKWTT)
	case "Training":
		mdl = new(hcmmodel.Training)
	case "OnBoarding":
		mdl = new(hcmmodel.OnBoarding)
	}

	return mdl
}

type TrackingGetApplicantRequest struct {
	Where struct {
		// only screening
		AgeFrom  int
		AgeTo    int
		Domicile []string

		Status string
		Stage  string
		JobID  string
	}
	Take int
	Skip int
}

func (m *TrackingHandler) GetApplicant(ctx *kaos.Context, payload *TrackingGetApplicantRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	empFilters := []*dbflex.Filter{}
	if payload.Where.AgeFrom != 0 {
		empFilters = append(empFilters, dbflex.Gte("Age", payload.Where.AgeFrom))
	}

	if payload.Where.AgeTo != 0 {
		empFilters = append(empFilters, dbflex.Lte("Age", payload.Where.AgeTo))
	}

	if len(payload.Where.Domicile) > 0 {
		empFilters = append(empFilters, dbflex.In("Domicile", payload.Where.Domicile...))
	}

	query := dbflex.NewQueryParam()
	if len(empFilters) > 0 {
		query.SetWhere(dbflex.And(empFilters...))
	}

	type employeeDetail struct {
		EmployeeID string
		Age        string
	}

	employeeDetails := make([]employeeDetail, 0)
	err := h.Gets(new(bagongmodel.EmployeeDetail), query.SetSelect("EmployeeID", "Age"), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	mapEmployeeDetail := map[string]string{}
	ids := make([]string, len(employeeDetails))
	for i, emp := range employeeDetails {
		ids[i] = emp.EmployeeID
		mapEmployeeDetail[emp.EmployeeID] = emp.Age
	}

	type employee struct {
		ID   string `bson:"_id"`
		Name string
	}

	// get employees
	employees := make([]employee, 0)
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", ids...),
	).SetSelect("_id", "Name"), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(detail employee) (string, string) {
		return detail.ID, detail.Name
	})

	stageFilters := []*dbflex.Filter{
		dbflex.Eq("JobVacancyID", payload.Where.JobID),
		dbflex.In("CandidateID", ids...),
	}
	mdl := m.getStage(payload.Where.Stage)
	applicants := make([]codekit.M, 0)
	err = h.Gets(mdl, dbflex.NewQueryParam().SetWhere(
		dbflex.And(stageFilters...),
	).SetSkip(payload.Skip).SetTake(payload.Take), &applicants)
	if err != nil {
		return nil, fmt.Errorf("error when get applicant: %s", err.Error())
	}

	for i := range applicants {
		id := applicants[i]["CandidateID"].(string)
		applicants[i]["Age"] = mapEmployeeDetail[id]
		applicants[i]["Name"] = mapEmployee[id]
	}

	count, err := h.Count(mdl, dbflex.NewQueryParam().SetWhere(
		dbflex.And(stageFilters...),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get applicant count: %s", err.Error())
	}

	return codekit.M{"count": count, "data": applicants}, nil
}

type TrackingMarkStatusRequest struct {
	JobID   string
	StageID []string
	Stage   string
	Status  string
}

func (m *TrackingHandler) MarkStatus(ctx *kaos.Context, payload *TrackingMarkStatusRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	current := m.getStage(payload.Stage)
	applicants := make([]codekit.M, 0)
	err := h.Gets(current, dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", payload.StageID...),
	).SetSelect("_id", "Status", "CandidateID", "JobVacancyID"), &applicants)
	if err != nil {
		return nil, fmt.Errorf("error when get applicant: %s", err.Error())
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if coID == "DEMO" || coID == "" {
		return nil, errors.New("missing: Company, please relogin")
	}

	type journal struct {
		ID               string `bson:"_id"`
		PostingProfileID string
	}
	// get journal type for plotting
	journalType := &journal{}
	if payload.Stage == "Training" {
		journals := []journal{}
		err := h.Gets(new(hcmmodel.JournalType), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("TransactionType", "OL & Plotting"),
		).SetSelect("_id", "PostingProfileID"), &journals)
		if err != nil {
			return nil, fmt.Errorf("error when get journal type: %s", err.Error())
		}

		if len(journals) == 1 {
			journalType = &journals[0]
		}
	}

	isPassedPKWTT := false
	isCreateProbation := false
	employeeID := make([]string, len(applicants))
	for i := range applicants {
		employeeID[i] = applicants[i]["CandidateID"].(string)

		field := ""
		if payload.Stage != "OLPlotting" {
			applicants[i]["Status"] = payload.Status
			field = "Status"
		} else {
			applicants[i]["StageStatus"] = payload.Status
			field = "StageStatus"
		}

		err = h.UpdateAny(current.TableName(), dbflex.Eq("_id", applicants[i]["_id"]), applicants[i], field)
		if err != nil {
			return nil, fmt.Errorf("error when update stage: %s", err.Error())
		}

		if payload.Status == "Passed" {
			var nextStage orm.DataModel
			switch payload.Stage {
			case "Screening":
				nextStage = new(hcmmodel.PsychologicalTest)
			case "PshycologicalTest":
				nextStage = new(hcmmodel.Interview)
			case "Interview":
				nextStage = new(hcmmodel.TechnicalInterview)
			case "TechnicalInterview":
				nextStage = new(hcmmodel.MCU)
			case "MCU":
				nextStage = new(hcmmodel.Training)
				reflector.From(nextStage).Set("TrainingStatus", "Candidate has not been assigned to any Training").Flush()
			case "Training":
				nextStage = new(hcmmodel.OLPlotting)
				reflector.From(nextStage).Set("JournalTypeID", journalType.ID).Flush()
				reflector.From(nextStage).Set("PostingProfileID", journalType.PostingProfileID).Flush()
				reflector.From(nextStage).Set("CompanyID", coID).Flush()
			case "OLPlotting":
				nextStage = new(hcmmodel.PKWTT)
			case "PKWTT":
				nextStage = new(hcmmodel.OnBoarding)
				isPassedPKWTT = true
			case "OnBoarding":
				isCreateProbation = true
			}

			if !isCreateProbation {
				id := primitive.NewObjectID().Hex()
				reflector.From(nextStage).Set("ID", id).Flush()
				reflector.From(nextStage).Set("CandidateID", applicants[i]["CandidateID"]).Flush()
				reflector.From(nextStage).Set("JobVacancyID", applicants[i]["JobVacancyID"]).Flush()

				if payload.Stage != "Training" {
					reflector.From(nextStage).Set("Status", "Not Selected").Flush()
				} else {
					// only for plotting, set different field
					// because status field used for posting approval
					reflector.From(nextStage).Set("StageStatus", "Not Selected").Flush()
				}

				err = h.Save(nextStage)
				if err != nil {
					return nil, fmt.Errorf("error when save next stage: %s", err.Error())
				}
			}
		}
	}

	if isPassedPKWTT {
		err = m.postSuccessMarkPKWTT(ctx, h, &PostSuccessMarkPKWTTPayload{EmployeeID: employeeID, JobID: payload.JobID})
		if err != nil {
			return nil, fmt.Errorf("error when updating employee : %s", err.Error())
		}
	}

	if isCreateProbation {
		err = m.createProbation(ctx, h, &PostSuccessMarkPKWTTPayload{EmployeeID: employeeID, JobID: payload.JobID})
		if err != nil {
			return nil, fmt.Errorf("error when create probation : %s", err.Error())
		}
	}

	return "success", nil
}

type TrackingSendPsychologicalTestRequest struct {
	JobID       string
	CandidateID []string
}

func (m *TrackingHandler) SendPsychologicalTest(ctx *kaos.Context, payload *TrackingSendPsychologicalTestRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	sendEmailTopic := Config.TopicSendMessageTemplate
	if sendEmailTopic == "" {
		return "", errors.New("missing_config: topic_send_message_template")
	}

	// get schedules
	schedules := make([]hcmmodel.TestSchedule, 0)
	err := h.Gets(new(hcmmodel.TestSchedule), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TestID", payload.JobID),
			dbflex.Gte("DateTo", time.Now()),
		),
	), &schedules)
	if err != nil {
		return nil, fmt.Errorf("error when get schedule: %s", err.Error())
	}

	templateIDs := lo.Map(schedules, func(t hcmmodel.TestSchedule, index int) string {
		return t.TemplateID
	})

	type template struct {
		ID   string `bson:"_id"`
		Name string
	}

	// get templates
	templates := make([]template, 0)
	err = h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", templateIDs...),
	).SetSelect("_id", "Name"), &templates)
	if err != nil {
		return nil, fmt.Errorf("error when get template: %s", err.Error())
	}

	mapTemplate := lo.Associate(templates, func(t template) (string, string) {
		return t.ID, t.Name
	})

	dates := make([]string, len(schedules))
	for i, test := range schedules {
		dates[i] = fmt.Sprintf("%d. %s : %s - %s", i+1, mapTemplate[test.TemplateID], test.DateFrom.Format("02 January 2006"), test.DateTo.Format("02 January 2006"))
	}

	// get employees
	employees := make([]tenantcoremodel.Employee, 0)
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", payload.CandidateID...)), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	emails := make([]string, len(employees))
	for i, d := range employees {
		emails[i] = d.Email
	}

	kind := "psychological-test"
	msg := kmsg.Message{
		Kind: kind,
		// To:     "hrdho@bagongbis.com",
		To:     "avinda@kanosolution.com",
		Method: "SMTP",
		Bcc:    emails,
	}

	sendMessageRequest := ksmsg.SendTemplateRequest{
		TemplateName: kind,
		Message:      &msg,
		LanguageID:   "en-us",
		Data:         codekit.M{"Dates": strings.Join(dates, "<br>")},
	}

	err = Config.EventHub.Publish(sendEmailTopic, sendMessageRequest, nil, &kaos.PublishOpts{Headers: codekit.M{}})
	if err != nil {
		return nil, fmt.Errorf("error when send email: %s", err.Error())
	}

	return nil, nil
}

type TrackingHandlerSaveMCURequest struct {
	CandidateID []string
	JobID       string
}

func (m *TrackingHandler) SaveMCU(ctx *kaos.Context, payload *TrackingHandlerSaveMCURequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get employees
	employees := make([]tenantcoremodel.Employee, 0)
	err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", payload.CandidateID...)), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	// get employee detail
	employeeDetails := make([]bagongmodel.EmployeeDetail, 0)
	err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(dbflex.In("EmployeeID", payload.CandidateID...)), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	mapEmployeeDetail := lo.Associate(employeeDetails, func(detail bagongmodel.EmployeeDetail) (string, bagongmodel.EmployeeDetail) {
		return detail.ID, detail
	})

	mcus := make([]hcmmodel.MCU, 0)
	err = h.Gets(new(hcmmodel.MCU), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("CandidateID", payload.CandidateID...),
			dbflex.Eq("JobVacancyID", payload.JobID),
		),
	), &mcus)
	if err != nil {
		return nil, fmt.Errorf("error when get mcu: %s", err.Error())
	}

	mapMcu := lo.Associate(mcus, func(detail hcmmodel.MCU) (string, hcmmodel.MCU) {
		return detail.CandidateID, detail
	})

	for _, c := range payload.CandidateID {
		if v, ok := mapMcu[c]; ok {
			mcuTransaction := shemodel.MCUTransaction{
				Category: "Candidate",
				Purpose:  "REKRUTMEN",
				Name:     mapEmployee[c],
				Status:   "DRAFT",
			}

			if v.MCUTransactionID == "" {
				tenantcorelogic.MWPreAssignSequenceNo("MCUTransaction", false, "_id")(ctx, &mcuTransaction)
			} else {
				mcuTransaction.ID = v.MCUTransactionID
			}

			if emp, ok := mapEmployeeDetail[c]; ok {
				mcuTransaction.Gender = emp.Gender
				mcuTransaction.Age = codekit.ToInt(emp.Age, codekit.RoundingAuto)
			}

			err = h.Save(&mcuTransaction)
			if err != nil {
				return nil, fmt.Errorf("error when save mcu transaction: %s", err.Error())
			}

			v.MCUTransactionID = mcuTransaction.ID
			err = h.Save(&v)
			if err != nil {
				return nil, fmt.Errorf("error when save mcu transaction: %s", err.Error())
			}
		}
	}

	return "success", nil
}

type TrackingHandlerCreateTrainingRequest struct {
	CandidateID []string
	JobID       string
}

func (m *TrackingHandler) CreateTraining(ctx *kaos.Context, payload *TrackingHandlerCreateTrainingRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	job := new(hcmmodel.ManpowerRequest)
	err := h.GetByID(job, payload.JobID)
	if err != nil {
		return nil, fmt.Errorf("error when get manpower request: %s", err.Error())
	}

	master := new(tenantcoremodel.MasterData)
	err = h.GetByID(master, job.JobVacancyTitle)
	if err != nil {
		return nil, fmt.Errorf("error when get manpower request: %s", err.Error())
	}

	tdc := hcmmodel.TrainingDevelopment{
		ID:            primitive.NewObjectID().Hex(),
		TrainingTitle: fmt.Sprintf("REKRUTMEN - TRAINING CANDIDATE - %s", master.Name),
	}
	err = h.Save(&tdc)
	if err != nil {
		return nil, fmt.Errorf("error when save training development: %s", err.Error())
	}

	for _, id := range payload.CandidateID {
		participant := hcmmodel.TrainingDevelopmentParticipant{
			TrainingCenterID:  tdc.ID,
			EmployeeID:        id,
			ManpowerRequestID: payload.JobID,
		}

		err = h.Save(&participant)
		if err != nil {
			return nil, fmt.Errorf("error when save training development participant: %s", err.Error())
		}
	}

	return "success", nil
}

type PostSuccessMarkPKWTTPayload struct {
	EmployeeID []string
	JobID      string
}

func (m *TrackingHandler) postSuccessMarkPKWTT(ctx *kaos.Context, h *datahub.Hub, payload *PostSuccessMarkPKWTTPayload) error {
	// check if candidate count fulfill number employees required
	manpower := new(hcmmodel.ManpowerRequest)
	err := h.GetByID(manpower, payload.JobID)
	if err != nil {
		return fmt.Errorf("err when get manpower request: %s", err.Error())
	}

	// get pkwtt
	pkwtts := make([]hcmmodel.PKWTT, 0)
	err = h.Gets(new(hcmmodel.PKWTT), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("CandidateID", payload.EmployeeID...),
			dbflex.Eq("JobVacancyID", payload.JobID),
			dbflex.Eq("Status", "Passed"),
		),
	), &pkwtts)
	if err != nil {
		return fmt.Errorf("error when get pkwtt: %s", err.Error())
	}

	// check if fulfill number employees required
	// first condition for manpower request with number employees required more than 1, additional number field must more than 0
	// second condition for manpower request with replacement for 1 employee, additional number field must be 0
	if (manpower.AdditionalNumber != 0 && manpower.AdditionalNumber <= len(pkwtts)) ||
		(manpower.AdditionalNumber == 0 && len(pkwtts) > 0) {
		manpower.IsClose = true
		err = h.Save(manpower)
		if err != nil {
			return fmt.Errorf("error when save manpower request: %s", err.Error())
		}
	}

	// get employees
	employees := make([]tenantcoremodel.Employee, 0)
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", payload.EmployeeID...)), &employees)
	if err != nil {
		return fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(detail tenantcoremodel.Employee) (string, tenantcoremodel.Employee) {
		return detail.ID, detail
	})

	// get employee detail
	employeeDetails := make([]bagongmodel.EmployeeDetail, 0)
	err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(dbflex.In("EmployeeID", payload.EmployeeID...)), &employeeDetails)
	if err != nil {
		return fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	mapEmployeeDetail := lo.Associate(employeeDetails, func(detail bagongmodel.EmployeeDetail) (string, bagongmodel.EmployeeDetail) {
		return detail.ID, detail
	})

	for _, pkwtt := range pkwtts {
		if v, ok := mapEmployee[pkwtt.CandidateID]; ok {
			v.EmploymentType = "PROBATION"
			v.JoinDate = pkwtt.JoinedDate

			err = h.Save(&v)
			if err != nil {
				return fmt.Errorf("error when save employee: %s", err.Error())
			}
		}

		if v, ok := mapEmployeeDetail[pkwtt.CandidateID]; ok {
			tenantcorelogic.MWPreAssignSequenceNo("EmployeeNo", false, "EmployeeNo")(ctx, &v)
			err = h.Save(&v)
			if err != nil {
				return fmt.Errorf("error when save employee detail: %s", err.Error())
			}
		}
	}

	return nil
}

func (m *TrackingHandler) createProbation(ctx *kaos.Context, h *datahub.Hub, payload *PostSuccessMarkPKWTTPayload) error {
	// get pkwtt
	pwktts := make([]hcmmodel.PKWTT, 0)
	err := h.Gets(new(hcmmodel.PKWTT), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("CandidateID", payload.EmployeeID...),
			dbflex.Eq("JobVacancyID", payload.JobID),
		),
	), &pwktts)
	if err != nil {
		return fmt.Errorf("error when get pkwtt: %s", err.Error())
	}

	// get man power request
	job := new(hcmmodel.ManpowerRequest)
	err = h.GetByID(job, payload.JobID)
	if err != nil {
		return fmt.Errorf("error when get manpower request: %s", err.Error())
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	for _, pkwtt := range pwktts {
		prob := hcmmodel.Contract{
			EmployeeID: pkwtt.CandidateID,
			JobID:      payload.JobID,
			JobTitle:   job.JobVacancyTitle,
			JoinedDate: pkwtt.JoinedDate,
			// JournalTypeID:       "JTContract",
			CompanyID: coID,
			// PostingProfileID:    "PPHCGS",
			ExpiredContractDate: pkwtt.ExpiredContractDate,
			Status:              ficomodel.JournalStatusDraft,
		}

		err = h.Save(&prob)
		if err != nil {
			return fmt.Errorf("error when get contract: %s", err.Error())
		}
	}

	return nil
}

type TrackingPlottingRequest struct {
	ID   string
	Site string
}

func (m *TrackingHandler) Plotting(ctx *kaos.Context, payload []TrackingPlottingRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	for _, p := range payload {
		mdl := new(hcmmodel.OLPlotting)
		err := h.GetByID(mdl, p.ID)
		if err != nil {
			return nil, fmt.Errorf("error when get plotting: %s", err.Error())
		}
		mdl.Plotting = p.Site

		err = h.Save(mdl)
		if err != nil {
			return nil, fmt.Errorf("error when save plotting: %s", err.Error())
		}
	}

	return "success", nil
}
