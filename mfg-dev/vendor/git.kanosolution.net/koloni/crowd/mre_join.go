package crowd

import (
	"fmt"
	"reflect"
)

type JoinType int

const (
	FnJoin FnType = "join"

	InnerJoin   JoinType = 1
	LeftJoin    JoinType = 2
	RightJoin   JoinType = 3
	FullJoin    JoinType = 4
	InvalidJoin JoinType = 5
)

func (mre *MRE) Join(data *MRE, joinType JoinType, joiner, mapper interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnJoin).
		functions(joiner, mapper).
		values(reflect.ValueOf(data), reflect.ValueOf(joinType)))
	return mre
}

func (mre *MRE) handleJoin(fn *fn, input reflect.Value) (reflect.Value, error) {
	wg := mre.wg
	if len(fn.fns) < 2 {
		return reflect.Value{}, fmt.Errorf("join need joiner and mapper functions")
	}

	if len(fn.vs) == 0 {
		return reflect.Value{}, fmt.Errorf("join need joiner MRE")
	}

	vj := reflect.ValueOf(fn.fns[0])
	tj := vj.Type()
	if tj.NumOut() == 0 || tj.Out(0).String() != "bool" {
		return reflect.Value{}, fmt.Errorf("joiner should return bool")
	}

	vm := reflect.ValueOf(fn.fns[1])
	tm := vm.Type()

	joinMRE := fn.vs[0].Interface().(*MRE)
	var joinMREOutType, joinMREOutElementType reflect.Type
	//-- prepare output of joinMRE, it needs to be a slice of something
	if len(joinMRE.funcs) == 0 {
		//-- if funcs is zero, return slice of source
		joinMREOutElementType = joinMRE.Source().ElementType()
		joinMREOutType = reflect.SliceOf(joinMREOutElementType)
	} else {
		//-- if has funcs get latest
		joinFnIdx := len(joinMRE.funcs) - 1
		var vjoinFnOut reflect.Type
		for {
			lastfn := joinMRE.funcs[joinFnIdx]
			tlastfn := reflect.TypeOf(lastfn)
			if tlastfn.NumOut() > 0 {
				vjoinFnOut = tlastfn.Out(0)
			}
			joinFnIdx--
			if joinFnIdx < 0 {
				break
			}
		}
		//-- nil, all functions does not have func. Make slice of source elemtn type
		if vjoinFnOut == nil {
			joinMREOutElementType = joinMRE.Source().ElementType()
			joinMREOutType = reflect.SliceOf(joinMREOutElementType)
		} else {
			joinMREOutElementType = vjoinFnOut
			joinMREOutType = reflect.SliceOf(vjoinFnOut)
		}
	}
	//fmt.Println("Preparing join mre")
	joinMREOut := reflect.New(joinMREOutType)
	joinMREerr := joinMRE.Run(joinMREOut.Interface())
	if joinMREerr != nil {
		return reflect.Value{}, fmt.Errorf("error reading joined MRE. %s", joinMREerr.Error())
	}

	//fmt.Printf("join mre: %v", joinMREOut.Elem().Interface())
	joinMREOut = joinMREOut.Elem()
	chanOut := makeChannel(tm.Out(0), reflect.BothDir, 0)
	joinType := JoinType(fn.vs[1].Int())

	wg.Add(1)
	go func() {
		//defer wg.Done()
		defer mre.handlePanic("join", input, chanOut)

		joinMREOutCount := joinMREOut.Len()
		for {
			vs, ok, _ := prepareInputFromChannel(tj, input, 0)
			if !ok {
				break
			}

			hasJoin := false
			for i := 0; i < joinMREOutCount; i++ {
				joinData := joinMREOut.Index(i)
				joinParms := []reflect.Value{vs[0], joinData}
				//fmt.Printf("\njoin parms: %v\n")

				var mparms []reflect.Value
				send := true
				jouts := vj.Call(joinParms)
				if jouts[0].Bool() == true {
					mparms = joinParms
					hasJoin = true

				} else {
					send = false
				}

				if send {
					mouts := vm.Call(mparms)
					//fmt.Printf("join send: %v\n", mouts[0].Interface())
					chanOut.Send(mouts[0])
				}
			}

			if !hasJoin && joinType == LeftJoin {
				mparms := []reflect.Value{vs[0], reflect.Zero(joinMREOutElementType)}
				mouts := vm.Call(mparms)
				chanOut.Send(mouts[0])
				//fmt.Printf("leftjoin send: %v\n", mouts[0].Interface())
			}
		}
		//chanOut.Close()
		//fmt.Println("ok")
	}()
	return chanOut, nil
}
