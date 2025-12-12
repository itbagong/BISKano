package tenantcorelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type UnitEngine struct{}

func (o *UnitEngine) GetsFilter(ctx *kaos.Context, payload *GeneralRequest) ([]tenantcoremodel.UoM, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	req := GetURLQueryParams(ctx)
	qfs := []*dbflex.Filter{}

	if req["ItemID"] != "" {
		item := new(tenantcoremodel.Item)
		if e := h.GetByID(item, req["ItemID"]); e != nil {
			return nil, fmt.Errorf("item not found: %s | e: %s", req["ItemID"], e.Error())
		}

		u := new(tenantcoremodel.UoM)
		h.GetByID(u, item.DefaultUnitID)

		if u.UOMCategory != "" {
			qfs = append(qfs, dbflex.Eq("UOMCategory", u.UOMCategory))
		}
	}

	parm := payload.GetQueryParam()
	if parm.Where != nil {
		qfs = append(qfs, parm.Where)
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
