package kaos

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/appkit"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/logger"
)

type Service struct {
	ctx    context.Context
	models []*ServiceModel
	mods   []Mod
	//evmodels map[string]*EventModel
	//hubs map[string]*datahub.Hub
	hubMgr *HubManager

	mws       []*MWItem
	postMWS   []*MWItem
	basePoint string
	log       *logger.LogEngine
	data      *SharedData // data
	routes    map[string]*ServiceRoute
	evHubs    map[string]EventHub
	hubFn     GetHubFn

	classicRoutes map[string]*ServiceRoute
}

func NewService() *Service {
	s := new(Service)
	s.hubMgr = NewHubManager(func(key, group string) (*datahub.Hub, error) {
		return nil, errors.New("missing: HubBuilder")
	})
	s.basePoint = "/"
	s.data = NewSharedData()
	s.evHubs = map[string]EventHub{}
	//s.evmodels = map[string]*EventModel{}
	return s
}

func (s *Service) Log() *logger.LogEngine {
	if s.log == nil {
		s.log = appkit.Log()
	}
	return s.log
}

func (s *Service) SetLogger(l *logger.LogEngine) *Service {
	s.log = l
	return s
}

func (s *Service) Data() *SharedData {
	if s.data == nil {
		s.data = NewSharedData()
	}
	return s.data
}

func (s *Service) Context() context.Context {
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	return s.ctx
}

func (s *Service) SetBasePoint(bp string) *Service {
	s.basePoint = bp
	if !strings.HasPrefix(bp, "/") {
		s.basePoint = "/" + s.basePoint
	}
	if !strings.HasSuffix(bp, "/") {
		s.basePoint += "/"
	}
	return s
}

func (s *Service) GetDataHub(name, group string) (*datahub.Hub, error) {
	if s.hubMgr == nil {
		return nil, errors.New("missing: HubManager")
	}
	return s.hubMgr.Get(name, group)
}

func (s *Service) BasePoint() string {
	return s.basePoint
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Models() []*ServiceModel {
	return s.models
}

func (s *Service) RegisterModel(model interface{}, name string) *ServiceModel {
	sm := NewServiceModel(model, name)
	s.AddServiceModel(sm)
	return sm
}

func (s *Service) AddServiceModel(sm *ServiceModel) {
	s.models = append(s.models, sm)
}

func (s *Service) ServiceModel(name string) *ServiceModel {
	name = strings.ToLower(name)
	for _, sm := range s.models {
		if strings.ToLower(sm.Name) == name {
			return sm
		}
	}
	return nil
}

func (s *Service) Mods() []Mod {
	return s.mods
}

func (s *Service) SetMod(mods ...Mod) *Service {
	s.mods = mods
	return s
}

func (s *Service) HubManager() *HubManager {
	return s.hubMgr
}

func (s *Service) SetHubManager(mgr *HubManager) *Service {
	s.hubMgr = mgr
	return s
}

func (s *Service) PrepareRoutes(deployerName string) ([]*ServiceRoute, error) {
	routes := []*ServiceRoute{}

	// model routes
	for _, model := range s.Models() {
		if !model.CanDeployTo(deployerName) {
			continue
		}
		svcRoutes, e := model.PrepareRoute(s)
		if e != nil {
			return routes, e
		}
		routes = append(routes, svcRoutes...)
	}

	// mod routes
	for _, mod := range s.mods {
		svcRoutes, e := mod.MakeGlobalRoute(s)
		if e != nil {
			return routes, e
		}
		routes = append(routes, svcRoutes...)
	}

	// classic routes
	for _, sr := range s.classicRoutes {
		sr.mws = s.mws
		sr.mws = append(sr.mws, s.mws...)
		sr.postMws = append(sr.postMws, s.postMWS...)
		routes = append(routes, sr)
	}

	s.routes = make(map[string]*ServiceRoute, len(routes))
	sdata := s.Data().Data()
	for _, r := range routes {
		for k, v := range sdata {
			r.Data().Set("service_"+k, v)
		}
		s.routes[r.Path] = r
	}
	return routes, nil
}

func (s *Service) GetRoute(routeName string) *ServiceRoute {
	return s.routes[routeName]
}

func (s *Service) Routes() []ServiceRoute {
	routes := make([]ServiceRoute, len(s.routes))
	index := 0
	for _, route := range s.routes {
		routes[index] = *route
	}
	return routes
}

func (s *Service) Call(name string, ctx *Context, parm interface{}) (interface{}, error) {
	sr, ok := s.routes[name]
	if !ok {
		return nil, fmt.Errorf("route %s is invalid", name)
	}
	c := NewContextFromService(s, sr)
	if ctx != nil {
		if ctx.ctx != nil {
			c.ctx = ctx.ctx
		}
		for k, v := range ctx.Data().data {
			c.Data().Set(k, v)
		}
	}
	res, err := sr.Run(c, s, parm)
	return res, err
}

func (s *Service) CallTo(name string, dest interface{}, ctx *Context, parm interface{}) error {
	anyIsZero := false
	vd := reflect.ValueOf(dest)
	any, e := s.Call(name, ctx, parm)
	if e != nil {
		return e
	}

	vany := reflect.ValueOf(any)
	if vd.Kind() == reflect.Ptr {
		if any == nil {
			anyIsZero = true
		}
	}

	if !anyIsZero {
		if vany.Kind() == reflect.Ptr {
			if vd.Kind() == reflect.Ptr {
				vd.Elem().Set(vany.Elem())
			} else {
				vd.Set(vany.Elem())
			}
		} else {
			if vd.Kind() == reflect.Ptr {
				vd.Elem().Set(vany)
			} else {
				vd.Set(vany)
			}
		}
	}
	return e
}

func (s *Service) RegisterEventHub(ev EventHub, name, secret string) *Service {
	bp := s.BasePoint()
	if bp != "" {
		ev.SetPrefix(bp)
	}
	ev.SetSecret(secret)
	ev.SetService(s)
	s.evHubs[name] = ev
	return s
}

func (s *Service) EventHubs() map[string]EventHub {
	return s.evHubs
}

func (s *Service) EventHub(name string) (EventHub, bool) {
	vev, ok := s.evHubs[name]
	return vev, ok
}

func (s *Service) SetContext(ctx context.Context) *Service {
	s.ctx = ctx
	return s
}

func (s *Service) HubFn() GetHubFn {
	return s.hubFn
}

func (s *Service) SetHubFn(fn GetHubFn) *Service {
	s.hubFn = fn
	return s
}
