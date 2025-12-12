package karalogic

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
)

type ShiftAssignment struct {
}
type GenerateShiftRequest struct {
	LocationId string
	StartDate  time.Time
	EndDate    time.Time
}
type GenerateShiftResponse struct {
}

func (s *ShiftAssignment) AssignShift(ctx *kaos.Context, payload *GenerateShiftRequest) (*GenerateShiftResponse, error) {
	response := GenerateShiftResponse{}
	hub, err := ctx.DefaultHub()
	if err != nil {
		return nil, err
	}
	location := karamodel.WorkLocation{}
	err = hub.GetByID(&location, payload.LocationId)
	if err != nil {
		return nil, err
	}
	err = GenerateShift(hub, payload.StartDate, payload.EndDate, location)
	return &response, err
}
