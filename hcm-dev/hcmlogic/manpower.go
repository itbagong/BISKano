package hcmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"go.mongodb.org/mongo-driver/bson"
)

type ManpowerHandler struct {
}

func (m *ManpowerHandler) GetAvailable(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	type manpower struct {
		ID   string `bson:"_id" json:"_id"`
		Name string
		// ClosedEmployeeTotal  int
		// AdditionalNumber     int
		// ReplacedEmployeeName string
	}

	payload = payload.MergeWhere(false, dbflex.Eq("IsClose", false))
	if payload.Where == nil {
		payload.Where = dbflex.In("Status", []string{"POSTED", "DRAFT", ""}...)
	} else {
		payload = payload.MergeWhere(false, dbflex.In("Status", []string{"POSTED", "DRAFT"}...))
	}
	// payload = payload.SetSelect("_id", "Name", "ClosedEmployeeTotal", "AdditionalNumber", "ReplacedEmployeeName")
	payload = payload.SetSelect("_id", "Name")

	requests := []manpower{}
	err := h.Gets(new(hcmmodel.ManpowerRequest), payload, &requests)
	if err != nil {
		return nil, fmt.Errorf("err when get manpower request: %s", err.Error())
	}

	// ids := lo.Map(requests, func(m manpower, index int) string {
	// 	return m.ID
	// })

	// // get passed employee from pkwt
	// pipe := []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"JobVacancyID": bson.M{"$in": ids},
	// 			"Status":       "Passed",
	// 		},
	// 	},
	// 	{
	// 		"$group": bson.M{
	// 			"_id":   "$JobVacancyID",
	// 			"Count": bson.M{"$sum": 1},
	// 		},
	// 	},
	// }

	// type pkwt struct {
	// 	JobID string `bson:"_id"`
	// 	Count int
	// }

	// pkwts := []pkwt{}
	// cmd := dbflex.From(new(hcmmodel.PKWTT).TableName()).Command("pipe", pipe)
	// if _, err := h.Populate(cmd, &pkwts); err != nil {
	// 	return nil, fmt.Errorf("err when get pkwt: %s", err.Error())
	// }

	// mapJob := lo.Associate(pkwts, func(p pkwt) (string, int) {
	// 	return p.JobID, p.Count
	// })

	// result := []codekit.M{}
	// for _, r := range requests {
	// 	// it means manpower is not replacement for 1 person
	// 	if r.AdditionalNumber != 0 {
	// 		// check if job still available or not
	// 		// by sum ClosedEmployeeTotal fill when job is closed & total candidate who pass pkwt stage
	// 		tmpTotal := r.ClosedEmployeeTotal + mapJob[r.ID]
	// 		if tmpTotal < r.AdditionalNumber {
	// 			result = append(result, codekit.M{
	// 				"_id":  r.ID,
	// 				"Name": r.Name,
	// 			})
	// 		}
	// 	} else {
	// 		// only need 1 person if manpower is replacement
	// 		if _, ok := mapJob[r.ID]; !ok {
	// 			result = append(result, codekit.M{
	// 				"_id":  r.ID,
	// 				"Name": r.Name,
	// 			})
	// 		}
	// 	}
	// }

	return requests, nil
}

type CloseRequest struct {
	JobID string
}

func (m *ManpowerHandler) CloseManpowerRequest(ctx *kaos.Context, payload *CloseRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	manpower := new(hcmmodel.ManpowerRequest)
	err := h.GetByID(manpower, payload.JobID)
	if err != nil {
		return nil, fmt.Errorf("err when get manpower request: %s", err.Error())
	}

	// get passed employee from pkwt
	pipe := []bson.M{
		{
			"$match": bson.M{
				"JobVacancyID": payload.JobID,
				"Status":       "Passed",
			},
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Count": bson.M{"$sum": 1},
			},
		},
	}

	type pkwt struct {
		Count int
	}

	pkwts := []pkwt{}
	cmd := dbflex.From(new(hcmmodel.PKWTT).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &pkwts); err != nil {
		return nil, fmt.Errorf("err when get pkwt: %s", err.Error())
	}

	var count int
	count, err = h.Count(new(hcmmodel.PKWTT), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("JobVacancyID", payload.JobID),
			dbflex.Eq("Status", "Passed"),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("err when get count pkwt: %s", err.Error())
	}

	// calculate ClosedEmployeeTotal because status always POSTED
	// so calculating ClosedEmployeeTotal to save remaining slot employee
	if manpower.AdditionalNumber != 0 {
		// manpower request for one or more employee
		manpower.ClosedEmployeeTotal = manpower.AdditionalNumber - count
	} else {
		// only replace 1 employee
		manpower.ClosedEmployeeTotal = 1 - count
	}

	manpower.IsClose = true
	err = h.Update(manpower, "ClosedEmployeeTotal", "IsClose")
	if err != nil {
		return nil, fmt.Errorf("err when update manpower: %s", err.Error())
	}

	return "success", nil
}
