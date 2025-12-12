package hcmlogic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type TrainingDevelopmentHandler struct {
}

type SaveParticipantEmployeeDetail struct {
	ID         string
	EmployeeID string
}

type SaveParticipantDetail struct {
	JobID     string // only for new hire
	Employees []SaveParticipantEmployeeDetail
}

type TDCSaveParticipantRequest struct {
	TrainingCenterID string
	Details          []SaveParticipantDetail
}

func (m *TrainingDevelopmentHandler) SaveParticipant(ctx *kaos.Context, payload *TDCSaveParticipantRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	for _, detail := range payload.Details {
		for _, emp := range detail.Employees {
			applicant := new(hcmmodel.TrainingDevelopmentParticipant)
			if emp.ID != "" {
				applicant.ID = emp.ID
			}

			if detail.JobID != "" {
				applicant.ManpowerRequestID = detail.JobID

				training := new(hcmmodel.Training)
				err := h.GetByParm(training, dbflex.NewQueryParam().SetWhere(
					dbflex.And(
						dbflex.Eq("JobVacancyID", applicant.ManpowerRequestID),
						dbflex.Eq("CandidateID", applicant.EmployeeID),
					),
				))
				if err != nil {
					return nil, fmt.Errorf("error when get training: %s", err.Error())
				}

				training.TrainingStatus = "Open"
				err = h.Save(training)
				if err != nil {
					return nil, fmt.Errorf("error when save training: %s", err.Error())
				}
			}

			applicant.EmployeeID = emp.EmployeeID
			applicant.TrainingCenterID = payload.TrainingCenterID

			err := h.Save(applicant)
			if err != nil {
				return nil, fmt.Errorf("error when save applicant: %s", err.Error())
			}
		}
	}

	return "success", nil
}

type ImportParticipantRequest struct {
	FileBase64       string
	TrainingCenterID string
}

func (m *TrainingDevelopmentHandler) ImportParticipant(ctx *kaos.Context, payload *ImportParticipantRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	data, err := base64.StdEncoding.DecodeString(payload.FileBase64)
	if err != nil {
		return nil, fmt.Errorf("error when decode: %s", err.Error())
	}

	file, err := excelize.OpenReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("error when open file: %s", err.Error())
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error when read file: %s", err.Error())
	}

	if len(rows) == 0 {
		return nil, errors.New("empty file")
	} else if len(rows[0]) > 2 {
		return nil, errors.New("invalid file template, should be only NIK and Name")
	} else {
		// check header
		if rows[0][0] != "NIK" || rows[0][1] != "Name" {
			return nil, errors.New("invalid file template")
		}
	}

	niks := make([]string, len(rows)-1)
	for i, row := range rows[1:] {
		if row[0] != "" {
			niks[i] = row[0]
		}
	}

	type employeeDetail struct {
		EmployeeID string
	}

	employeeDetails := make([]employeeDetail, 0)
	err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeNo", niks...),
	).SetSelect("EmployeeID"), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err)
	}

	for _, emp := range employeeDetails {
		applicant := new(hcmmodel.TrainingDevelopmentParticipant)
		applicant.EmployeeID = emp.EmployeeID
		applicant.TrainingCenterID = payload.TrainingCenterID

		err := h.Save(applicant)
		if err != nil {
			return nil, fmt.Errorf("error when save applicant: %s", err.Error())
		}
	}

	return "success", nil
}

type TDCGetSchedulesRequest struct {
	Date   time.Time
	ShowBy string
}

func (m *TrainingDevelopmentHandler) GetSchedules(ctx *kaos.Context, payload *TDCGetSchedulesRequest) (interface{}, error) {
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

	var startFilter time.Time
	var end time.Time
	if payload.ShowBy == "MONTH" {
		startFilter = time.Date(payload.Date.Year(), payload.Date.Month(), 1, 0, 0, 0, 0, loc)
		end = startFilter.AddDate(0, 1, 0)
	} else {
		startFilter = time.Date(payload.Date.Year(), time.January, 1, 0, 0, 0, 0, loc)
		end = startFilter.AddDate(1, 0, 0)
	}

	trainings := make([]hcmmodel.TrainingDevelopment, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopment), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Or(
				dbflex.And(
					dbflex.Gte("RequestTrainingDateFrom", startFilter),
				),
				dbflex.And(
					dbflex.Lt("RequestTrainingDateFrom", startFilter),
					dbflex.Gt("RequestTrainingDateTo", startFilter),
				),
			),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted),
		),
	), &trainings)
	if err != nil {
		return nil, fmt.Errorf("error when get training development: %s", err)
	}

	type schedule struct {
		ID               string
		Title            string
		Start            time.Time
		End              time.Time
		Description      string
		TrainingDateFrom time.Time
		TrainingDateTo   time.Time
		AssessmentType   string
	}

	ids := make([]string, len(trainings))
	titleIds := make([]string, len(trainings))
	for i, t := range trainings {
		ids[i] = t.ID
		titleIds[i] = t.TrainingTitle
	}

	details := make([]hcmmodel.TrainingDevelopmentDetail, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopmentDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("TrainingCenterID", ids...),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted),
		),
	), &details)
	if err != nil {
		return nil, fmt.Errorf("error when get training development detail: %s", err)
	}

	mapDetail := lo.Associate(details, func(item hcmmodel.TrainingDevelopmentDetail) (string, hcmmodel.TrainingDevelopmentDetail) {
		return item.TrainingCenterID, item
	})

	titles := make([]hcmmodel.TrainingDevelopmentTitle, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopmentTitle), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", titleIds...),
		),
	), &titles)
	if err != nil {
		return nil, fmt.Errorf("error when get training development title: %s", err)
	}

	mapTitle := lo.Associate(titles, func(item hcmmodel.TrainingDevelopmentTitle) (string, string) {
		return item.ID, item.Name
	})

	result := []schedule{}
	for _, t := range trainings {
		if v, ok := mapDetail[t.ID]; ok {
			var start time.Time
			if startFilter.After(t.RequestTrainingDateFrom) {
				start = startFilter
			} else {
				start = t.RequestTrainingDateFrom
			}

			for start.Before(end) {
				startTime := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
				data := schedule{
					ID:               t.ID,
					Title:            mapTitle[t.TrainingTitle],
					Start:            startTime,
					End:              startTime,
					TrainingDateFrom: v.TrainingDateFrom,
					TrainingDateTo:   v.TrainingDateTo,
					AssessmentType:   v.AssessmentType,
				}

				result = append(result, data)

				start = start.AddDate(0, 0, 1)

				if start.After(t.RequestTrainingDateTo) {
					break
				}
			}
		}
	}

	return result, nil
}

type TDCGetReportsRequest struct {
	EmployeeID   string
	TrainingType string // Exgternal, Internal
	Search       string
}

type TDCGetReportsResponse struct {
	ID               string
	Date             time.Time
	TrainingName     string
	Organizer        string
	AssessmentType   string
	TrainingDateFrom time.Time
	TrainingDateTo   time.Time
}

func (m *TrainingDevelopmentHandler) GetReports(ctx *kaos.Context, payload *TDCGetReportsRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	participants := make([]hcmmodel.TrainingDevelopmentParticipant, 0)
	err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("EmployeeID", payload.EmployeeID),
	), &participants)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err)
	}

	ids := lo.Map(participants, func(item hcmmodel.TrainingDevelopmentParticipant, index int) string {
		return item.TrainingCenterID
	})

	filters := []*dbflex.Filter{
		dbflex.In("TrainingCenterID", ids...),
	}
	if strings.ToLower(payload.TrainingType) == "internal" {
		filters = append(filters, dbflex.Eq("ExternalTraining", false))
	} else {
		filters = append(filters, dbflex.Eq("ExternalTraining", true))
	}

	trainingDetails := make([]hcmmodel.TrainingDevelopmentDetail, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopmentDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	), &trainingDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get training detail: %s", err)
	}

	detailMap := make(map[string]hcmmodel.TrainingDevelopmentDetail)
	for _, detail := range trainingDetails {
		detailMap[detail.TrainingCenterID] = detail
	}

	filters = []*dbflex.Filter{
		dbflex.In("_id", ids...),
	}
	if payload.Search != "" {
		filters = append(filters, dbflex.Contains("TrainingTitle", payload.Search))
	}

	trainings := make([]hcmmodel.TrainingDevelopment, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopment), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	), &trainings)
	if err != nil {
		return nil, fmt.Errorf("error when get training: %s", err)
	}

	titleIds := make([]string, len(trainings))
	for i, t := range trainings {
		titleIds[i] = t.TrainingTitle
	}

	titles := make([]hcmmodel.TrainingDevelopmentTitle, 0)
	err = h.Gets(new(hcmmodel.TrainingDevelopmentTitle), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", titleIds...),
		),
	), &titles)
	if err != nil {
		return nil, fmt.Errorf("error when get training development title: %s", err)
	}

	mapTitle := lo.Associate(titles, func(item hcmmodel.TrainingDevelopmentTitle) (string, string) {
		return item.ID, item.Name
	})

	result := make([]TDCGetReportsResponse, len(trainings))
	for i, tr := range trainings {
		detail, exists := detailMap[tr.ID]
		result[i] = TDCGetReportsResponse{
			ID:               tr.ID,
			Date:             tr.RequestDate,
			TrainingName:     mapTitle[tr.TrainingTitle],
			Organizer:        "",
			TrainingDateFrom: detail.TrainingDateFrom,
			TrainingDateTo:   detail.TrainingDateTo,
			AssessmentType:   detail.AssessmentType,
		}

		if !exists {
			result[i].TrainingDateFrom = time.Time{}
			result[i].TrainingDateTo = time.Time{}
			result[i].AssessmentType = ""
		}
	}

	return result, nil
}

type TrainingSaveAnswerRequest struct {
	TrainingCenterID string
	TemplateTestID   string
	Details          []TrainingSaveAnswerDetail
}

type TrainingSaveAnswerDetail struct {
	QuestionID  string
	AnswerID    string
	AnswerValue string
}

func (m *TrainingDevelopmentHandler) SaveAnswer(ctx *kaos.Context, payload *TrainingSaveAnswerRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)

	// get test
	test := new(shemodel.MCUItemTemplate)
	err := h.GetByID(test, payload.TemplateTestID)
	if err != nil {
		return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
	}

	type question struct {
		Condition   []shemodel.MCULineCondition
		AnswerValue int
	}

	mapQuestion := lo.Associate(test.Lines, func(m shemodel.MCUItemTemplateLine) (string, *question) {
		return m.ID, &question{Condition: m.Condition, AnswerValue: m.AnswerValue}
	})

	// get participant
	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err = h.GetByFilter(participant, dbflex.And(
		dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		dbflex.Eq("EmployeeID", userID),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	questionIDs := make([]string, len(payload.Details))
	for i, d := range payload.Details {
		questionIDs[i] = d.QuestionID
	}

	answers := []hcmmodel.TrainingAnswerHistory{}
	err = h.Gets(new(hcmmodel.TrainingAnswerHistory), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("QuestionID", questionIDs...),
			dbflex.Eq("TemplateTestID", payload.TemplateTestID),
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
			dbflex.Eq("ParticipantID", participant.ID),
		),
	), &answers)
	if err != nil {
		return nil, fmt.Errorf("error when get answer history: %s", err.Error())
	}

	mapAnswer := lo.Associate(answers, func(m hcmmodel.TrainingAnswerHistory) (string, hcmmodel.TrainingAnswerHistory) {
		return m.QuestionID, m
	})

	for _, d := range payload.Details {
		history := hcmmodel.TrainingAnswerHistory{}
		if v, ok := mapAnswer[d.QuestionID]; ok {
			history = v
		} else {
			history = hcmmodel.TrainingAnswerHistory{
				TrainingCenterID: payload.TrainingCenterID,
				TemplateTestID:   payload.TemplateTestID,
				QuestionID:       d.QuestionID,
				AnswerID:         d.AnswerID,
				ParticipantID:    participant.ID,
			}
		}
		history.AnswerID = d.AnswerID

		// if condition not filled, question is essay
		// if condition filled, question is multiple choice
		isEssay := true
		if v, ok := mapQuestion[d.QuestionID]; ok {
			if len(v.Condition) > 0 {
				isEssay = false
			}
		}

		// if essay
		if isEssay {
			history.Answer = d.AnswerValue
		} else {
			// get question
			if v, ok := mapQuestion[d.QuestionID]; ok {
				// get correct answer
				for _, answer := range v.Condition {
					if answer.ID == d.AnswerID && answer.Value {
						history.Score = v.AnswerValue
						break
					}
				}
			}
		}

		err = h.Save(&history)
		if err != nil {
			return nil, fmt.Errorf("error when save answer: %s", err.Error())
		}
	}

	return "success", nil
}

type TrainingAssignEssayScoreRequest struct {
	TrainingCenterID string
	TemplateTestID   string
	EmployeeID       string
	Details          []TrainingAssignEssayScoreDetail
}

type TrainingAssignEssayScoreDetail struct {
	QuestionID string
	Score      int
}

func (m *TrainingDevelopmentHandler) AssignEssayScore(ctx *kaos.Context, payload *TrainingAssignEssayScoreRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get participant
	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err := h.GetByFilter(participant, dbflex.And(
		dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		dbflex.Eq("EmployeeID", payload.EmployeeID),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	questionIDs := make([]string, len(payload.Details))
	for i, d := range payload.Details {
		questionIDs[i] = d.QuestionID
	}

	answers := []hcmmodel.TrainingAnswerHistory{}
	err = h.Gets(new(hcmmodel.TrainingAnswerHistory), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("QuestionID", questionIDs...),
			dbflex.Eq("TemplateTestID", payload.TemplateTestID),
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
			dbflex.Eq("ParticipantID", participant.ID),
		),
	), &answers)
	if err != nil {
		return nil, fmt.Errorf("error when get answer history: %s", err.Error())
	}

	mapAnswer := lo.Associate(answers, func(m hcmmodel.TrainingAnswerHistory) (string, hcmmodel.TrainingAnswerHistory) {
		return m.QuestionID, m
	})

	for _, d := range payload.Details {
		if v, ok := mapAnswer[d.QuestionID]; ok {
			v.Score = d.Score
			err = h.Save(&v)
			if err != nil {
				return nil, fmt.Errorf("error when save answer: %s", err.Error())
			}
		}
	}

	return "success", nil
}

type TrainingSubmitRequest struct {
	TrainingCenterID string
	TemplateTestID   string
}

func (m *TrainingDevelopmentHandler) Submit(ctx *kaos.Context, payload *TrainingSubmitRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get test
	test := new(shemodel.MCUItemTemplate)
	err := h.GetByID(test, payload.TemplateTestID)
	if err != nil {
		return nil, fmt.Errorf("error when get template test: %s", err.Error())
	}

	// only get multiple choice question
	questionID := []string{}
	for _, q := range test.Lines {
		if len(q.Condition) > 0 {
			questionID = append(questionID, q.ID)
		}
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	// get participant
	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err = h.GetByFilter(participant, dbflex.And(
		dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		dbflex.Eq("EmployeeID", userID),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	indexChanged := 0
	for i := range participant.TestDetails {
		if participant.TestDetails[i].TemplateID == payload.TemplateTestID {
			indexChanged = i
			break
		}
	}

	// only calculate multiple choice
	pipe := []bson.M{
		{
			"$match": bson.M{
				"QuestionID":       bson.M{"$in": questionID},
				"TrainingCenterID": payload.TrainingCenterID,
				"TemplateTestID":   payload.TemplateTestID,
				"ParticipantID":    participant.ID,
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Score": bson.M{"$sum": "$Score"},
			},
		},
	}

	type trainingScore struct {
		Score int
	}

	scores := []trainingScore{}
	cmd := dbflex.From(new(hcmmodel.TrainingAnswerHistory).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &scores); err != nil {
		return nil, fmt.Errorf("err when get history answer: %s", err.Error())
	}

	if len(scores) > 0 {
		participant.TestDetails[indexChanged].Score += scores[0].Score
	}

	participant.TestDetails[indexChanged].Status = hcmmodel.TrainingDevelopmentParticipantStatusDone
	err = h.Save(participant)
	if err != nil {
		return nil, fmt.Errorf("error when save schedule: %s", err.Error())
	}

	return "success", nil
}

type TrainingSubmitEssayRequest struct {
	ParticipantID  string
	TemplateTestID string
	MultipleScore  int
}

func (m *TrainingDevelopmentHandler) SubmitEssay(ctx *kaos.Context, payload *TrainingSubmitEssayRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get test
	test := new(shemodel.MCUItemTemplate)
	err := h.GetByID(test, payload.TemplateTestID)
	if err != nil {
		return nil, fmt.Errorf("error when get template test: %s", err.Error())
	}

	// only get essay question
	questionID := []string{}
	for _, q := range test.Lines {
		if len(q.Condition) == 0 {
			questionID = append(questionID, q.ID)
		}
	}

	// get participant
	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err = h.GetByID(participant, payload.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	pipe := []bson.M{
		{
			"$match": bson.M{
				"QuestionID":       bson.M{"$in": questionID},
				"TrainingCenterID": participant.TrainingCenterID,
				"TemplateTestID":   payload.TemplateTestID,
				"ParticipantID":    participant.ID,
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Score": bson.M{"$sum": "$Score"},
			},
		},
	}

	type trainingScore struct {
		Score int
	}

	scores := []trainingScore{}
	cmd := dbflex.From(new(hcmmodel.TrainingAnswerHistory).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &scores); err != nil {
		return nil, fmt.Errorf("err when get history answer: %s", err.Error())
	}

	if len(scores) > 0 {
		for i := range participant.TestDetails {
			if participant.TestDetails[i].TemplateID == payload.TemplateTestID {
				participant.TestDetails[i].Score = scores[0].Score + payload.MultipleScore
				break
			}
		}
	}

	err = h.Save(participant)
	if err != nil {
		return nil, fmt.Errorf("error when save schedule: %s", err.Error())
	}

	return "success", nil
}

func (m *TrainingDevelopmentHandler) GetQuestionAnswer(ctx *kaos.Context, payload *TrainingSubmitEssayRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get test
	test := new(shemodel.MCUItemTemplate)
	err := h.GetByID(test, payload.TemplateTestID)
	if err != nil {
		return nil, fmt.Errorf("error when get template test: %s", err.Error())
	}

	// get participant
	participant := new(hcmmodel.TrainingDevelopmentParticipant)
	err = h.GetByID(participant, payload.ParticipantID)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	type answer struct {
		AnswerID   string
		Score      int
		Answer     string
		QuestionID string
	}

	answers := []answer{}
	err = h.Gets(new(hcmmodel.TrainingAnswerHistory), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", participant.TrainingCenterID),
			dbflex.Eq("TemplateTestID", payload.TemplateTestID),
			dbflex.Eq("ParticipantID", participant.ID),
		),
	).SetSelect("QuestionID", "AnswerID", "Answer", "Score"), &answers)
	if err != nil {
		return nil, fmt.Errorf("error when get answer history: %s", err.Error())
	}

	type scoreAnswer struct {
		Score  int
		Answer string
	}

	// mapAnswer := map[string]*answer{}
	mapAnswer := lo.Associate(answers, func(a answer) (string, *scoreAnswer) {
		return a.QuestionID + a.AnswerID, &scoreAnswer{Score: a.Score, Answer: a.Answer}
	})

	type multipleChoice struct {
		shemodel.MCULineCondition
		IsAnswer bool
		Score    int
	}

	result := []codekit.M{}
	i := 1
	for _, q := range test.Lines {
		// multiple choice
		if len(q.Condition) > 0 {
			details := make([]multipleChoice, len(q.Condition))

			for i, c := range q.Condition {
				detail := multipleChoice{
					MCULineCondition: c,
				}
				if v, ok := mapAnswer[q.ID+c.ID]; ok {
					detail.Score = v.Score
					detail.IsAnswer = true
				}
				details[i] = detail
			}

			result = append(result, codekit.M{"No": i, "QuestionID": q.ID, "Question": q.Description, "Details": details})
		} else if q.Type == shemodel.MCU_TYPE_STRING {
			data := codekit.M{
				"No":         i,
				"QuestionID": q.ID,
				"Question":   q.Description,
				"MaxScore":   q.AnswerValue,
				"Answer":     "",
				"Score":      0,
			}

			// essay
			if v, ok := mapAnswer[q.ID]; ok {
				data["Answer"] = v.Answer
				data["Score"] = v.Score
			}

			result = append(result, data)
		}

		i++
	}

	return result, nil
}

type AssesmentRequest struct {
	TrainingCenterID string
	Take             int
	Skip             int
}

type writtenTestScoreDetail struct {
	TemplateID   string
	TemplateName string
	Score        int
}

type writtenTestDetail struct {
	Stage       string
	TestDetails []writtenTestScoreDetail
}

type employee struct {
	ID   string `bson:"_id"`
	Name string
}

func (m *TrainingDevelopmentHandler) AssesmentStaff(ctx *kaos.Context, payload *AssesmentRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	type participant struct {
		ID          string `bson:"_id"`
		EmployeeID  string
		TestDetails []hcmmodel.TrainingDevelopmentParticipantDetail
	}

	// get participant
	participants := []participant{}
	err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	).
		SetSelect("_id", "EmployeeID", "TestDetails").
		SetSkip(payload.Skip).
		SetTake(payload.Take),
		&participants)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	ids := make([]string, len(participants))
	employeeIDs := make([]string, len(participants))
	for i, p := range participants {
		ids[i] = p.ID
		employeeIDs[i] = p.EmployeeID
	}

	// get count
	var count int
	count, err = h.Count(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	// get employee
	employees := []employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", ids...),
		),
	).SetSelect("_id", "Name"), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m employee) (string, string) {
		return m.ID, m.Name
	})

	type practiceTest struct {
		ParticipantID string
		TotalScore    float64
	}

	// get practice score
	praticeTests := []practiceTest{}
	err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeTestStaff), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("ParticipantID", employeeIDs...),
		),
	).SetSelect("ParticipantID", "TotalScore"), &praticeTests)
	if err != nil {
		return nil, fmt.Errorf("error when get practice test: %s", err.Error())
	}

	mapPratice := lo.Associate(praticeTests, func(m practiceTest) (string, float64) {
		return m.ParticipantID, m.TotalScore
	})

	type assesment struct {
		ParticipantID     string
		EmployeeID        string
		Name              string
		WrittenTest       float64
		WrittenTestDetail []writtenTestDetail
		PracticeTest      float64
	}

	result := make([]assesment, len(participants))
	for i, p := range participants {
		asses := assesment{
			ParticipantID: p.ID,
			EmployeeID:    p.EmployeeID,
			PracticeTest:  mapPratice[p.ID],
			Name:          mapEmployee[p.EmployeeID],
		}

		mapWrittenTest := map[string][]writtenTestScoreDetail{}
		writtenScore := 0.0
		for _, td := range p.TestDetails {
			writtenScore += float64(td.Score)

			// build detail writted test
			mapWrittenTest[string(td.Stage)] = append(mapWrittenTest[string(td.Stage)], writtenTestScoreDetail{
				TemplateID:   td.TemplateID,
				TemplateName: td.TemplateName,
				Score:        td.Score,
			})
		}

		// writted details
		for k, v := range mapWrittenTest {
			asses.WrittenTestDetail = append(asses.WrittenTestDetail, writtenTestDetail{
				Stage:       k,
				TestDetails: v,
			})
		}

		asses.WrittenTest = writtenScore
		result[i] = asses
	}

	return codekit.M{"data": result, "count": count}, nil
}

func (m *TrainingDevelopmentHandler) AssesmentDriver(ctx *kaos.Context, payload *AssesmentRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	type participant struct {
		ID               string `bson:"_id"`
		TrainingCenterID string
		EmployeeID       string
		TestDetails      []hcmmodel.TrainingDevelopmentParticipantDetail
	}

	// get participant
	participants := []participant{}
	err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	).
		SetSelect("_id", "EmployeeID", "TrainingCenterID", "TestDetails").
		SetSkip(payload.Skip).
		SetTake(payload.Take),
		&participants)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	ids := make([]string, len(participants))
	employeeIDs := make([]string, len(participants))
	for i, p := range participants {
		ids[i] = p.ID
		employeeIDs[i] = p.EmployeeID
	}

	// get count
	var count int
	count, err = h.Count(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	// get employee
	employees := []employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", employeeIDs...),
		),
	).SetSelect("_id", "Name"), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m employee) (string, string) {
		return m.ID, m.Name
	})

	pipe := []bson.M{
		{
			"$match": bson.M{
				"ParticipantID": bson.M{
					"$in": ids,
				},
			},
		},
		{
			"$group": bson.M{
				"_id":      "$ParticipantID",
				"Duration": bson.M{"$sum": "$Duration"},
			},
		},
	}

	type practiceDuration struct {
		ParticipantID string `bson:"_id"`
		Duration      float64
	}

	// get practice duration
	practiceDurations := []practiceDuration{}
	cmd := dbflex.From(new(hcmmodel.TrainingDevelopmentPracticeDuration).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &practiceDurations); err != nil {
		return nil, fmt.Errorf("err when get pratice duration: %s", err.Error())
	}

	mapPracticeDuration := lo.Associate(practiceDurations, func(m practiceDuration) (string, float64) {
		return m.ParticipantID, m.Duration
	})

	type practiceScore struct {
		TrainingCenterID string
		EmployeeID       string
		FinalScore       float64
		Type             string
	}

	// get practice test score
	practiceScores := []practiceScore{}
	err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeScore), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
			dbflex.In("EmployeeID", employeeIDs...),
		),
	).SetSelect("TrainingCenterID", "EmployeeID", "FinalScore", "Type"), &practiceScores)
	if err != nil {
		return nil, fmt.Errorf("error when get practice test score: %s", err.Error())
	}

	type practiceScoreDetail struct {
		FinalScore float64
		Type       string
	}

	mapPracticeScore := map[string][]practiceScoreDetail{}
	for _, ps := range practiceScores {
		mapPracticeScore[ps.TrainingCenterID+ps.EmployeeID] = append(mapPracticeScore[ps.TrainingCenterID+ps.EmployeeID], practiceScoreDetail{
			FinalScore: ps.FinalScore,
			Type:       ps.Type,
		})
	}

	type assesment struct {
		ParticipantID            string
		EmployeeID               string
		Name                     string
		PracticeTestScore        float64
		PracticeTestScoreDetails []practiceScoreDetail
		WrittenTest              float64
		WrittenTestDetail        []writtenTestDetail
		PracticeTestDuration     float64
	}

	result := make([]assesment, len(participants))
	for i, p := range participants {
		asses := assesment{
			ParticipantID:            p.ID,
			EmployeeID:               p.EmployeeID,
			PracticeTestDuration:     mapPracticeDuration[p.ID],
			PracticeTestScoreDetails: mapPracticeScore[p.TrainingCenterID+p.EmployeeID],
			Name:                     mapEmployee[p.EmployeeID],
		}

		mapWrittenTest := map[string][]writtenTestScoreDetail{}
		writtenScore := 0.0
		for _, td := range p.TestDetails {
			writtenScore += float64(td.Score)

			// build detail writted test
			mapWrittenTest[string(td.Stage)] = append(mapWrittenTest[string(td.Stage)], writtenTestScoreDetail{
				TemplateID:   td.TemplateID,
				TemplateName: td.TemplateName,
				Score:        td.Score,
			})
		}
		asses.WrittenTest = lo.Ternary(len(p.TestDetails) == 0, 0, writtenScore/float64(len(p.TestDetails)))

		// writted details
		for k, v := range mapWrittenTest {
			asses.WrittenTestDetail = append(asses.WrittenTestDetail, writtenTestDetail{
				Stage:       k,
				TestDetails: v,
			})
		}

		// set practice test score
		for _, d := range asses.PracticeTestScoreDetails {
			asses.PracticeTestScore += d.FinalScore
		}

		asses.PracticeTestScore = lo.Ternary(len(asses.PracticeTestScoreDetails) == 0, 0, asses.PracticeTestScore/float64(len(asses.PracticeTestScoreDetails)))

		result[i] = asses
	}

	return codekit.M{"data": result, "count": count}, nil
}

func (m *TrainingDevelopmentHandler) AssesmentMechanic(ctx *kaos.Context, payload *AssesmentRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	type participant struct {
		ID               string `bson:"_id"`
		EmployeeID       string
		TrainingCenterID string
		TestDetails      []hcmmodel.TrainingDevelopmentParticipantDetail
	}

	// get participant
	participants := []participant{}
	err := h.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	).
		SetSelect("_id", "EmployeeID", "TestDetails", "TrainingCenterID").
		SetSkip(payload.Skip).
		SetTake(payload.Take),
		&participants)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	ids := make([]string, len(participants))
	employeeIDs := make([]string, len(participants))
	for i, p := range participants {
		ids[i] = p.ID
		employeeIDs[i] = p.EmployeeID
	}

	// get count
	var count int
	count, err = h.Count(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	// get employee
	employees := []employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", employeeIDs...),
		),
	).SetSelect("_id", "Name"), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m employee) (string, string) {
		return m.ID, m.Name
	})

	type practiceScore struct {
		TrainingCenterID string
		EmployeeID       string
		FinalScore       float64
	}

	// get practice test score
	practiceScores := []practiceScore{}
	err = h.Gets(new(hcmmodel.TrainingDevelopmentPracticeScore), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("TrainingCenterID", payload.TrainingCenterID),
			dbflex.In("EmployeeID", employeeIDs...),
		),
	).SetSelect("TrainingCenterID", "EmployeeID", "FinalScore"), &practiceScores)
	if err != nil {
		return nil, fmt.Errorf("error when get practice test score: %s", err.Error())
	}

	mapPracticeScore := map[string]float64{}
	for _, ps := range practiceScores {
		mapPracticeScore[ps.TrainingCenterID+ps.EmployeeID] = ps.FinalScore
	}

	type assesment struct {
		ParticipantID     string
		EmployeeID        string
		Name              string
		PracticeTestScore float64
		WrittenTest       float64
		WrittenTestDetail []writtenTestDetail
	}

	result := make([]assesment, len(participants))
	for i, p := range participants {
		asses := assesment{
			ParticipantID:     p.ID,
			EmployeeID:        p.EmployeeID,
			PracticeTestScore: mapPracticeScore[p.TrainingCenterID+p.EmployeeID],
			Name:              mapEmployee[p.EmployeeID],
		}

		mapWrittenTest := map[string][]writtenTestScoreDetail{}
		writtenScore := 0.0
		for _, td := range p.TestDetails {
			writtenScore += float64(td.Score)

			// build detail writted test
			mapWrittenTest[string(td.Stage)] = append(mapWrittenTest[string(td.Stage)], writtenTestScoreDetail{
				TemplateID:   td.TemplateID,
				TemplateName: td.TemplateName,
				Score:        td.Score,
			})
		}
		asses.WrittenTest = lo.Ternary(len(p.TestDetails) == 0, 0, writtenScore/float64(len(p.TestDetails)))

		// writted details
		for k, v := range mapWrittenTest {
			asses.WrittenTestDetail = append(asses.WrittenTestDetail, writtenTestDetail{
				Stage:       k,
				TestDetails: v,
			})
		}

		result[i] = asses
	}

	return codekit.M{"data": result, "count": count}, nil
}
