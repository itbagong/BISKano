package crowd

import (
	"reflect"
	"sort"
)

func (mre *MRE) Sort(sortFn interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnSort).functions(sortFn))
	return mre
}

func (mre *MRE) handleSort(fnv, input reflect.Value) (reflect.Value, error) {
	fnt := fnv.Type()
	inputBaseType := input.Type().Elem()
	cap := mre.HeapSize()
	slice := reflect.MakeSlice(reflect.SliceOf(inputBaseType), cap, cap)
	chanOut := makeChannel(inputBaseType, reflect.BothDir, 0)

	mre.wg.Add(1)
	go func() {
		defer mre.handlePanic("sort", input, chanOut)

		idx := 0
		for {
			vs, ok, _ := prepareInputFromChannel(fnt, input, 0)
			//vs, ok := input.Recv()
			if !ok {
				break
			}

			//fmt.Printf("vs: %v\n", vs)
			slice.Index(idx).Set(vs[0])
			idx++
		}

		//fmt.Printf("\nbefore sort: %v\n", slice)
		slice = trimSliceCap(slice, idx)
		sliceSorter := &sorter{slice, fnv}
		sort.Sort(sliceSorter)
		slice = sliceSorter.data
		//fmt.Printf("\nafter sort: %v\n", slice)

		sl := slice.Len()
		for li := 0; li < sl; li++ {
			v := slice.Index(li)
			chanOut.Send(v)
		}
	}()

	return chanOut, nil
}

type sorter struct {
	data reflect.Value
	fn   reflect.Value
}

func (s *sorter) Len() int {
	return s.data.Len()
}

func (s *sorter) Swap(i, j int) {
	newslice := reflect.MakeSlice(s.data.Type(), 2, 2)
	newslice.Index(0).Set(s.data.Index(i))
	newslice.Index(1).Set(s.data.Index(j))
	//fmt.Printf("\n0. o: %v\nn: %v", s.data, newslice)

	s.data.Index(i).Set(newslice.Index(1))
	s.data.Index(j).Set(newslice.Index(0))
}

func (s *sorter) Less(i, j int) bool {
	parms := []reflect.Value{s.data.Index(i), s.data.Index(j)}
	less := s.fn.Call(parms)
	return less[0].Bool()
}
