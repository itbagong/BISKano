package crowd

import (
	"errors"
	"reflect"
)

type SliceSource struct {
	slice reflect.Value
	err   error
	index int
	count int
}

func NewSliceSource(data interface{}) *SliceSource {
	ss := new(SliceSource)
	ss.from(data)
	return ss
}

func (ss *SliceSource) from(source interface{}) error {
	rv := reflect.Indirect(reflect.ValueOf(source))
	rt := rv.Type()

	if rt.Kind() != reflect.Slice {
		ss.err = errors.New("Source is not a slice")
		return ss.err
	}

	ss.index = -1
	ss.slice = rv
	ss.count = rv.Len()
	return nil
}

func (ss *SliceSource) ElementType() reflect.Type {
	return ss.slice.Type().Elem()
}

func (ss *SliceSource) Reset() {
	ss.index = -1
}

func (ss *SliceSource) Count() int {
	return ss.count
}

// return value and bool to identify already at end of the slice
func (ss *SliceSource) Next() (reflect.Value, bool) {
	var v reflect.Value

	ss.index++
	if ss.count == 0 {
		return v, true
	}

	if ss.index >= ss.count {
		return ss.slice.Index(ss.count - 1), true
	}

	return ss.slice.Index(ss.index), false
}

func (ss *SliceSource) Position() int {
	return ss.index
}
