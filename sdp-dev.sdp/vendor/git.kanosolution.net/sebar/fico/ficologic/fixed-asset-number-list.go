package ficologic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
)

type FixedAssetNumberListEngine struct{}

func (o *FixedAssetNumberListEngine) GetsFilter(ctx *kaos.Context, _ *interface{}) (*[]ficomodel.FixedAssetNumberList, error) {
	p := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	filters := []*dbflex.Filter{}
	if p["FixedAssetNumberID"] != "" {
		filters = append(filters, dbflex.Eq("FixedAssetNumberID", p["FixedAssetNumberID"]))
	}

	if p["FixedAssetGrup"] != "" {
		filters = append(filters, dbflex.Eq("FixedAssetGrup", p["FixedAssetGrup"]))
	}

	if p["GroupCode"] != "" {
		filters = append(filters, dbflex.Eq("GroupCode", p["GroupCode"]))
	}

	if isUsed, ok := p["IsUsed"]; ok {
		if isUsed == "true" {
			filters = append(filters, dbflex.Eq("IsUsed", true))
		} else {
			filters = append(filters, dbflex.Eq("IsUsed", false))
		}
	}

	filter := new(dbflex.Filter)
	if len(filters) > 0 {
		filter = dbflex.And(filters...)
	}

	res := []ficomodel.FixedAssetNumberList{}
	if e := h.GetsByFilter(new(ficomodel.FixedAssetNumberList), filter, &res); e != nil {
		return nil, e
	}

	return &res, nil
}

type FixedAssetNumberListUseRequest struct {
	IDs    []string
	IsUsed bool
}

func (o *FixedAssetNumberListEngine) Use(ctx *kaos.Context, payload *FixedAssetNumberListUseRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "nok", errors.New("missing: connection")
	}

	err := h.UpdateField(&ficomodel.FixedAssetNumberList{IsUsed: payload.IsUsed}, dbflex.In("_id", payload.IDs...), "IsUsed")
	return lo.Ternary(err == nil, "ok", "nok"), err
}
