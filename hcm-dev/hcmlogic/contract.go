package hcmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type ContractHandler struct {
}

func (m *ContractHandler) GetContracts(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	pipe := []bson.M{}
	if payload.Where != nil {
		match := bson.M{}
		if len(payload.Where.Items) > 0 {
			for _, item := range payload.Where.Items {
				match[item.Field] = bson.M{string(item.Op): item.Value}
			}
		} else {
			match[payload.Where.Field] = bson.M{string(payload.Where.Op): payload.Where.Value}
		}

		pipe = append(pipe, bson.M{"$match": match})
	}

	pipe = append(pipe, []bson.M{
		{
			"$project": bson.M{
				"RemainingProbation": bson.M{
					"$divide": bson.A{
						bson.M{"$subtract": bson.A{"$ExpiredContractDate", "$JoinedDate"}},
						1000 * 60 * 60 * 24,
					},
				},
				"EmployeeID":          1,
				"JobTitle":            1,
				"JoinedDate":          1,
				"ExpiredContractDate": 1,
				"Status":              1,
				"_id":                 1,
			},
		},
		{
			"$sort": bson.M{"RemainingProbation": 1},
		},
		{
			"$skip": payload.Skip,
		},
		{
			"$limit": payload.Take,
		},
	}...)

	contracts := []codekit.M{}
	cmd := dbflex.From(new(hcmmodel.Contract).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &contracts); err != nil {
		return nil, fmt.Errorf("err when get contract: %s", err.Error())
	}

	empID := make([]string, len(contracts))
	jobID := make([]string, len(contracts))
	for i, c := range contracts {
		empID[i] = c.GetString("EmployeeID")
		jobID[i] = c.GetString("JobTitle")
	}

	// get master employee
	employees := []tenantcoremodel.Employee{}
	err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", empID...),
		),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("err when get employee: %s", err.Error())
	}
	mapEmployee := lo.Associate(employees, func(source tenantcoremodel.Employee) (string, string) {
		return source.ID, source.Name
	})

	// get master
	masters := []tenantcoremodel.MasterData{}
	err = h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", jobID...),
		),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("err when get master: %s", err.Error())
	}
	mapMaster := lo.Associate(masters, func(source tenantcoremodel.MasterData) (string, string) {
		return source.ID, source.Name
	})

	count, err := h.Count(new(hcmmodel.Contract), dbflex.NewQueryParam().SetWhere(payload.Where))
	if err != nil {
		return nil, fmt.Errorf("err when get count contract: %s", err.Error())
	}

	for _, c := range contracts {
		c["EmployeeID"] = mapEmployee[c.GetString("EmployeeID")]
		c["JobTitle"] = mapMaster[c.GetString("JobTitle")]
	}

	return codekit.M{"count": count, "data": contracts}, nil
}
