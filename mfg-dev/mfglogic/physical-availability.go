package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/golang-module/carbon/v2"
	"github.com/samber/lo"
)

type PhysicalAvailabilityEngine struct{}

type CalculateRequest struct {
	IDs []interface{}
}

func (o *PhysicalAvailabilityEngine) Calculate(ctx *kaos.Context, payload *CalculateRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	startMonth := carbon.Now("UTC").StartOfMonth()
	endMonth := carbon.Now("UTC").EndOfMonth()
	//get wo
	wos := []mfgmodel.WorkOrder{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrder), dbflex.In("_id", payload.IDs...), &wos)
	if e != nil {
		return nil, fmt.Errorf("Work Order not found")
	}

	woByEquipments := lo.GroupBy(wos, func(item mfgmodel.WorkOrder) string {
		return item.EquipmentNo
	})

	woByDayBreak := lo.MapEntries(woByEquipments, func(key string, wos []mfgmodel.WorkOrder) (string, float64) {
		breakDay := lo.SumBy(wos, func(item mfgmodel.WorkOrder) float64 {
			crbnStartTime := carbon.CreateFromStdTime(*item.StartDownTime)
			crbnFinishTime := carbon.CreateFromStdTime(*item.FinishDownTime)
			diffDays := crbnStartTime.DiffInDays(crbnFinishTime)
			if diffDays == 0 {
				return 1
			}

			return float64(diffDays)
		})

		phisicalAval := ((30 - breakDay) / 30) * 100
		return key, phisicalAval
	})

	h.DeleteByFilter(new(mfgmodel.PhysicalAvailability), dbflex.In("Unit", payload.IDs...))

	for _, id := range payload.IDs {
		idString := id.(string)
		physical, e := datahub.GetByID(h, new(mfgmodel.PhysicalAvailability), idString)
		if e != nil {
			return woByDayBreak, e
		}

		actual := 100.0
		if breakDay, ok := woByDayBreak[idString]; ok {
			actual = breakDay
		}

		physical.PAActual = actual
		physical.Month = startMonth.Day()
		physical.Year = endMonth.Year()

		e = h.Save(physical)
		if e != nil {
			return nil, e
		}
	}

	return woByDayBreak, nil
}
