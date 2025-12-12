package tenantcoremodel

import "github.com/sebarcode/codekit"

type ReferenceItem struct {
	Key   string
	Value interface{}
}

type References []ReferenceItem

func (r References) Set(k string, v interface{}) References {
	for index, item := range r {
		if item.Key == k {
			item.Value = v
			r[index] = item
			return r
		}
	}

	r = append(r, ReferenceItem{Key: k, Value: v})
	return r
}

func (r References) Get(k string, def interface{}) interface{} {
	for _, item := range r {
		if item.Key == k {
			return item.Value
		}
	}
	return def
}

func (r References) ToM() codekit.M {
	res := codekit.M{}

	for _, i := range r {
		res[i.Key] = i.Value
	}

	return res
}

func MergeReference(refs ...References) References {
	res := References{}
	for _, reference := range refs {
		for _, item := range reference {
			res = res.Set(item.Key, item.Value)
		}
	}
	return res
}
