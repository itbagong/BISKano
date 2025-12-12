package sdplogic_test

import (
	"fmt"
	"reflect"

	"git.kanosolution.net/kano/kaos"
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

// func insertJournal(ctx *kaos.Context, j *scmmodel.InventJournal) (*scmmodel.InventJournal, error) {
// 	j, err := callAPI(ctx, "/v1/scm/inventjournal/insert", j, j)
// 	return j, err
// }

// func insertPurchaseOrder(ctx *kaos.Context, po *scmmodel.PurchaseOrder) (*scmmodel.PurchaseOrder, error) {
// 	j, err := callAPI(ctx, "/v1/scm/purchaseorder/insert", po, po)
// 	return j, err
// }

// func insertPurchaseOrderDetail(ctx *kaos.Context, pods []*scmmodel.PurchaseOrderDetail) ([]*scmmodel.PurchaseOrderDetail, error) {
// 	lo.ForEach(pods, func(pod *scmmodel.PurchaseOrderDetail, index int) {
// 		callAPI(ctx, "/v1/scm/purchaseorder/detail/insert", pod, pod)

// 	})

// 	return pods, nil
// }
