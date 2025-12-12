package rbacmodel

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type DimensionItem struct {
	Key   string
	Value string
}

type Dimension []*DimensionItem

func (d Dimension) Set(key, value string) Dimension {
	for _, v := range d {
		if v.Key == key {
			v.Value = value
			return d
		}
	}

	d = append(d, &DimensionItem{
		Key:   key,
		Value: value,
	})

	return d
}

func (d Dimension) Get(key string) string {
	for _, v := range d {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
}

func (d Dimension) Hash() string {
	if len(d) == 0 {
		return ""
	}

	res := ""
	for _, di := range d {
		res += fmt.Sprintf("|%s=%s", di.Key, di.Value)
	}

	return codekit.MD5String(res)
}

func (d Dimension) MultiHash() []string {
	res := []string{}

	dims := Dimension{}
	for _, di := range d {
		dims.Set(di.Key, di.Value)
		res = append(res, dims.Hash())
	}

	return res
}

func (d Dimension) SetupCompare(other Dimension) bool {
	if len(d) > len(other) {
		return false
	}

	if len(d) == 0 {
		return true
	}

	for _, do := range d {
		v := other.Get(do.Key)
		if v != do.Value {
			return false
		}
	}

	return true
}

func (d Dimension) ObjectCompare(setup Dimension) bool {
	if len(d) < len(setup) {
		return false
	}

	if len(setup) == 0 {
		return true
	}

	for _, do := range setup {
		v := d.Get(do.Key)
		if v != do.Value {
			return false
		}
	}

	return true
}

func (d Dimension) DbWhere() *dbflex.Filter {
	if len(d) == 0 {
		return nil
	}

	dimGroup := lo.GroupBy(d, func(item *DimensionItem) string {
		return item.Key
	})

	ws := make([]*dbflex.Filter, len(dimGroup))
	for _, val := range dimGroup {
		df := make([]*dbflex.Filter, len(val))
		for idx, di := range val {
			df[idx] = dbflex.ElemMatch("Dimension", dbflex.Eq("Key", di.Key), dbflex.Eq("Value", di.Value))
		}
		ws = append(ws, dbflex.Or(df...))
	}

	return dbflex.And(ws...)
}
