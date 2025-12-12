package tenantcorelogic

import (
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
)

func TernaryDimension(dimSource tenantcoremodel.Dimension, dimCopies ...tenantcoremodel.Dimension) tenantcoremodel.Dimension {
	mapSource := lo.Associate(dimSource, func(item tenantcoremodel.DimensionItem) (string, string) {
		return item.Key, item.Value
	})

	for _, dim := range dimCopies {
		for _, item := range dim {
			if item.Value == "" {
				continue
			}

			v, ok := mapSource[item.Key]
			if !ok || v == "" {
				mapSource[item.Key] = item.Value
			}
		}
	}

	dimSource = lo.MapToSlice(mapSource, func(key string, value string) tenantcoremodel.DimensionItem {
		return tenantcoremodel.DimensionItem{Key: key, Value: value}
	})
	return dimSource
}

func TernaryString(vals ...string) string {
	res := ""
	for _, val := range vals {
		if val != "" && res == "" {
			return val
		}
	}
	return res
}

func TernaryTrxModule(vals ...tenantcoremodel.TrxModule) tenantcoremodel.TrxModule {
	res := ""
	for _, val := range vals {
		if val != "" && res == "" {
			return val
		}
	}
	return tenantcoremodel.TrxModule(res)
}

func TernaryCustom[T any](isBlank func(v T) bool, vals ...T) T {
	var res T
	if len(vals) > 0 {
		res = vals[0]
	}
	if len(vals) == 1 {
		return res
	}
	for _, val := range vals[1:] {
		if !isBlank(val) && isBlank(res) {
			return val
		}
	}
	return res
}
