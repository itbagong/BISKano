package tenantcorelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
)

type UnitEngine struct{}

func (o *UnitEngine) GetsFilter(ctx *kaos.Context, payload *GeneralRequest) ([]tenantcoremodel.UoM, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		payload = &GeneralRequest{
			Skip: 0,
			Take: 10,
		}
	}

	req := GetURLQueryParams(ctx)
	qfs := []*dbflex.Filter{}

	if req["ItemID"] != "" {
		item := new(tenantcoremodel.Item)
		if e := h.GetByID(item, req["ItemID"]); e != nil {
			return nil, fmt.Errorf("item not found: %s | e: %s", req["ItemID"], e.Error())
		}

		convers := []tenantcoremodel.UnitConversion{}
		h.GetsByFilter(new(tenantcoremodel.UnitConversion), dbflex.Eq("FromUnit", item.DefaultUnitID), &convers)

		converIDs := lo.Map(convers, func(d tenantcoremodel.UnitConversion, i int) string {
			return d.ToUnit
		})

		if item.DefaultUnitID != "" {
			converIDs = append(converIDs, item.DefaultUnitID)
		}

		if len(converIDs) > 0 {
			qfs = append(qfs, dbflex.In("_id", converIDs...))
		}
	}

	parm := payload.GetQueryParam()
	if parm.Where != nil {
		qfs = append(qfs, parm.Where)
	}

	if req["_id"] != "" {
		qfs = append(qfs, dbflex.Eq("_id", req["_id"]))
	}

	if len(qfs) > 0 {
		parm.SetWhere(dbflex.And(qfs...))
	}

	res := []tenantcoremodel.UoM{}
	if e := h.Gets(new(tenantcoremodel.UoM), parm, &res); e != nil {
		return nil, e
	}

	return res, nil
}
