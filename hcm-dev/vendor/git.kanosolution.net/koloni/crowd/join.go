package crowd

import (
	"fmt"
	"reflect"
)

func Join(s1, s2 Source, joinType JoinType, joiner, mapper, output interface{}) error {
	vj := reflect.ValueOf(joiner)
	vm := reflect.ValueOf(mapper)
	vo := reflect.ValueOf(output)

	if vj.Kind() != reflect.Func {
		return fmt.Errorf("joiner should be function and not %v", vj.Kind())
	}

	if vm.Kind() != reflect.Func {
		return fmt.Errorf("mapper should be function and not %v", vm.Kind())
	}

	if !isPtr(vo.Type()) || !isSlice(vo.Type()) {
		return fmt.Errorf("output should be a pointer of slice")
	}

	cap := s1.Count() * s2.Count()
	mapout := reflect.MakeSlice(vo.Type().Elem(), cap, cap)
	idx := 0

	s1.Reset()
	for {
		v1, end1 := s1.Next()
		if end1 {
			break
		}
		hasJoin := false

		s2.Reset()
		for {
			v2, end2 := s2.Next()
			if end2 {
				break
			}

			joinParms := []reflect.Value{v1, v2}
			joinRes := vj.Call(joinParms)
			if len(joinRes) == 0 {
				return fmt.Errorf("join function output should be > 0")
			}

			if joinRes[0].Bool() {
				hasJoin = true
				mapRes := vm.Call(joinParms)
				if len(mapRes) == 0 {
					return fmt.Errorf("mapper function output should be > 0")
				}
				mapout.Index(idx).Set(mapRes[0])
				idx++
			}
		}

		if !hasJoin && joinType == LeftJoin {
			joinParms := []reflect.Value{v1, reflect.Zero(s2.ElementType())}
			mapRes := vm.Call(joinParms)
			if len(mapRes) == 0 {
				return fmt.Errorf("mapper function output should be > 0")
			}
			mapout.Index(idx).Set(mapRes[0])
			idx++
		}
	}

	mapout = trimSliceCap(mapout, idx)
	vo.Elem().Set(mapout)
	return nil
}

func isSlice(v reflect.Type) bool {
	if isPtr(v) {
		return v.Elem().Kind() == reflect.Slice
	}

	return v.Kind() == reflect.Slice
}

func isPtr(v reflect.Type) bool {
	return v.Kind() == reflect.Ptr
}
