package sebar

import (
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func ExtractDbmodGetsResult[T any](ctx *kaos.Context, model T) (codekit.M, []T, error) {
	empty := []T{}

	fnResGen := ctx.Data().Get("FnResult", nil)
	fnRes, ok := fnResGen.(codekit.M)
	if !ok {
		return fnRes, empty, fmt.Errorf("invalid result type, %t", fnResGen)
	}

	datasetPtr, ok := fnRes["data"].(*([]T))
	if !ok {
		return fnRes, empty, fmt.Errorf("invalid data result type, %t", fnRes["data"])
	}
	dataset := *datasetPtr
	return fnRes, dataset, nil
}

func ExtractDbmodFindResult[T any](ctx *kaos.Context, model T) ([]T, error) {
	empty := make([]T, 0)
	fnResGen := ctx.Data().Get("FnResult", nil)
	fnRes, ok := fnResGen.(*([]T))
	if !ok {
		return empty, fmt.Errorf("invalid result type: %t", fnResGen)
	}
	return *fnRes, nil
}
