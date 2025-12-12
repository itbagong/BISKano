package scmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type MovementOutDetailEngine struct{}

func (o *MovementOutDetailEngine) SaveMultiple(ctx *kaos.Context, payload []scmmodel.MovementOutDetail) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if len(payload) == 0 {
		return nil, errors.New("missing: payload")
	}

	mov := new(scmmodel.MovementOut)
	if e := h.GetByID(mov, payload[0].MovementOutID); e != nil {
		return nil, errors.New("no movement out data found: " + e.Error())
	}

	// validation: itemID and SKU can't be double
	pm := map[string]bool{}
	for _, p := range payload {
		uniq := fmt.Sprintf("%s||%s", p.ItemID, p.SKU)

		if exist, ok := pm[uniq]; ok && exist {
			return nil, fmt.Errorf("duplication exist for item id '%s' and sku '%s'", p.ItemID, p.SKU)
		}

		pm[uniq] = true
	}

	// multiple saving
	h.DeleteByFilter(new(scmmodel.MovementOutDetail), dbflex.Eq("MovementOutID", payload[0].MovementOutID))
	for _, p := range payload {
		p.MovementOutID = mov.ID
		p.InventoryDimension = mov.InventoryDimension
		if e := h.Save(&p); e != nil {
			return nil, errors.New("error update Item Spec: " + e.Error())
		}
	}

	return payload, nil
}
