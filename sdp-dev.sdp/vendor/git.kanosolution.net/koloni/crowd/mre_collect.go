package crowd

import (
	"reflect"
)

const (
	FnCollect    FnType = "collect"
	FnCollectMap FnType = "collectmap"
)

func (mre *MRE) Collect() *MRE {
	mre.funcs = append(mre.funcs, newFn(FnCollect))
	return mre
}

func (mre *MRE) CollectMap() *MRE {
	mre.funcs = append(mre.funcs, newFn(FnCollectMap))
	return mre
}

func (mre *MRE) handleCollect(input reflect.Value) (reflect.Value, error) {
	cap := mre.HeapSize()
	inputType := input.Type().Elem()
	sliceType := reflect.SliceOf(inputType)
	slice := reflect.MakeSlice(sliceType, cap, cap)
	chanOut := makeChannel(sliceType, reflect.BothDir, 0)

	mre.wg.Add(1)
	go func() {
		defer mre.handlePanic("collect", input, chanOut)
		idx := 0
		for {
			v, ok := input.Recv()
			if !ok {
				break
			}

			slice.Index(idx).Set(v)
			idx++
		}

		//-- reduce capacity to its value
		slice = trimSliceCap(slice, idx)
		chanOut.Send(slice)
	}()

	return chanOut, nil
}

func (mre *MRE) handleCollectMap(input reflect.Value) (reflect.Value, error) {
	var v reflect.Value
	chanOut := makeChannel(reflect.TypeOf(v), reflect.BothDir, 0)
	mre.wg.Add(1)
	go func() {
		//var chanOut reflect.Value
		defer mre.handlePanic("collectmap", input, chanOut)

		var vmap reflect.Value
		mapIsCreated := false
		for {
			//fmt.Println("\nget data")
			v, ok := input.Recv()
			if !ok {
				//fmt.Println("chan is not ok")
				break
			}
			vs, ok := v.Interface().([]reflect.Value)
			if !ok {
				//fmt.Println("cast is not ok")
				break
			}
			if len(vs) < 2 {
				//fmt.Println("lem chan is", len(vs))
				break
			}

			//fmt.Println("receive data:", vs)
			if !mapIsCreated {
				vmap = reflect.MakeMap(reflect.MapOf(vs[0].Type(), vs[1].Type()))
				//chanOut = makeChannel(vmap.Type(), reflect.BothDir, 0)
				mapIsCreated = true
				//fmt.Println("create map with type:", vmap.Type().String())
				//chanRes <- chanOut
			}

			vmap.SetMapIndex(vs[0], vs[1])
		}

		//fmt.Println("vmap:", vmap.Interface())
		chanOut.Send(reflect.ValueOf(vmap))
	}()

	return chanOut, nil
}
