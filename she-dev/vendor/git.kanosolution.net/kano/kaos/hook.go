package kaos

import (
	"fmt"
	"reflect"
)

func (sm *ServiceModel) RegisterHook(fn interface{}, name string) *ServiceModel {
	v := reflect.ValueOf(fn)
	t := v.Type()
	if v.Kind() != reflect.Func {
		return sm
	}
	if !(t.In(0).String() == "*kaos.Context" && t.NumOut() == 1 && t.Out(0).String() == "error") {
		return sm
	}
	sm.hooks[name] = v
	return sm
}

func (sm *ServiceModel) HasHook(name string) bool {
	_, ok := sm.hooks[name]
	return ok
}

func (sm *ServiceModel) CallHook(name string, ctx *Context, req interface{}) error {
	hook, ok := sm.hooks[name]
	if !ok {
		return nil
	}

	var e error
	func() {
		defer func() {
			if r := recover(); r != nil {
				ctx.Log().Errorf("%s.%s error: %v", sm.Name, name, r)
				e = fmt.Errorf("%s.%s error: %v", sm.Name, name, r)
			}
		}()
		outs := hook.Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(req)})
		o := outs[0].Interface()
		if o != nil {
			e = o.(error)
		}
	}()

	return e
}
