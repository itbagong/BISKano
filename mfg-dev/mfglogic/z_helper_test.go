package mfglogic_test

import (
	"fmt"
	"reflect"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
)

func callAPI[T any](ctxMain *kaos.Context, path string, payload interface{}, dest T) (T, error) {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return dest, fmt.Errorf("target should be pointer")
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

func insertWorkOrder(ctx *kaos.Context, j *mfgmodel.WorkOrderJournal) (*mfgmodel.WorkOrderJournal, error) {
	j, err := callAPI(ctx, "/v1/mfg/work-order-journal/insert", j, j)
	return j, err
}
