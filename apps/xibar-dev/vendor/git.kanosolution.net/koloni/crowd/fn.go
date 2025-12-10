package crowd

import "reflect"

type FnType string

const (
	FnEach    FnType = "each"
	FnMap     FnType = "map"
	FnFilter  FnType = "filter"
	FnReduce  FnType = "reduce"
	FnSum     FnType = "sum"
	FnAvg     FnType = "avg"
	FnCount   FnType = "count"
	FnMin     FnType = "min"
	FnMax     FnType = "max"
	FnSort    FnType = "sort"
	FnGeneral FnType = "general"
)

type fn struct {
	fnType FnType
	fns    []interface{}
	vs     []reflect.Value
}

func newFn(t FnType) *fn {
	f := new(fn)
	f.fnType = t
	return f
}

func (fn *fn) functions(fs ...interface{}) *fn {
	fn.fns = append(fn.fns, fs...)
	return fn
}

func (fn *fn) values(vs ...reflect.Value) *fn {
	fn.vs = append(fn.vs, vs...)
	return fn
}
