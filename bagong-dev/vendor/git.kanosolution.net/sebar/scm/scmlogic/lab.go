package scmlogic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type Lab struct {
}

type ItemBalanceRequest struct {
	BalanceDate *time.Time
	Opts        ItemBalanceOpt
}

func (obj *Lab) GetItemBalance(ctx *kaos.Context, payload *ItemBalanceRequest) ([]*scmmodel.ItemBalance, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: database")
	}

	if payload == nil {
		payload = new(ItemBalanceRequest)
		payload.Opts = ItemBalanceOpt{
			CompanyID: "DEMO00",
			ItemIDs:   []string{"BAN_LUAR"},
			GroupBy:   []string{"WarehouseID"},
		}
	}

	bal := NewItemBalanceHub(db)
	bals, err := bal.Get(payload.BalanceDate, payload.Opts)
	if err != nil {
		return nil, err
	}
	return bals, nil
}

func (obj *Lab) SyncItemBalance(ctx *kaos.Context, payload *ItemBalanceOpt) ([]*scmmodel.ItemBalance, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: database")
	}

	if payload == nil {
		payload = new(ItemBalanceOpt)
	}
	if payload.CompanyID == "" {
		payload.CompanyID, _ = GetCompanyIDFromContext(ctx)
	}
	if payload.CompanyID == "" {
		payload.CompanyID = "DEMO00"
	}
	payload.DisableGrouping = true

	bal := NewItemBalanceHub(db)
	return bal.Sync(nil, *payload)
}

func (obj *Lab) FindUniqueDims(ctx *kaos.Context, payload ItemBalanceOpt) ([]string, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: database")
	}

	return FindInventDimIDs(db, payload.CompanyID, payload.ItemIDs[0], "InventDim", payload.InventDim)
}
