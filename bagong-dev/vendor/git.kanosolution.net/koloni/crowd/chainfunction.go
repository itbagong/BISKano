package crowd

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func ChainFunctions(
	input interface{},
	lastOutput interface{},
	functions ...interface{}) (*sync.WaitGroup, error) {
	inValue := reflect.ValueOf(input)
	inType := inValue.Type()

	if inType.Kind() != reflect.Chan {
		return nil, errors.New("Input should be a channel")
	}

	wg := new(sync.WaitGroup)
	var err error
	inChannel := inValue
	for idx, fn := range functions {
		inChannel, err = wrapFn(fn, inChannel, wg)
		if err != nil {
			return nil, fmt.Errorf("Function %d error: %s", idx, err.Error())
		}
	}

	if inChannel.Kind() != reflect.Invalid && lastOutput != nil {
		wg.Add(1)
		lastMutex := struct {
			m     sync.Mutex
			value reflect.Value
		}{value: reflect.Indirect(reflect.ValueOf(lastOutput))}
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			var lastValue reflect.Value
			//fmt.Println("handling last channel")
			for ok := true; ok; {
				//fmt.Println("receiving last channel")
				lastValue, ok = inChannel.Recv()
				if ok {
					lastMutex.m.Lock()
					if lastValue.Kind() == reflect.Ptr {
						lastMutex.value.Set(reflect.Indirect(lastValue))
					} else {
						lastMutex.value.Set(lastValue)
					}
					lastMutex.m.Unlock()
				}
			}
		}(wg)
	}

	return wg, nil
}

// wrapFn will auto make channel for chained functions
func wrapFn(fn interface{}, input reflect.Value, wg *sync.WaitGroup) (reflect.Value, error) {
	if wg != nil {
		wg.Add(1)
	}
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return reflect.Value{}, errors.New("Fn parameter should be a function")
	}

	if input.Type().Kind() != reflect.Chan {
		return reflect.Value{}, errors.New("input should be a channel")
	}

	var outputChan reflect.Value

	hasOutput := false
	//useContext := fnType.NumIn() > 0 && fnType.In(0).Name() == "*crowd.Context"

	if fnType.NumOut() > 0 {
		hasOutput = true
		outputChan = makeChannel(fnType.Out(0), reflect.BothDir, 0)
	}

	go func(wg *sync.WaitGroup) {
		if wg != nil {
			defer wg.Done()
		}
		var inputVal reflect.Value
		for ok := true; ok; {
			if inputVal, ok = input.Recv(); ok {
				var inparms []reflect.Value
				inparms = append(inparms, inputVal)

				outputs := fnValue.Call(inparms)
				if hasOutput {
					//fmt.Printf("sending: %v \n", codekit.JsonString(outputs[0].Interface()))
					outputChan.Send(outputs[0])
				}
			}
		}

		if hasOutput {
			outputChan.Close()
		}
	}(wg)

	return outputChan, nil
}
