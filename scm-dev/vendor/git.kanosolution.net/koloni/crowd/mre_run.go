package crowd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/sebarcode/codekit"
)

func (mre *MRE) Exec() (interface{}, error) {
	var res interface{}
	err := mre.Run(&res)
	return res, err
}

func (mre *MRE) Run(final interface{}) error {
	defer func() {
		mre.source = nil
	}()

	var err error
	mre.wg = new(sync.WaitGroup)
	cdata := makeChannel(mre.source.ElementType(), reflect.BothDir, 0)

	cvinput := cdata

	if len(mre.funcs) > 0 {
		//-- if there is funcs
		for idx, fn := range mre.funcs {
			cvinput, err = mre.wrapMREFn(fn, cvinput)
			if err != nil {
				err = fmt.Errorf("unable to exec MRE element %d. %s", idx, err.Error())
				return err
			}
		}
	} else {
		//-- if no funcs, run to slice
		cvinput, err = mre.wrapMREFn(newFn(FnCollect), cvinput)
		if err != nil {
			err = fmt.Errorf("unable to exec MRE collect. %s", err.Error())
			return err
		}
	}

	//-- last output
	if cvinput.Kind() != reflect.Invalid {
		mre.wg.Add(1)
		go func() {
			defer mre.handlePanic("last output", cvinput)

			lastValIsRV := false
			lastValTypeChecked := false

			for ok := true; ok; {
				vs, ok, _ := valuesFromChannel(cvinput, 0)
				if !ok {
					break
				}

				if !lastValTypeChecked {
					lastValTypeChecked = true
					lastValIsRV = reflect.TypeOf(vs[0].Interface()).String() == "reflect.Value"
				}

				if lastValIsRV {
					reflect.ValueOf(final).Elem().Set(vs[0].Interface().(reflect.Value))
				} else {
					if vs[0].Kind() == reflect.Ptr {
						reflect.ValueOf(final).Elem().Set(vs[0].Elem())
					} else {
						reflect.ValueOf(final).Elem().Set(vs[0])
					}
				}
			}
		}()
	}

	//-- send the source as value
	for {
		v, b := mre.source.Next()
		if b {
			break
		}
		//codekit.Println("\nsending: ", mre.source.Position(), v.Interface())

		ableToSend := true
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					ableToSend = false
				}
			}()
			cdata.Send(v)
		}()
		if !ableToSend {
			break
		}
	}
	closeChannel(cdata)

	mre.wg.Wait()
	return mre.err
}

func (mre *MRE) wrapMREFn(fn *fn, input reflect.Value) (reflect.Value, error) {
	wg := mre.wg
	var (
		fnType  reflect.Type
		fnValue reflect.Value
	)

	if !codekit.HasMember([]FnType{FnSum, FnCollect, FnCollectMap, FnSubset}, fn.fnType) {
		fnValue = reflect.ValueOf(fn.fns[0])
		fnType = fnValue.Type()

		fnKind := fnType.Kind()
		if fnKind != reflect.Func {
			return reflect.Value{}, fmt.Errorf("Fn parameter should be a function, instead got %s", fnKind)
		}
	}

	inputKind := input.Kind()
	if inputKind != reflect.Chan {
		return reflect.Value{}, fmt.Errorf("input should be a channel, instead got %s", inputKind)
	}

	switch fn.fnType {
	case FnGeneral:
		if fn, ok := fn.fns[0].(func(reflect.Value) (reflect.Value, error)); ok {
			return fn(input)
		}
		mre.handlePanic("general", input)
		return input, fmt.Errorf("error on serializing general function")

	case FnSum:
		return mre.handleSum(input)

	case FnFilter:
		return mre.handleFilter(fnValue, input)

	case FnReduce:
		return mre.handleReduce(fnValue, fn.vs[0], input)

	case FnGroup:
		return mre.handleGroup(fnValue, input)

	case FnJoin:
		return mre.handleJoin(fn, input)

	case FnCollect:
		return mre.handleCollect(input)

	case FnCollectMap:
		return mre.handleCollectMap(input)

	case FnSort:
		return mre.handleSort(fnValue, input)

	case FnSubset:
		return mre.handleSubset(int(fn.vs[0].Int()), int(fn.vs[1].Int()), input)
	}

	//-- taken care of each, general and map
	wg.Add(1)
	//-- this is for classic Map Reduce Filter
	var outputChan reflect.Value
	hasOutput := false
	if fnType.NumOut() == 0 {
		outputChan = makeChannel(input.Type().Elem(), reflect.BothDir, 0)
	} else if fnType.NumOut() == 1 {
		hasOutput = true
		outputChan = makeChannel(fnType.Out(0), reflect.BothDir, 0)
	} else {
		hasOutput = true
		outputChan = makeChannel(reflect.SliceOf(reflect.ValueOf(outputChan).Type()), reflect.BothDir, 0)
	}

	//-- map
	go func(wg *sync.WaitGroup) {
		defer mre.handlePanic(string(fn.fnType), input, outputChan)
		var (
			inputVal reflect.Value
			outputs  []reflect.Value
		)

		for {
			inparms, ok, _ := prepareInputFromChannel(fnType, input, 0)
			if !ok {
				break
			}
			/*fmt.Printf("type: %s inparms: %v\n",
			input.Type().Elem().String(),
			inparms[0].Interface())
			*/
			outputs = fnValue.Call(inparms)
			if hasOutput {
				if len(outputs) == 1 {
					outputChan.Send(outputs[0])
				} else {
					outputChan.Send(reflect.ValueOf(outputs))
				}
			} else {
				outputChan.Send(inputVal)
			}
		}
	}(wg)

	return outputChan, nil
}

func isInt(t reflect.Type) bool {
	tstring := strings.ToLower(t.String())
	ok, _ := codekit.MemberIndex([]string{"int", "uint",
		"int8", "int16", "int32", "int64",
		"uint8", "uint16", "uint32", "uint64"}, tstring)
	return ok
}

func isFloat(t reflect.Type) bool {
	tstring := strings.ToLower(t.String())
	ok, _ := codekit.MemberIndex([]string{"float32", "float64"}, tstring)
	return ok
}

func (mre *MRE) handleFilter(fn reflect.Value, input reflect.Value) (reflect.Value, error) {
	var err error

	chanOut := makeChannel(input.Type().Elem(), reflect.BothDir, 0)
	mre.wg.Add(1)
	go func() {
		defer mre.handlePanic("filter", input, chanOut)
		//defer mre.wg.Done()

		//mre.Log().Infof("filtering data")
		for {
			v, ok, _ := prepareInputFromChannel(fn.Type(), input, 0)
			if !ok {
				break
			}

			outs := fn.Call(v)
			if outs[0].Bool() {
				chanOut.Send(v[0])
			}
		}
		//chanOut.Close()
	}()

	return chanOut, err
}

func (mre *MRE) handleReduce(fn, accuHolder, input reflect.Value) (chanOut reflect.Value, err error) {
	var outputs []reflect.Value
	hasOutput := fn.Type().NumOut() > 0
	if hasOutput {
		chanOut = makeChannel(fn.Type().Out(0), reflect.BothDir, 0)
	}

	rvReducer := reflect.Indirect(reflect.New(accuHolder.Type()))
	rvReducer.Set(reflect.Indirect(accuHolder))

	mre.wg.Add(1)
	go func() {
		defer mre.handlePanic("reduce", input, chanOut)

		for {
			inparms, ok, _ := prepareInputFromChannel(fn.Type(), input, 1)
			if !ok {
				break
			}
			inparms = append(inparms, rvReducer)
			outputs = fn.Call(inparms)
			rvReducer.Set(reflect.Indirect(outputs[0]))
		}
		if hasOutput {
			chanOut.Send(rvReducer)
		}
		//chanOut.Close()
	}()

	return
}

func prepareInputFromChannel(fnt reflect.Type, input reflect.Value, offset int) ([]reflect.Value, bool, error) {
	inputCount := fnt.NumIn() - offset
	if inputCount < 1 {
		return []reflect.Value{}, false, fmt.Errorf("function %s has no input", fnt.Name())
	}

	return valuesFromChannel(input, inputCount)
}

func valuesFromChannel(input reflect.Value, inputCount int) ([]reflect.Value, bool, error) {
	vs, ok := input.Recv()
	if !ok {
		return []reflect.Value{}, false, errors.New("EOF")
	}

	//-- if vs is slice of values, means: prev command returned more than one value
	if vs.Type().String() == "[]reflect.Value" {
		values := []reflect.Value{}

		l := vs.Len()
		i := 0

		for {
			var v reflect.Value
			v = vs.Index(i).Interface().(reflect.Value)

			//fmt.Printf("v: %v\n", v)
			values = append(values, v)
			i++

			//fmt.Println(i, l, inputCount, values)
			if l != 0 && i == l || i == inputCount {
				break
			}
		}

		//fmt.Println("is a slice channel")
		//fmt.Printf("values: %v\n", values[0])
		return values, true, nil
	}

	//-- no slice, return single index parm
	return []reflect.Value{vs}, true, nil
}
