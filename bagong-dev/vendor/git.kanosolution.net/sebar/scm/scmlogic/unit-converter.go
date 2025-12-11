package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

func ConvertUnit(h *datahub.Hub, qtyFrom float64, unitIDFrom string, unitIDTo string) (float64, error) {
	if qtyFrom == 0 {
		return 0, nil
	}

	if unitIDFrom == "" && unitIDTo != "" {
		unitIDFrom = unitIDTo
	}

	if unitIDTo == "" {
		unitIDTo = unitIDFrom
	}

	if unitIDFrom == unitIDTo {
		return qtyFrom, nil
	}

	cv := new(tenantcoremodel.UnitConversion)
	if e := h.GetByFilter(cv, dbflex.And(
		dbflex.Eq("FromUnit", unitIDFrom),
		dbflex.Eq("ToUnit", unitIDTo),
	)); e != nil {
		return 0, fmt.Errorf("no conversion for '%s' to '%s' | err: %s", unitIDFrom, unitIDTo, e)
	}

	if cv.ID == "" {
		return 0, fmt.Errorf("no conversion for '%s' to '%s'", unitIDFrom, unitIDTo)
	}

	if cv.ToQty == 0 {
		return 0, fmt.Errorf("no 'To Qty' set for conversion '%s' to '%s'", unitIDFrom, unitIDTo)
	}

	return qtyFrom * cv.ToQty, nil
}

func MoreThanDefaultUnit(h *datahub.Hub, qty1 float64, unitID1 string, qty2 float64, itemID string) (isMore bool, defaultUnit tenantcoremodel.UoM, err error) {
	item := new(tenantcoremodel.Item)
	if e := h.GetByID(item, itemID); e != nil {
		err = fmt.Errorf("MoreThanUnit: could not get item: %s", e.Error())
		return
	}

	defaultUnit = tenantcoremodel.UoM{}
	h.GetByID(&defaultUnit, item.DefaultUnitID)

	isMore, err = MoreThanUnit(h, qty1, unitID1, qty2, item.DefaultUnitID, itemID)
	return
}

func MoreThanUnit(h *datahub.Hub, qty1 float64, unitID1 string, qty2 float64, unitID2 string, itemID string) (bool, error) {
	var res bool

	item := new(tenantcoremodel.Item)
	if e := h.GetByID(item, itemID); e != nil {
		return res, fmt.Errorf("MoreThanUnit: could not get item: %s", e.Error())
	}

	defaultUnitID := item.DefaultUnitID

	qtyDef1, e := ConvertUnit(h, qty1, unitID1, defaultUnitID)
	if e != nil {
		return res, e
	}

	qtyDef2, e := ConvertUnit(h, qty2, unitID2, defaultUnitID)
	if e != nil {
		return res, e
	}

	return moreThan(qtyDef1, qtyDef2, true), nil
}
