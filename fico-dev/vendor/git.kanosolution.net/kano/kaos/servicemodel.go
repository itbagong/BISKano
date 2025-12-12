package kaos

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/sebarcode/codekit"
)

type ServiceModel struct {
	Name      string
	Model     interface{}
	ModelType reflect.Type

	//modExcludes   []string
	//mwExcludes    []string
	alias   map[string]string
	mods    []Mod
	mws     []*MWItem
	postMws []*MWItem
	hooks   map[string]reflect.Value

	deployers []string

	disableRoutes   []string
	allowOnlyRoutes []string
	hubName         string
	evName          string
}

func NewServiceModel(model interface{}, name string) *ServiceModel {
	sm := new(ServiceModel)
	sm.Name = name
	sm.Model = model
	sm.alias = map[string]string{}

	sm.hooks = make(map[string]reflect.Value)
	rv := reflect.ValueOf(model)
	if rv.Kind() == reflect.Ptr {
		sm.ModelType = rv.Type().Elem()
	} else {
		sm.ModelType = rv.Type()
	}

	sm.hubName = "default"
	sm.evName = "default"
	return sm
}

func (sm *ServiceModel) DisableRoute(names ...string) *ServiceModel {
	sm.disableRoutes = names
	return sm
}

func (sm *ServiceModel) DisableRoutes() []string {
	return sm.disableRoutes
}

func (sm *ServiceModel) SetAlias(method, alias string) *ServiceModel {
	if sm.alias == nil {
		sm.alias = make(map[string]string)
	}
	sm.alias[method] = alias
	return sm
}

func (sm *ServiceModel) SetFullAlias(alias map[string]string) *ServiceModel {
	sm.alias = alias
	return sm
}

func (sm *ServiceModel) AllowOnlyRoute(names ...string) *ServiceModel {
	sm.allowOnlyRoutes = names
	return sm
}

func (sm *ServiceModel) AllowOnlyRoutes() []string {
	return sm.allowOnlyRoutes
}

func (sm *ServiceModel) SetDeployer(deployers ...string) *ServiceModel {
	sm.deployers = append(sm.deployers, deployers...)
	return sm
}

func (sm *ServiceModel) CanDeployTo(deployer string) bool {
	if len(sm.deployers) == 0 {
		return true
	}

	for _, deployerItem := range sm.deployers {
		if deployerItem == deployer {
			return true
		}
	}

	return false
}

func (sm *ServiceModel) SetHubName(name string) *ServiceModel {
	sm.hubName = name
	return sm
}

func (sm *ServiceModel) HubName() string {
	return sm.hubName
}

func (sm *ServiceModel) SetEventName(name string) *ServiceModel {
	sm.evName = name
	return sm
}

func (sm *ServiceModel) EventName() string {
	return sm.evName
}

func (sm *ServiceModel) SetMod(mods ...Mod) *ServiceModel {
	sm.mods = mods
	return sm
}

func (sm *ServiceModel) Mods() []Mod {
	return sm.mods
}

func (sm *ServiceModel) RegisterMW(fn MWFunc, name string) *ServiceModel {
	found := false
	for idx, m := range sm.mws {
		if m.Name == name {
			found = true
			sm.mws[idx] = &MWItem{name, fn}
		}
	}
	if !found {
		sm.mws = append(sm.mws, &MWItem{name, fn})
	}
	return sm
}

func (sm *ServiceModel) RegisterPostMW(fn MWFunc, name string) *ServiceModel {
	found := false
	for idx, m := range sm.postMws {
		if m.Name == name {
			found = true
			sm.postMws[idx] = &MWItem{name, fn}
		}
	}
	if !found {
		sm.postMws = append(sm.postMws, &MWItem{name, fn})
	}
	return sm
}

func (sm *ServiceModel) RegisterMWs(fns ...MWFunc) *ServiceModel {
	for _, fn := range fns {
		fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
		if fnName == "" {
			fnName = "midware function " + codekit.RandomString(5)
		}
		if strings.Contains(fnName, "/") {
			names := strings.Split(fnName, "/")
			fnName = names[len(names)-1]
		}
		sm.mws = append(sm.mws, &MWItem{fnName, fn})
	}
	return sm
}

func (sm *ServiceModel) RegisterPostMWs(fns ...MWFunc) *ServiceModel {
	for _, fn := range fns {
		sm.postMws = append(sm.postMws, &MWItem{"fn-" + codekit.RandomString(5), fn})
	}
	return sm
}

func (sm *ServiceModel) Middlewares() []*MWItem {
	return sm.mws
}

func (sm *ServiceModel) PostMiddlewares() []*MWItem {
	return sm.postMws
}

func (sm *ServiceModel) PrepareRoute(svc *Service) ([]*ServiceRoute, error) {
	if sm.Model == nil {
		return []*ServiceRoute{}, nil
	}

	rv := reflect.ValueOf(sm.Model)
	rt := rv.Type()
	alias, hasAlias := sm.alias[sm.Name]
	if !hasAlias {
		alias = sm.Name
	}
	fnCount := rv.NumMethod()

	extensionDisableRoutes := map[string]bool{}
	extensionRoutes := map[string]*ServiceRoute{}

	for _, name := range sm.DisableRoutes() {
		extensionDisableRoutes[name] = true
	}

	for i := 0; i < rv.NumMethod(); i++ {
		vf := rv.Method(i)
		tf := rt.Method(i)
		fnName := EndPointName(tf.Name)

		// extension routes
		if fnName == "MapRoutes" && tf.Type.NumOut() == 1 && tf.Type.Out(0) == reflect.TypeOf(map[string]*ServiceRoute{}) {
			outs := vf.Call([]reflect.Value{})
			extensionRoutes = outs[0].Interface().(map[string]*ServiceRoute)
		}

		// extension disable routes from each model if exists
		if fnName == "DisableRoutes" && tf.Type.NumOut() == 1 && tf.Type.Out(0) == reflect.TypeOf([]string{}) {
			outs := vf.Call([]reflect.Value{})
			for _, name := range outs[0].Interface().([]string) {
				extensionDisableRoutes[name] = true
			}
		}
	}

	routes := []*ServiceRoute{}
	for fnIdx := 0; fnIdx < fnCount; fnIdx++ {
		vm := rv.Method(fnIdx)
		tm := rt.Method(fnIdx)
		tmOpt := tm.Type

		if !(tmOpt.NumIn() == 3 &&
			((tmOpt.In(1).String() == "*kaos.Context" && tmOpt.NumOut() == 2 && tmOpt.Out(1).String() == "error") ||
				(tmOpt.In(1).String() == "kaos.EventHub" && tmOpt.NumOut() == 1 && tmOpt.Out(0).String() == "error"))) {
			continue
		}

		lcaseRouteName := EndPointName(tm.Name)
		_, disable := extensionDisableRoutes[lcaseRouteName]
		if disable {
			continue
		}

		if len(sm.allowOnlyRoutes) > 0 {
			if !codekit.HasMember(sm.allowOnlyRoutes, lcaseRouteName) {
				continue
			}
		}

		sr := new(ServiceRoute)
		sr.model = sm.Model
		sr.Path = path.Join(svc.BasePoint(), alias, lcaseRouteName)
		sr.Path = strings.ReplaceAll(sr.Path, "\\", "/")
		sr.Fn = vm
		sr.DefaultHubName = "default"
		sr.RequestType = tmOpt.In(2)
		sr.ResponseType = tmOpt.Out(0)
		if len(extensionRoutes) > 0 {
			if esr, ok := extensionRoutes[tm.Name]; ok {
				sr.Merge(esr)
			}
		}
		routes = append(routes, sr)
	}

	mods := svc.mods
	mods = append(mods, sm.mods...)

	for _, mod := range mods {
		if modRoutes, e := mod.MakeModelRoute(svc, sm); e == nil {
			//routes = append(routes, modRoutes...)
			for idx, mr := range modRoutes {
				mr.Path = strings.ReplaceAll(mr.Path, "\\", "/")
				modRoutes[idx].Path = mr.Path
				names := strings.Split(mr.Path, "/")
				name := names[len(names)-1]
				name = EndPointName(name)

				lcaseRouteName := name
				_, disable := extensionDisableRoutes[lcaseRouteName]
				if disable {
					continue
				}

				if len(sm.allowOnlyRoutes) > 0 {
					if !codekit.HasMember(sm.allowOnlyRoutes, lcaseRouteName) {
						continue
					}
				}

				//mr.Path = mr.Path
				routes = append(routes, mr)
			}
		} else {
			return routes, fmt.Errorf("unable to create route for service %s model %s mod %s: %s",
				svc.basePoint, sm.Name, mod.Name(), e.Error())
		}
	}

	for _, r := range routes {
		r.mws = svc.mws
		r.mws = append(r.mws, sm.mws...)
		r.postMws = append(r.postMws, sm.postMws...)
	}

	// apply alias
	for _, r := range routes {
		routePaths := strings.Split(r.Path, "/")
		lastPath := routePaths[len(routePaths)-1]
		alias, hasAlias := sm.alias[lastPath]
		if !hasAlias {
			continue
		}
		if hasAlias {
			routePaths[len(routePaths)-1] = alias
			r.Path = strings.Join(routePaths, "/")
		}
	}

	return routes, nil
}
