package tenantcoremodel

import (
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type Dimension []DimensionItem

type DimensionItem struct {
	Key   string
	Value string
}

func (dims Dimension) ToString() string {
	res := []string{}
	for _, v := range dims {
		if v.Value != "" {
			txt := fmt.Sprintf("%s=%s", v.Key, v.Value)
			res = append(res, txt)
		}
	}
	return strings.Join(res, " | ")
}

func (dims Dimension) Hash() string {
	source := "no-dimension"
	if len(dims) > 0 {
		source = strings.Join(lo.Map(dims, func(d DimensionItem, _ int) string {
			return fmt.Sprintf("%s=%s", d.Key, d.Value)
		}), "&&")
	}
	return codekit.ShaString(source, "")
}

func (dims Dimension) ToMap() map[string]string {
	return lo.SliceToMap(dims, func(d DimensionItem) (string, string) { return d.Key, d.Value })
}

func (dim Dimension) Where() *dbflex.Filter {
	if len(dim) == 0 {
		return nil
	}

	fs := make([]*dbflex.Filter, len(dim))
	for index, item := range dim {
		fs[index] = dbflex.ElemMatch("Dimension", dbflex.Eq("Key", item.Key), dbflex.Eq("Value", item.Value))
	}

	return dbflex.And(fs...)
}

func (dim Dimension) Compare(otherDim Dimension) bool {
	for _, d := range dim {
		if d.Value == "" {
			continue
		}
		found := false
	iterOther:
		for _, od := range otherDim {
			if od.Key == d.Key {
				if od.Value != d.Value {
					return false
				}
				found = true
				break iterOther
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (dim Dimension) Set(key, val string) Dimension {
	for index, d := range dim {
		if d.Key == key {
			d.Value = val
			dim[index] = d
			return dim
		}
	}

	dim = append(dim, DimensionItem{Key: key, Value: val})
	return dim
}

func (dim Dimension) Sets(parts ...string) Dimension {
	var (
		key string
	)
	for _, v := range parts {
		if key == "" {
			key = v
		} else {
			dim = dim.Set(key, v)
			key = ""
		}
	}
	return dim
}

func (dim Dimension) Get(key string) string {
	for _, d := range dim {
		if d.Key == key {
			return d.Value
		}
	}

	return ""
}

func (dim Dimension) Keys() []string {
	res := make([]string, len(dim))
	for idx, d := range dim {
		res[idx] = d.Key
	}
	return res
}
