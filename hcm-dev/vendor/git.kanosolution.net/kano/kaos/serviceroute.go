package kaos

import (
	"fmt"
	"reflect"

	"github.com/sebarcode/codekit"
)

type ServiceRoute struct {
	Path               string
	Fn                 reflect.Value
	RequestType        reflect.Type
	ResponseType       reflect.Type
	DefaultHubName     string
	ExcludeMiddlewares []string

	model   interface{}
	data    *SharedData
	mws     []*MWItem
	postMws []*MWItem
}

func (sr *ServiceRoute) Model() interface{} {
	return sr.model
}

func (sr *ServiceRoute) Data() *SharedData {
	if sr.data == nil {
		sr.data = NewSharedData()
	}
	return sr.data
}

func (sr *ServiceRoute) Call(inputs []reflect.Value) []reflect.Value {
	ctx, ok := inputs[0].Interface().(*Context)
	if !ok {
		return sr.Error(fmt.Errorf("context is not properly initialized"))
	}
	parm := inputs[1].Interface()

	for _, mw := range ctx.mws {
		if isok, e := mw.Fn(ctx, parm); e != nil {
			sr.Error(fmt.Errorf("middleware %s fail to run: %s", mw.Name, e.Error()))
		} else if !isok {
			var nilErr error
			var nilValue interface{}

			return []reflect.Value{
				reflect.ValueOf(nilValue),
				reflect.ValueOf(nilErr),
			}
		}
	}
	return sr.Fn.Call(inputs)
}

func (sr *ServiceRoute) Error(err error) []reflect.Value {
	var rt reflect.Type
	if sr.ResponseType != nil {
		rt = sr.ResponseType
	} else {
		rt = sr.Fn.Type().Out(0)
	}
	var res reflect.Value
	if rt.String()[0] == '*' {
		res = reflect.New(rt.Elem())
	} else {
		res = reflect.Zero(rt)
	}
	return []reflect.Value{res, reflect.ValueOf(err)}
}

func (sr *ServiceRoute) Merge(other *ServiceRoute) *ServiceRoute {
	if other.DefaultHubName != "" {
		sr.DefaultHubName = other.DefaultHubName
	}

	if other.RequestType != nil {
		sr.RequestType = other.RequestType
	}

	if other.ResponseType != nil {
		sr.ResponseType = other.ResponseType
	}
	for k, v := range other.Data().data {
		sr.Data().Set(k, v)
	}
	for _, mw := range other.ExcludeMiddlewares {
		if !codekit.HasMember(sr.ExcludeMiddlewares, mw) {
			sr.ExcludeMiddlewares = append(sr.ExcludeMiddlewares, mw)
		}
	}
	return sr
}

func (sr *ServiceRoute) Middlewares() []*MWItem {
	return sr.mws
}

func (sr *ServiceRoute) PostMiddlewares() []*MWItem {
	return sr.postMws
}

func (sr *ServiceRoute) Run(ctx *Context, s *Service, parm interface{}) (interface{}, error) {
	ctx.Data().Set("RoutePath", sr.Path)
	mws := []*MWItem{}
	postMws := []*MWItem{}
	if len(sr.ExcludeMiddlewares) == 0 {
		mws = sr.Middlewares()
		postMws = sr.PostMiddlewares()
	} else {
		for _, mw := range sr.Middlewares() {
			if codekit.HasMember(sr.ExcludeMiddlewares, mw.Name) {
				continue
			}
			mws = append(mws, mw)
		}

		for _, mw := range sr.PostMiddlewares() {
			if codekit.HasMember(sr.ExcludeMiddlewares, mw.Name) {
				continue
			}
			postMws = append(mws, mw)
		}
	}

	for _, mw := range mws {
		if ok, e := mw.Fn(ctx, parm); e != nil {
			return nil, fmt.Errorf("mod %s error: %s", mw.Name, e.Error())
		} else if !ok {
			return nil, nil
		}
	}

	ins := make([]reflect.Value, 2)
	ins[0] = reflect.ValueOf(ctx)
	v1 := reflect.ValueOf(parm)
	if parm == nil {
		ins[1] = reflect.Zero(sr.RequestType)
	} else {
		ins[1] = v1
	}
	o := sr.Fn.Call(ins)

	if !o[1].IsNil() {
		return o[0].Interface(), o[1].Interface().(error)
	}

	ctx.Data().Set("FnResult", o[0].Interface())
	for _, mw := range postMws {
		if ok, e := mw.Fn(ctx, parm); e != nil {
			return nil, fmt.Errorf("mod %s error: %s", mw.Name, e.Error())
		} else if !ok {
			return nil, nil
		}
	}

	return ctx.data.data["FnResult"], nil
}
