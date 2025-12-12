package hcmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type PsychologicalTestHandler struct {
}

type SaveAnswerDetail struct {
	QuestionID         string
	AnswerID           string
	AnswerValue        string
	IsMostQuestionType bool
}

type PsychologicalTestSaveAnswerRequest struct {
	JobID          string
	TemplateTestID string
	Details        []SaveAnswerDetail
	TestType       hcmmodel.TestScheduleType
}

func (m *PsychologicalTestHandler) SaveAnswer(ctx *kaos.Context, payload *PsychologicalTestSaveAnswerRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// get test
	test := new(shemodel.MCUItemTemplate)
	err := h.GetByID(test, payload.TemplateTestID)
	if err != nil {
		return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
	}

	master := new(tenantcoremodel.MasterData)
	err = h.GetByID(master, test.Instruction.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("error when get calculation method: %s", err.Error())
	}

	id := ""
	if payload.TestType == hcmmodel.TestScheduleTypeTD {
		// get talent development assesment
		td := new(hcmmodel.TalentDevelopmentAssesment)
		err = h.GetByFilter(td, dbflex.And(
			dbflex.Eq("_id", payload.JobID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get talent development test: %s", err.Error())
		}
		id = td.ID
	} else if payload.TestType == hcmmodel.TestScheduleTypePsychological {
		userID := sebar.GetUserIDFromCtx(ctx)
		if userID == "" {
			return nil, errors.New("missing: User, please relogin")
		}
		// get psychologival test
		psychoTest := new(hcmmodel.PsychologicalTest)
		err = h.GetByFilter(psychoTest, dbflex.And(
			dbflex.Eq("JobVacancyID", payload.JobID),
			dbflex.Eq("CandidateID", userID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
		}
		id = psychoTest.ID
	}

	questionIDs := make([]string, len(payload.Details))
	for i, d := range payload.Details {
		questionIDs[i] = d.QuestionID
	}

	answers := []hcmmodel.PsychologicalAnswerHistory{}
	err = h.Gets(new(hcmmodel.PsychologicalAnswerHistory), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("QuestionID", questionIDs...),
			dbflex.Eq("TemplateTestID", payload.TemplateTestID),
			dbflex.Eq("TestID", id),
		),
	), &answers)
	if err != nil {
		return nil, fmt.Errorf("error when get answer history: %s", err.Error())
	}

	mapAnswer := lo.Associate(answers, func(m hcmmodel.PsychologicalAnswerHistory) (string, hcmmodel.PsychologicalAnswerHistory) {
		return m.QuestionID, m
	})

	if master.Name == "CFIT" {
		return m.saveCFIT(h, mapAnswer, test, payload, id)
	} else {
		return m.saveDISC(h, mapAnswer, payload, id)
	}
}

func (m *PsychologicalTestHandler) saveCFIT(h *datahub.Hub, mapAnswer map[string]hcmmodel.PsychologicalAnswerHistory, test *shemodel.MCUItemTemplate, payload *PsychologicalTestSaveAnswerRequest, id string) (interface{}, error) {
	for _, d := range payload.Details {
		history := hcmmodel.PsychologicalAnswerHistory{}
		if v, ok := mapAnswer[d.QuestionID]; ok {
			history = v
		} else {
			history = hcmmodel.PsychologicalAnswerHistory{
				TestID:         id,
				TemplateTestID: payload.TemplateTestID,
				QuestionID:     d.QuestionID,
				AnswerID:       d.AnswerID,
			}
		}
		history.AnswerID = d.AnswerID

		// get question
		for _, line := range test.Lines {
			if line.ID == d.QuestionID {
				// get answer
				for _, answer := range line.Condition {
					if answer.ID == d.AnswerID {
						// if correct answer
						if answer.Value {
							history.Score = line.AnswerValue
						}
						break
					}
				}
				break
			}
		}

		err := h.Save(&history)
		if err != nil {
			return nil, fmt.Errorf("error when save answer: %s", err.Error())
		}
	}

	return "success", nil
}

func (m *PsychologicalTestHandler) saveDISC(h *datahub.Hub, mapAnswer map[string]hcmmodel.PsychologicalAnswerHistory, payload *PsychologicalTestSaveAnswerRequest, id string) (interface{}, error) {
	ids := make([]string, len(payload.Details))
	for i, d := range payload.Details {
		ids[i] = d.AnswerValue
	}

	types := []tenantcoremodel.MasterData{}
	err := h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", ids...),
		),
	), &types)
	if err != nil {
		return nil, fmt.Errorf("error when get condition type: %s", err.Error())
	}

	mapType := lo.Associate(types, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	for _, d := range payload.Details {
		history := hcmmodel.PsychologicalAnswerHistory{}
		if v, ok := mapAnswer[d.QuestionID]; ok {
			history = v
		} else {
			history = hcmmodel.PsychologicalAnswerHistory{
				TestID:         id,
				TemplateTestID: payload.TemplateTestID,
				QuestionID:     d.QuestionID,
				AnswerID:       d.AnswerID,
			}
		}
		history.AnswerID = d.AnswerID
		history.Answer = mapType[d.AnswerValue]
		history.IsMostQuestionType = d.IsMostQuestionType

		err := h.Save(&history)
		if err != nil {
			return nil, fmt.Errorf("error when save answer: %s", err.Error())
		}
	}

	return "success", nil
}

type PsychologicalTestSubmitRequest struct {
	JobID          string
	TemplateTestID string
	TestType       hcmmodel.TestScheduleType
}

func (m *PsychologicalTestHandler) Submit(ctx *kaos.Context, payload *PsychologicalTestSubmitRequest) (interface{}, error) {
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

	master := new(tenantcoremodel.MasterData)
	err = h.GetByID(master, test.Instruction.CalculationMethod)
	if err != nil {
		return nil, fmt.Errorf("error when get calculation method: %s", err.Error())
	}

	var mdl orm.DataModel
	if payload.TestType == hcmmodel.TestScheduleTypeTD {
		// get talent development
		mdl = new(hcmmodel.TalentDevelopment)
		err = h.GetByFilter(mdl, dbflex.And(
			dbflex.Eq("_id", payload.JobID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get talent development test: %s", err.Error())
		}
	} else if payload.TestType == hcmmodel.TestScheduleTypePsychological {
		userID := sebar.GetUserIDFromCtx(ctx)
		if userID == "" {
			return nil, errors.New("missing: User, please relogin")
		}
		// get psychologival test
		mdl = new(hcmmodel.PsychologicalTest)
		err = h.GetByFilter(mdl, dbflex.And(
			dbflex.Eq("JobVacancyID", payload.JobID),
			dbflex.Eq("CandidateID", userID),
		))
		if err != nil {
			return nil, fmt.Errorf("error when get psychological test: %s", err.Error())
		}
	}

	if master.Name == "CFIT" {
		err = m.submitCFIT(h, payload, mdl)
	} else {
		err = m.submitDISC(h, payload, mdl)
	}
	if err != nil {
		return nil, fmt.Errorf("error when submit: %s", err.Error())
	}

	return "success", nil
}

func (m *PsychologicalTestHandler) submitCFIT(h *datahub.Hub, payload *PsychologicalTestSubmitRequest, mdl orm.DataModel) error {
	id, err := reflector.From(mdl).Get("ID")
	if err != nil {
		return err
	}
	pipe := []bson.M{
		{
			"$match": bson.M{
				"TestID":         id,
				"TemplateTestID": payload.TemplateTestID,
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Score": bson.M{"$sum": "$Score"},
			},
		},
	}

	type psychoTestScore struct {
		Score int
	}

	scores := []psychoTestScore{}
	cmd := dbflex.From(new(hcmmodel.PsychologicalAnswerHistory).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &scores); err != nil {
		return fmt.Errorf("err when get history answer: %s", err.Error())
	}

	iq := 0.0
	score := 0
	if len(scores) > 0 {
		score = scores[0].Score

		switch score {
		case 31:
			iq = 127
		case 30:
			iq = 124
		case 29:
			iq = 121
		case 28:
			iq = 119
		case 27:
			iq = 116
		case 26:
			iq = 113
		case 25:
			iq = 109
		case 24:
			iq = 106
		case 23:
			iq = 103
		case 22:
			iq = 100
		case 21:
			iq = 96
		case 20:
			iq = 94
		case 19:
			iq = 91
		case 18:
			iq = 88
		case 16:
			iq = 81
		case 15:
			iq = 78
		case 14:
			iq = 75
		case 13:
			iq = 72
		case 12:
			iq = 70
		case 11:
			iq = 67
		case 10:
			iq = 63
		case 9:
			iq = 60
		case 8:
			iq = 57
		case 7:
			iq = 55
		case 6:
			iq = 52
		case 5:
			iq = 49
		case 4:
			iq = 46
		case 3:
			iq = 43
		case 2:
			iq = 40
		case 1:
			iq = 37
		case 0:
			iq = 0
		}
	}

	testDetails, err := reflector.From(mdl).Get("Details")
	if err != nil {
		return err
	}

	details := testDetails.([]hcmmodel.PsychologicalTestDetail)
	for i := range details {
		if details[i].TemplateID == payload.TemplateTestID {
			details[i].Detail = codekit.M{
				"Score":   score,
				"IQScore": iq,
			}
			details[i].Status = hcmmodel.PsychologicalTestDetailStatusDone
			break
		}
	}

	if payload.TestType == hcmmodel.TestScheduleTypeTD {
		reflector.From(mdl).Set("PsychoTests", details).Flush()
	} else if payload.TestType == hcmmodel.TestScheduleTypePsychological {
		reflector.From(mdl).Set("Details", details).Flush()
	}

	err = h.Save(mdl)
	if err != nil {
		return fmt.Errorf("error when save psychological test: %s", err.Error())
	}

	return nil
}

func (m *PsychologicalTestHandler) submitDISC(h *datahub.Hub, payload *PsychologicalTestSubmitRequest, mdl orm.DataModel) error {
	id, err := reflector.From(mdl).Get("ID")
	if err != nil {
		return err
	}
	pipe := []bson.M{
		{
			"$match": bson.M{
				"TestID":         id,
				"TemplateTestID": payload.TemplateTestID,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"IsMostQuestionType": "$IsMostQuestionType",
					"Answer":             "$Answer",
				},
				"Count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"_id":                0,
				"IsMostQuestionType": "$_id.IsMostQuestionType",
				"Answer":             "$_id.Answer",
				"Count":              1,
			},
		},
	}

	type psychoTestAnswer struct {
		IsMostQuestionType bool
		Answer             string
		Count              int
	}

	answers := []psychoTestAnswer{}
	cmd := dbflex.From(new(hcmmodel.PsychologicalAnswerHistory).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &answers); err != nil {
		return fmt.Errorf("err when get history answer: %s", err.Error())
	}

	type detail struct {
		Answer string `bson:"Answer"`
		Count  int    `bson:"Count"`
	}

	type discResult struct {
		IsMostQuestionType bool     `bson:"IsMostQuestionType"`
		Detail             []detail `bson:"Detail"`
	}

	mapDisc := map[bool][]detail{
		true:  make([]detail, 0),
		false: make([]detail, 0),
	}
	for _, a := range answers {
		mapDisc[a.IsMostQuestionType] = append(mapDisc[a.IsMostQuestionType], detail{
			Answer: a.Answer,
			Count:  a.Count,
		})
	}

	orders := []string{"D", "I", "S", "C", "."}
	discDetails := make([]discResult, 2)
	i := 0
	for k, details := range mapDisc {
		discDetails[i] = discResult{
			IsMostQuestionType: k,
			Detail:             make([]detail, 6),
		}

		total := 0
		for j, o := range orders {
			detail := detail{
				Answer: o,
			}
			for _, v := range details {
				if o == v.Answer {
					total += v.Count
					detail.Count = v.Count
					break
				}
			}

			discDetails[i].Detail[j] = detail
		}
		discDetails[i].Detail[5] = detail{
			Answer: "Total",
			Count:  total,
		}

		i++
	}

	testDetails, err := reflector.From(mdl).Get("Details")
	if err != nil {
		return err
	}

	details := testDetails.([]hcmmodel.PsychologicalTestDetail)
	for i := range details {
		if details[i].TemplateID == payload.TemplateTestID {
			details[i].Detail = discDetails
			details[i].Status = hcmmodel.PsychologicalTestDetailStatusDone
			break
		}
	}

	if payload.TestType == hcmmodel.TestScheduleTypeTD {
		reflector.From(mdl).Set("PsychoTests", details).Flush()
	} else if payload.TestType == hcmmodel.TestScheduleTypePsychological {
		reflector.From(mdl).Set("Details", details).Flush()
	}

	err = h.Save(mdl)
	if err != nil {
		return fmt.Errorf("error when save psychological test: %s", err.Error())
	}

	return nil
}
