package hcmlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"go.mongodb.org/mongo-driver/bson"
)

type TDCPracticeDurationHandler struct {
}

type TDCSavePracticeDurationRequest struct {
	Details []hcmmodel.TrainingDevelopmentPracticeDuration
}

func (m *TDCPracticeDurationHandler) Save(ctx *kaos.Context, payload *TDCSavePracticeDurationRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	for _, detail := range payload.Details {
		err := h.Save(&detail)
		if err != nil {
			return nil, fmt.Errorf("error when save practive: %s", err.Error())
		}
	}

	return "success", nil
}

func (m *TDCPracticeDurationHandler) GetDate(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	pipe := []bson.M{}
	match := bson.M{}
	if len(payload.Where.Items) > 0 {
		for _, f := range payload.Where.Items {
			match[f.Field] = bson.M{string(f.Op): f.Value}
		}

		pipe = append(pipe, bson.M{"$match": match})
	}

	pipe = append(pipe, []bson.M{
		{
			"$group": bson.M{
				"_id": "$Date",
			},
		},
		{
			"$project": bson.M{
				"_id":  0,
				"Date": "$_id",
			},
		},
	}...)

	type date struct {
		Date time.Time
	}

	dates := []date{}
	cmd := dbflex.From(new(hcmmodel.TrainingDevelopmentPracticeDuration).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &dates); err != nil {
		return nil, fmt.Errorf("err when get get practice duration: %s", err.Error())
	}

	return dates, nil
}
