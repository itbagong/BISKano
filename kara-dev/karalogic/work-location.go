package karalogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"go.mongodb.org/mongo-driver/bson"
)

type WorkLocationHandler struct {
}

type WorkLocation struct {
	ID                string
	Name              string
	Enable            bool
	Virtual           bool
	AcceptNonRule     bool
	TimeLoc           string
	DistanceTolerance float32
	Address           string
	Longitude         float64
	Latitude          float64
	Created           time.Time
	LastUpdate        time.Time
}

func (m *WorkLocationHandler) Create(ctx *kaos.Context, payload *WorkLocation) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	data := karamodel.WorkLocation{
		ID:                payload.ID,
		Name:              payload.Name,
		Enable:            payload.Enable,
		Virtual:           payload.Virtual,
		AcceptNonRule:     payload.AcceptNonRule,
		TimeLoc:           payload.TimeLoc,
		DistanceTolerance: payload.DistanceTolerance,
		Address:           payload.Address,
		Location: &karamodel.Location{
			Type:        "Point",
			Coordinates: []float64{payload.Longitude, payload.Latitude},
		},
		Created:    payload.Created,
		LastUpdate: payload.LastUpdate,
	}

	if err := h.Insert(&data); err != nil {
		return nil, fmt.Errorf("error when save work location: %s", err.Error())
	}

	return data, nil
}

type FindNearestLocation struct {
	Latitude  float64
	Longitude float64
}

func (m *WorkLocationHandler) FindNearestLocation(ctx *kaos.Context, payload *FindNearestLocation) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	coordinates := []float64{payload.Longitude, payload.Latitude}
	pipe := []bson.M{
		{
			"$geoNear": bson.M{
				"near": bson.M{
					"type":        "Point",
					"coordinates": coordinates,
				},
				"spherical":     true,
				"distanceField": "Distance",
			},
		},
		{
			"$limit": 1,
		},
	}

	// get count
	locations := []karamodel.WorkLocation{}
	cmd := dbflex.From(new(karamodel.WorkLocation).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &locations); err != nil {
		return nil, fmt.Errorf("err when get locations: %s", err.Error())
	}

	return locations, nil
}
