package kaos

import (
	"path"
	"reflect"
	"strings"
)

func (s *Service) RegisterRoute(fn interface{}, name string) *ServiceRoute {
	if s.classicRoutes == nil {
		s.classicRoutes = make(map[string]*ServiceRoute)
	}

	vm := reflect.ValueOf(fn)
	tm := reflect.TypeOf(fn)
	if vm.Kind() != reflect.Func {
		return nil
	}

	if !(tm.NumIn() == 2 &&
		((tm.In(0).String() == "*kaos.Context" && tm.NumOut() == 2 && tm.Out(1).String() == "error") ||
			(tm.In(0).String() == "kaos.EventHub" && tm.NumOut() == 1 && tm.Out(0).String() == "error"))) {
		return nil
	}

	epName := EndPointName(name)
	sr := new(ServiceRoute)
	sr.model = nil
	sr.Path = path.Join(s.BasePoint(), epName)
	sr.Path = strings.ReplaceAll(sr.Path, "\\", "/")
	sr.Fn = vm
	sr.DefaultHubName = "default"
	sr.RequestType = tm.In(1)
	sr.ResponseType = tm.Out(0)
	s.classicRoutes[epName] = sr

	return sr
}
