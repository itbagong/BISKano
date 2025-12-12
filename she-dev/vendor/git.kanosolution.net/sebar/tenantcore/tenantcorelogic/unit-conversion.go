package tenantcorelogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type UnitConversionEngine struct{}

type UnitConversionSaveMultipleRequest struct {
	FromUnit        string
	UnitConversions []tenantcoremodel.UnitConversion
}

func (o *UnitConversionEngine) SaveMultiple(ctx *kaos.Context, payload *UnitConversionSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if e := h.DeleteByFilter(new(tenantcoremodel.UnitConversion), dbflex.Eq("FromUnit", payload.FromUnit)); e != nil {
		return nil, errors.New("error clear item details: " + e.Error())
	}

	for _, uc := range payload.UnitConversions {
		uc.FromUnit = payload.FromUnit
		if e := h.Save(&uc); e != nil {
			return nil, errors.New("error update Movement  Detail: " + e.Error())
		}

		reverse := new(tenantcoremodel.UnitConversion)
		h.GetByFilter(reverse, dbflex.And(dbflex.Eq("FromUnit", uc.ToUnit), dbflex.Eq("ToUnit", uc.FromUnit)))
		reverse.FromUnit = uc.ToUnit
		reverse.ToUnit = uc.FromUnit
		reverse.ToQty = 1 / uc.ToQty
		h.Save(reverse)
	}

	return payload, nil
}
