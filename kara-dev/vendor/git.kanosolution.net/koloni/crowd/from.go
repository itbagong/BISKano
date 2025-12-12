package crowd

import (
	"errors"
	"reflect"

	"github.com/sebarcode/codekit"
)

func FromSlice(s interface{}) *MRE {
	MRE := new(MRE)
	MRE.source = NewSliceSource(s)
	MRE.SetHeapSize(codekit.SliceLen(s))
	return MRE
}

func FromMap(m interface{}) *MRE {
	mre := new(MRE)

	vm := reflect.ValueOf(m)
	if vm.Kind() == reflect.Ptr {
		vm = vm.Elem()
	}
	vt := vm.Type()
	if vm.Kind() != reflect.Map {
		mre.err = errors.New("source should be a map")
	}

	keys := vm.MapKeys()
	slice := reflect.MakeSlice(reflect.SliceOf(vt.Elem()), len(keys), len(keys))
	for idx, key := range keys {
		item := vm.MapIndex(key)
		slice.Index(idx).Set(item)
	}

	return FromSlice(slice.Interface())
}
