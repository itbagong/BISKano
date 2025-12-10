package crowd

import (
	"reflect"

	"github.com/sebarcode/codekit"
)

const (
	FnSubset FnType = "subset"
)

func (mre *MRE) First() *MRE {
	mre.funcs = append(mre.funcs, newFn(FnGeneral).functions(func(input reflect.Value) (reflect.Value, error) {
		inputType := input.Type().Elem()
		mre.wg.Add(1)
		var err error
		chanOut := makeChannel(inputType, reflect.BothDir, 0)

		go func() {
			defer mre.handlePanic("First", input, chanOut)

			v, ok := input.Recv()
			if !ok {
				return
			}
			chanOut.Send(v)
		}()

		return chanOut, err
	}))
	return mre
}

func (mre *MRE) Last() *MRE {
	mre.funcs = append(mre.funcs, newFn(FnGeneral).functions(func(input reflect.Value) (reflect.Value, error) {
		inputType := input.Type().Elem()
		mre.wg.Add(1)
		var err error
		chanOut := makeChannel(inputType, reflect.BothDir, 0)

		go func() {
			defer mre.handlePanic("First", input, chanOut)

			var last reflect.Value
			for {
				v, ok := input.Recv()
				if !ok {
					break
				}

				last = v
			}
			chanOut.Send(last)
		}()

		return chanOut, err
	}))
	return mre
}

func (mre *MRE) Subset(from, to int) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnSubset).values(reflect.ValueOf(from), reflect.ValueOf(to)))
	return mre
}

func (mre *MRE) handleSubset(from, to int, input reflect.Value) (reflect.Value, error) {
	inputType := input.Type().Elem()
	mre.wg.Add(1)
	var err error
	chanOut := makeChannel(inputType, reflect.BothDir, 0)

	go func() {
		defer mre.handlePanic("subset", input, chanOut)

		i := 0
		for {
			v, ok := input.Recv()
			if !ok {
				break
			}

			if i >= from && i <= to {
				chanOut.Send(v)
			}
			i++
		}
	}()

	return chanOut, err
}

func (mre *MRE) Sum() *MRE {
	mre.funcs = append(mre.funcs, newFn(FnSum))
	return mre
}

func (mre *MRE) handleSum(input reflect.Value) (reflect.Value, error) {
	var err error
	mre.wg.Add(1)
	inputType := input.Type().Elem()
	inputIsInt := isInt(inputType)
	inputIsFloat := isFloat(inputType)

	chanOut := makeChannel(inputType, reflect.BothDir, 0)
	int0 := int64(0)
	float0 := float64(0)
	txt := ""

	go func() {
		defer mre.handlePanic("sum", input, chanOut)
		for {
			v, ok := input.Recv()
			if !ok {
				break
			}

			if inputIsInt {
				int0 += v.Int()
			} else if inputIsFloat {
				float0 += v.Float()
			} else {
				txt += codekit.ToString(v.Interface())
			}
		}

		if inputIsInt {
			chanOut.Send(reflect.ValueOf(codekit.ToInt(int0, codekit.RoundingAuto)))
		} else if inputIsFloat {
			if inputType == reflect.TypeOf(float32(0)) {
				chanOut.Send(reflect.ValueOf(codekit.ToFloat32(float0, 10, codekit.RoundingAuto)))
			} else {
				chanOut.Send(reflect.ValueOf(codekit.ToFloat64(float0, 10, codekit.RoundingAuto)))
			}
		} else {
			chanOut.Send(reflect.ValueOf(txt))
		}
	}()

	return chanOut, err
}
