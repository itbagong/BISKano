package hcmlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TestScheduleHandler struct {
}

type FindCandidateTestResponse struct {
	TemplateID       string
	TestName         string
	TestID           string
	TestScheduleType hcmmodel.TestScheduleType
	Instruction      shemodel.Instructions
}

type FindCandidateTestRequest struct {
	Date time.Time
}

func (m *TestScheduleHandler) FindCandidateTest(ctx *kaos.Context, payload *FindCandidateTestRequest) ([]FindCandidateTestResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		return nil, errors.New("missing: User, please relogin")
	}

	testIDs := []string{}
	templateIDs := []string{}
	// get training participant
	trainingTests := []hcmmodel.TrainingDevelopmentParticipant{}
	err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("EmployeeID", userID),
	), &trainingTests)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	for _, t := range trainingTests {
		testIDs = append(testIDs, t.TrainingCenterID)
		for _, s := range t.TestDetails {
			templateIDs = append(templateIDs, s.TemplateID)
		}
	}

	// get psychological test
	psychologicalTests := []hcmmodel.PsychologicalTest{}
	err = h.Gets(new(hcmmodel.PsychologicalTest), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("CandidateID", userID),
	), &psychologicalTests)
	if err != nil {
		return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
	}

	for _, t := range psychologicalTests {
		testIDs = append(testIDs, t.JobVacancyID)
		for _, s := range t.Details {
			templateIDs = append(templateIDs, s.TemplateID)
		}
	}

	// get talent development test
	talentDevelopment := []hcmmodel.TalentDevelopment{}
	err = h.Gets(new(hcmmodel.TalentDevelopment), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("EmployeeID", userID),
	), &talentDevelopment)
	if err != nil {
		return nil, fmt.Errorf("error when get talent development: %s", err.Error())
	}

	ids := lo.Map(talentDevelopment, func(td hcmmodel.TalentDevelopment, index int) string {
		return td.ID
	})

	// get talent development assesment test
	tdAssesment := []hcmmodel.TalentDevelopmentAssesment{}
	err = h.Gets(new(hcmmodel.TalentDevelopmentAssesment), dbflex.NewQueryParam().SetWhere(
		dbflex.In("TalentDevelopmentID", ids...),
	), &tdAssesment)
	if err != nil {
		return nil, fmt.Errorf("error when get talent development assesment: %s", err.Error())
	}

	for _, t := range tdAssesment {
		testIDs = append(testIDs, t.ID)
		for _, s := range t.PsychoTests {
			templateIDs = append(templateIDs, s.TemplateID)
		}
	}

	type schedule struct {
		TestID     string
		TemplateID string
	}

	now := time.Now()
	// get only active schedule
	schedules := []schedule{}
	err = h.Gets(new(hcmmodel.TestSchedule), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("TestID", testIDs...),
			dbflex.In("TemplateID", templateIDs...),
			dbflex.Gte("DateTo", now),
			dbflex.Lte("DateFrom", now),
		),
	).SetSelect("TestID", "TemplateID"), &schedules)
	if err != nil {
		return nil, fmt.Errorf("error when get schelude: %s", err.Error())
	}

	mapSchedule := lo.Associate(schedules, func(s schedule) (string, string) {
		return s.TestID + s.TemplateID, s.TemplateID
	})

	templateIDs = []string{}
	result := []FindCandidateTestResponse{}
	for _, t := range trainingTests {
		for _, d := range t.TestDetails {
			if d.Status == hcmmodel.TrainingDevelopmentParticipantStatusOpen {
				if _, ok := mapSchedule[t.TrainingCenterID+d.TemplateID]; ok {
					templateIDs = append(templateIDs, d.TemplateID)
					result = append(result, FindCandidateTestResponse{
						TemplateID:       d.TemplateID,
						TestName:         d.TemplateName,
						TestID:           t.TrainingCenterID,
						TestScheduleType: hcmmodel.TestScheduleTypeTDC,
					})
				}
			}
		}
	}

	for _, t := range psychologicalTests {
		for _, d := range t.Details {
			if d.Status == hcmmodel.PsychologicalTestDetailStatusOpen {
				if _, ok := mapSchedule[t.JobVacancyID+d.TemplateID]; ok {
					templateIDs = append(templateIDs, d.TemplateID)
					result = append(result, FindCandidateTestResponse{
						TemplateID:       d.TemplateID,
						TestName:         d.TemplateName,
						TestID:           t.JobVacancyID,
						TestScheduleType: hcmmodel.TestScheduleTypePsychological,
					})
				}
			}
		}
	}

	for _, t := range tdAssesment {
		for _, d := range t.PsychoTests {
			if d.Status == hcmmodel.PsychologicalTestDetailStatusOpen {
				if _, ok := mapSchedule[t.ID+d.TemplateID]; ok {
					templateIDs = append(templateIDs, d.TemplateID)
					result = append(result, FindCandidateTestResponse{
						TemplateID:       d.TemplateID,
						TestName:         d.TemplateName,
						TestID:           t.ID,
						TestScheduleType: hcmmodel.TestScheduleTypeTD,
					})
				}
			}
		}
	}

	// get template id
	templates := []shemodel.MCUItemTemplate{}
	err = h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", templateIDs...),
	), &templates)
	if err != nil {
		return nil, fmt.Errorf("error when get question template: %s", err.Error())
	}

	mapInstruction := lo.Associate(templates, func(m shemodel.MCUItemTemplate) (string, shemodel.Instructions) {
		return m.ID, m.Instruction
	})

	for i := range result {
		result[i].Instruction = mapInstruction[result[i].TemplateID]
	}

	return result, nil
}

type GetQuestionsRequest struct {
	TemplateID       string
	TestID           string
	TestScheduleType hcmmodel.TestScheduleType
	Take             int
	Skip             int
}

type GetQuestionsResponse struct {
	Lines []GetQuestionsResponseLine
}

type GetQuestionsResponseLine struct {
	ID                     string
	Number                 string
	Type                   shemodel.MCUTemplateType
	Unit                   string
	Condition              []interface{}
	Description            string
	IsGender               bool
	Range                  []shemodel.MCURange
	Parent                 string
	IsSelected             bool
	AssessmentTypeIsNumber bool
	QuestionTypeIsMost     bool
	AnswerValue            int
	Attachment             string
	Result                 string
	Note                   string
	EmployeeAnswerID       string
	EmployeeAnswer         string
}

type MCULineCondition struct {
	ID      string
	Name    string
	Vnumber int
	Letter  string
}

func (m *TestScheduleHandler) GetQuestions(ctx *kaos.Context, payload *GetQuestionsRequest) (interface{}, error) {
	return m.getQuestionDetail(ctx, payload, false)
}

func (m *TestScheduleHandler) GetEmployeeAnswers(ctx *kaos.Context, payload *GetQuestionsRequest) (interface{}, error) {
	return m.getQuestionDetail(ctx, payload, true)
}

func (m *TestScheduleHandler) getQuestionDetail(ctx *kaos.Context, payload *GetQuestionsRequest, isShowAnswer bool) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	question := new(shemodel.MCUItemTemplate)
	err := h.GetByID(question, payload.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("err when get template test: %s", err.Error())
	}

	count := len(question.Lines)
	lines := []shemodel.MCUItemTemplateLine{}
	if payload.Take+payload.Skip > count {
		lines = question.Lines[payload.Skip:]
	} else {
		lines = question.Lines[payload.Skip:(payload.Take + payload.Skip)]
	}

	result := make([]GetQuestionsResponseLine, len(lines))
	if len(lines) > 0 {
		questionIDs := make([]string, len(lines))
		for i, l := range lines {
			questionIDs[i] = l.ID

			res := GetQuestionsResponseLine{
				ID:                     l.ID,
				Number:                 l.Number,
				Type:                   l.Type,
				Unit:                   l.Unit,
				Description:            l.Description,
				IsGender:               l.IsGender,
				Range:                  l.Range,
				Parent:                 l.Parent,
				IsSelected:             l.IsSelected,
				AssessmentTypeIsNumber: l.AssessmentTypeIsNumber,
				QuestionTypeIsMost:     l.QuestionTypeIsMost,
				AnswerValue:            l.AnswerValue,
				Attachment:             l.Attachment,
				Result:                 l.Result,
				Note:                   l.Note,
				Condition:              make([]interface{}, len(l.Condition)),
			}
			for i, c := range l.Condition {
				if isShowAnswer {
					res.Condition[i] = c
				} else {
					res.Condition[i] = MCULineCondition{
						ID:      c.ID,
						Name:    c.Name,
						Vnumber: c.Vnumber,
						Letter:  c.Letter,
					}
				}
			}

			result[i] = res
		}

		answers := []codekit.M{}
		if payload.TestScheduleType == hcmmodel.TestScheduleTypeTDC {
			err := h.Gets(new(hcmmodel.TrainingAnswerHistory), dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.In("QuestionID", questionIDs...),
					dbflex.Eq("TemplateTestID", payload.TemplateID),
					dbflex.Eq("TrainingCenterID", payload.TestID),
				),
			).SetSelect("QuestionID", "AnswerID", "Answer"), &answers)
			if err != nil {
				return nil, fmt.Errorf("error when get answer history: %s", err.Error())
			}
		} else {
			err := h.Gets(new(hcmmodel.PsychologicalAnswerHistory), dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.In("QuestionID", questionIDs...),
					dbflex.Eq("TemplateTestID", payload.TemplateID),
					dbflex.Eq("TestID", payload.TestID),
				),
			).SetSelect("QuestionID", "AnswerID", "Answer"), &answers)
			if err != nil {
				return nil, fmt.Errorf("error when get answer history: %s", err.Error())
			}
		}

		mapAnswer := lo.Associate(answers, func(m codekit.M) (string, codekit.M) {
			return m["QuestionID"].(string), m
		})

		for i := range result {
			if v, ok := mapAnswer[result[i].ID]; ok {
				result[i].EmployeeAnswerID = v["AnswerID"].(string)
				result[i].EmployeeAnswer = v["Answer"].(string)
			}
		}
	}

	return codekit.M{"data": result, "count": count}, nil
}

func (m *TestScheduleHandler) SavePsychologicalSchedule(ctx *kaos.Context, payload []hcmmodel.TestSchedule) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	// get company
	company, err := datahub.GetByParm(h, new(tenantcoremodel.Company), dbflex.NewQueryParam().
		SetWhere(dbflex.Eqs("_id", coID)))
	if err != nil {
		return nil, fmt.Errorf("error when get company: %s", err)
	}

	loc, err := time.LoadLocation(company.LocationCode)
	if err != nil {
		return nil, fmt.Errorf("error convert company location: %s", err)
	}

	if len(payload) > 0 {
		details := make([]hcmmodel.PsychologicalTestDetail, len(payload))
		templateIDs := make([]string, len(payload))
		for i := range payload {
			start := time.Date(payload[i].DateFrom.Year(), payload[i].DateFrom.Month(), payload[i].DateFrom.Day(), 0, 0, 0, 0, loc)
			end := time.Date(payload[i].DateTo.Year(), payload[i].DateTo.Month(), payload[i].DateTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1).Add(-1 * time.Second)

			payload[i].DateFrom = &start
			payload[i].DateTo = &end

			templateIDs[i] = payload[i].TemplateID
			if payload[i].ID == "" {
				payload[i].ID = primitive.NewObjectID().Hex()
			}

			details[i] = hcmmodel.PsychologicalTestDetail{
				TemplateID: payload[i].TemplateID,
				Status:     hcmmodel.PsychologicalTestDetailStatusOpen,
			}

			err := h.Save(&payload[i])
			if err != nil {
				return nil, fmt.Errorf("error when save schedule: %s", err.Error())
			}
		}

		templates := []shemodel.MCUItemTemplate{}
		err := h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", templateIDs...),
		), &templates)
		if err != nil {
			return nil, fmt.Errorf("error when get template: %s", err.Error())
		}

		mapTemplate := lo.Associate(templates, func(m shemodel.MCUItemTemplate) (string, string) {
			return m.ID, m.Name
		})

		for i := range details {
			details[i].TemplateName = mapTemplate[details[i].TemplateID]
		}

		tests := []hcmmodel.PsychologicalTest{}
		err = h.Gets(new(hcmmodel.PsychologicalTest), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("JobVacancyID", payload[0].TestID),
			),
		), &tests)
		if err != nil {
			return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
		}

		for _, d := range tests {
			d.Details = details
			err := h.Save(&d)
			if err != nil {
				return nil, fmt.Errorf("error when save psychological test: %s", err.Error())
			}
		}
	}

	return payload, nil
}

func (m *TestScheduleHandler) SaveTrainingSchedule(ctx *kaos.Context, payload []hcmmodel.TestSchedule) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	// get company
	company, err := datahub.GetByParm(h, new(tenantcoremodel.Company), dbflex.NewQueryParam().
		SetWhere(dbflex.Eqs("_id", coID)))
	if err != nil {
		return nil, fmt.Errorf("error when get company: %s", err)
	}

	loc, err := time.LoadLocation(company.LocationCode)
	if err != nil {
		return nil, fmt.Errorf("error convert company location: %s", err)
	}

	if len(payload) > 0 {
		details := make([]hcmmodel.TrainingDevelopmentParticipantDetail, len(payload))
		templateIDs := make([]string, len(payload))
		for i := range payload {
			start := time.Date(payload[i].DateFrom.Year(), payload[i].DateFrom.Month(), payload[i].DateFrom.Day(), 0, 0, 0, 0, loc)
			end := time.Date(payload[i].DateTo.Year(), payload[i].DateTo.Month(), payload[i].DateTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1).Add(-1 * time.Second)

			payload[i].DateFrom = &start
			payload[i].DateTo = &end

			templateIDs[i] = payload[i].TemplateID

			if payload[i].ID == "" {
				payload[i].ID = primitive.NewObjectID().Hex()
			}

			details[i] = hcmmodel.TrainingDevelopmentParticipantDetail{
				Stage:      payload[i].TestType,
				TemplateID: payload[i].TemplateID,
				Status:     hcmmodel.TrainingDevelopmentParticipantStatusOpen,
			}

			err := h.Save(&payload[i])
			if err != nil {
				return nil, fmt.Errorf("error when save schedule: %s", err.Error())
			}
		}

		templates := []shemodel.MCUItemTemplate{}
		err := h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", templateIDs...),
		), &templates)
		if err != nil {
			return nil, fmt.Errorf("error when get template: %s", err.Error())
		}

		mapTemplate := lo.Associate(templates, func(m shemodel.MCUItemTemplate) (string, string) {
			return m.ID, m.Name
		})

		for i := range details {
			details[i].TemplateName = mapTemplate[details[i].TemplateID]
		}

		// get participant
		participants := []hcmmodel.TrainingDevelopmentParticipant{}
		err = h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("TrainingCenterID", payload[0].TestID),
		), &participants)
		if err != nil {
			return nil, fmt.Errorf("error when get participant: %s", err.Error())
		}

		for _, p := range participants {
			p.TestDetails = details

			err := h.Save(&p)
			if err != nil {
				return nil, fmt.Errorf("error when save participant: %s", err.Error())
			}
		}
	}

	return payload, nil
}

func (m *TestScheduleHandler) SaveTalentDevelopmentSchedule(ctx *kaos.Context, payload []hcmmodel.TestSchedule) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	// get company
	company, err := datahub.GetByParm(h, new(tenantcoremodel.Company), dbflex.NewQueryParam().
		SetWhere(dbflex.Eqs("_id", coID)))
	if err != nil {
		return nil, fmt.Errorf("error when get company: %s", err)
	}

	loc, err := time.LoadLocation(company.LocationCode)
	if err != nil {
		return nil, fmt.Errorf("error convert company location: %s", err)
	}

	if len(payload) > 0 {
		details := make([]hcmmodel.PsychologicalTestDetail, len(payload))
		templateIDs := make([]string, len(payload))
		for i := range payload {
			start := time.Date(payload[i].DateFrom.Year(), payload[i].DateFrom.Month(), payload[i].DateFrom.Day(), 0, 0, 0, 0, loc)
			end := time.Date(payload[i].DateTo.Year(), payload[i].DateTo.Month(), payload[i].DateTo.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1).Add(-1 * time.Second)

			payload[i].DateFrom = &start
			payload[i].DateTo = &end

			templateIDs[i] = payload[i].TemplateID
			if payload[i].ID == "" {
				payload[i].ID = primitive.NewObjectID().Hex()
			}

			details[i] = hcmmodel.PsychologicalTestDetail{
				TemplateID: payload[i].TemplateID,
				Status:     hcmmodel.PsychologicalTestDetailStatusOpen,
			}

			err := h.Save(&payload[i])
			if err != nil {
				return nil, fmt.Errorf("error when save schedule: %s", err.Error())
			}
		}

		templates := []shemodel.MCUItemTemplate{}
		err := h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", templateIDs...),
		), &templates)
		if err != nil {
			return nil, fmt.Errorf("error when get template: %s", err.Error())
		}

		mapTemplate := lo.Associate(templates, func(m shemodel.MCUItemTemplate) (string, string) {
			return m.ID, m.Name
		})

		for i := range details {
			details[i].TemplateName = mapTemplate[details[i].TemplateID]
		}

		tds := []hcmmodel.TalentDevelopmentAssesment{}
		err = h.Gets(new(hcmmodel.TalentDevelopmentAssesment), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("_id", payload[0].TestID),
			),
		), &tds)
		if err != nil {
			return nil, fmt.Errorf("error when get talent development assesment: %s", err.Error())
		}

		for _, d := range tds {
			d.PsychoTests = details
			err := h.Save(&d)
			if err != nil {
				return nil, fmt.Errorf("error when save talent development assesment: %s", err.Error())
			}
		}
	}

	return payload, nil
}

func (m *TestScheduleHandler) DeleteSchedule(ctx *kaos.Context, id string) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	test := new(hcmmodel.TestSchedule)
	err := h.GetByID(test, id)
	if err != nil {
		return nil, fmt.Errorf("error when get schedule: %v", err)
	}

	switch test.TestScheduleType {
	case hcmmodel.TestScheduleTypePsychological:
		// get psychological test
		psychologicalTests := []hcmmodel.PsychologicalTest{}
		err = h.Gets(new(hcmmodel.PsychologicalTest), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("JobVacancyID", test.TestID),
		), &psychologicalTests)
		if err != nil {
			return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
		}

		for _, t := range psychologicalTests {
			details := make([]hcmmodel.PsychologicalTestDetail, 0)
			for _, d := range t.Details {
				if d.TemplateID != test.TemplateID {
					details = append(details, d)
				}
			}

			t.Details = details

			go func(psycho hcmmodel.PsychologicalTest) {
				h.Update(&psycho)
			}(t)
		}
	case hcmmodel.TestScheduleTypeTDC:
		// get training participant
		trainingTests := []hcmmodel.TrainingDevelopmentParticipant{}
		err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("TrainingCenterID", test.TestID),
		), &trainingTests)
		if err != nil {
			return nil, fmt.Errorf("error when get participant: %s", err.Error())
		}

		for _, t := range trainingTests {
			details := make([]hcmmodel.TrainingDevelopmentParticipantDetail, 0)
			for _, d := range t.TestDetails {
				if d.TemplateID != test.TemplateID {
					details = append(details, d)
				}
			}
			t.TestDetails = details

			go func(psycho hcmmodel.TrainingDevelopmentParticipant) {
				h.Update(&psycho)
			}(t)
		}
	case hcmmodel.TestScheduleTypeTD:
		// get talent development assesment test
		tdAssesment := []hcmmodel.TalentDevelopmentAssesment{}
		err = h.Gets(new(hcmmodel.TalentDevelopmentAssesment), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("_id", test.TestID),
		), &tdAssesment)
		if err != nil {
			return nil, fmt.Errorf("error when get talent development assesment: %s", err.Error())
		}

		for _, t := range tdAssesment {
			details := make([]hcmmodel.PsychologicalTestDetail, 0)
			for _, d := range t.PsychoTests {
				if d.TemplateID != test.TemplateID {
					details = append(details, d)
				}
			}
			t.PsychoTests = details

			go func(psycho hcmmodel.TalentDevelopmentAssesment) {
				h.Update(&psycho)
			}(t)
		}
	}

	h.Delete(test)

	return "success", nil
}
