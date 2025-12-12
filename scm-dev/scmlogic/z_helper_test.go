package scmlogic_test

import (
	"fmt"
	"reflect"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
)

func callAPI[T any](ctxMain *kaos.Context, path string, payload interface{}, dest T) (T, error) {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return dest, fmt.Errorf("target should be a pointer")
	}
	sr := svc.GetRoute(path)
	if sr == nil {
		return dest, fmt.Errorf("invalid route: %s", path)
	}
	ctx := kaos.NewContextFromService(svc, sr)
	for _, key := range ctxMain.Data().Keys() {
		ctx.Data().Set(key, ctxMain.Data().Get(key, nil))
	}
	err := svc.CallTo(path, dest, ctx, payload)
	return dest, err
}

func insertJournal(ctx *kaos.Context, j *scmmodel.InventJournal) (*scmmodel.InventJournal, error) {
	j, err := callAPI(ctx, "/v1/scm/inventjournal/insert", j, j)
	return j, err
}

func insertPurchaseOrder(ctx *kaos.Context, po *scmmodel.PurchaseOrderJournal) (*scmmodel.PurchaseOrderJournal, error) {
	j, err := callAPI(ctx, "/v1/scm/purchase/order/insert", po, po)
	return j, err
}

func insertItemRequest(ctx *kaos.Context, ir *scmmodel.ItemRequest) (*scmmodel.ItemRequest, error) {
	j, err := callAPI(ctx, "/v1/scm/purchase/order/insert", ir, ir)
	return j, err
}
