package crowd

import (
	"reflect"
)

const (
	FnGroup FnType = "group"
)

func (mre *MRE) Group(fn interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnGroup).functions(fn))
	return mre
}

func (mre *MRE) handleGroup(fnval, input reflect.Value) (chanOut reflect.Value, err error) {
	mre.wg.Add(1)
	values := []reflect.Value{}
	chanOut = makeChannel(reflect.TypeOf(values), reflect.BothDir, 0)
	inputType := input.Type().Elem()
	outputType := fnval.Type().Out(0)

	vmap := reflect.MakeMap(reflect.MapOf(outputType, reflect.SliceOf(inputType)))

	go func() {
		defer mre.handlePanic("group", input, chanOut)
		cap := mre.HeapSize()
		sizes := map[interface{}]int{}
		for {
			v, ok := input.Recv()
			if !ok {
				break
			}

			outputs := fnval.Call([]reflect.Value{v})
			var size int
			slice := vmap.MapIndex(outputs[0])
			size = sizes[outputs[0].Interface()]
			if !slice.IsValid() {
				slice = reflect.MakeSlice(reflect.SliceOf(inputType), cap, cap)
			}
			slice.Index(size).Set(v)
			vmap.SetMapIndex(outputs[0], slice)
			sizes[outputs[0].Interface()] = size + 1
		}

		keys := vmap.MapKeys()
		for _, k := range keys {
			vmapItem := vmap.MapIndex(k)
			size := sizes[k.Interface()]
			vmapItem = trimSliceCap(vmapItem, size)
			outvalues := []reflect.Value{k, vmapItem}
			cout := reflect.ValueOf(outvalues)
			chanOut.Send(cout)
		}
	}()

	return
}
