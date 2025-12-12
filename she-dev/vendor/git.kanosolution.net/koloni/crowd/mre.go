package crowd

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024
)

var defaultHeapSize int

func SetDefaultHeapSize(n int) {
	defaultHeapSize = n
}

func DefaultHeapSize() int {
	if defaultHeapSize == 0 {
		defaultHeapSize = MB
	}
	return defaultHeapSize
}

type MRE struct {
	funcs  []*fn
	source Source

	log      *logger.LogEngine
	wg       *sync.WaitGroup
	err      error
	heapSize int
}

func (mre *MRE) SetHeapSize(n int) *MRE {
	mre.heapSize = n
	return mre
}

func (mre *MRE) HeapSize() int {
	if mre.heapSize == 0 {
		mre.heapSize = DefaultHeapSize()
	}
	return mre.heapSize
}

func (mre *MRE) SetLogger(l *logger.LogEngine) *MRE {
	mre.log = l
	return mre
}

func (mre *MRE) Log() *logger.LogEngine {
	if mre.log == nil {
		mre.log, _ = logger.NewLog(true, false, "", "", "")
	}
	return mre.log
}

func (mre *MRE) NoStdOut() *MRE {
	mre.Log().LogToStdOut = false
	return mre
}

func (mre *MRE) Source() Source {
	return mre.source
}

func (mre *MRE) Each(fn interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnEach).functions(fn))
	return mre
}

func (mre *MRE) Filter(fn interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnFilter).functions(fn))
	return mre
}

func (mre *MRE) Map(fn interface{}) *MRE {
	mre.funcs = append(mre.funcs, newFn(FnMap).functions(fn))
	return mre
}

func (mre *MRE) MapFields(fields ...string) *MRE {
	fn := func(obj interface{}) codekit.M {
		m := codekit.M{}
		om, err := codekit.ToM(obj)
		if err != nil {
			//-- do nothing
		} else {
			for _, field := range fields {
				m.Set(field, om[field])
			}
		}
		return m
	}

	mre.funcs = append(mre.funcs, newFn(FnMap).functions(fn))
	return mre
}

func (mre *MRE) Reduce(accu interface{}, fn interface{}) *MRE {
	mre.funcs = append(mre.funcs,
		newFn(FnReduce).
			functions(fn).
			values(reflect.Indirect(reflect.ValueOf(accu))))
	//Fn{FnReduce, fn,
	//reflect.Indirect(reflect.ValueOf(accu))})
	return mre
}

func (mre *MRE) handlePanic(title string, inputChannel reflect.Value, channels ...reflect.Value) {
	if rec := recover(); rec != nil {
		txt := fmt.Sprintf("panic on %s: %s", title, rec)
		mre.Log().Error(txt)
		if mre.err == nil {
			traces := strings.Split(string(debug.Stack()), "\n")
			mretraces := []string{}
			for _, trace := range traces {
				//if strings.Contains(trace, "crowd") {
				mretraces = append(mretraces, trace)
				//}
			}
			mre.err = errors.New(txt + "\nTraces:\n" + strings.Join(mretraces, "\n"))
		}
		closeChannel(inputChannel)
	}

	if mre.wg != nil {
		mre.wg.Done()
	}

	for _, channel := range channels {
		closeChannel(channel)
	}
}

func closeChannel(cv reflect.Value) {
	defer func() {
		_ = recover()
	}()

	cv.Close()
}

func trimSliceCap(vs reflect.Value, cap int) reflect.Value {
	n := reflect.MakeSlice(vs.Type(), cap, cap)
	reflect.Copy(n, vs)
	return n
}
