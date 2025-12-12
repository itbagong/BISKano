package kaos

import (
	"context"
	"errors"
	"fmt"

	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
)

type Context struct {
	log              *logger.LogEngine
	defaultHubName   string
	defaultEventName string
	path             string

	//hubs map[string]*datahub.Hub
	hubMgr *HubManager
	evs    map[string]EventHub
	data   *SharedData
	ctx    context.Context
	mws    []*MWItem

	hubFn GetHubFn
}

func NewContext(c context.Context,
	logger *logger.LogEngine,
	hubMgr *HubManager,
	evs map[string]EventHub,
	data *SharedData,
	mws ...MWFunc) *Context {
	ctx := new(Context)
	ctx.ctx = c
	ctx.log = logger
	ctx.hubMgr = hubMgr
	ctx.evs = evs
	ctx.data = data

	ctx.mws = make([]*MWItem, len(mws))
	for idx, mw := range mws {
		ctx.mws[idx] = &MWItem{
			Name: fmt.Sprintf("mw-%d", idx),
			Fn:   mw}
	}

	return ctx
}

func NewContextFromService(svc *Service, sr *ServiceRoute) *Context {
	ctx := new(Context)
	ctx.ctx = svc.Context()
	ctx.data = NewSharedData()
	ctx.hubFn = svc.hubFn
	ctx.defaultHubName = "default"

	excludeMws := []string{}
	if sr != nil {
		ctx.path = sr.Path
		ctx.defaultHubName = sr.DefaultHubName
		excludeMws = sr.ExcludeMiddlewares
		for k, v := range sr.Data().data {
			ctx.data.Set(k, v)
		}
	}

	ctx.log = svc.Log()
	ctx.hubMgr = svc.hubMgr

	for _, mw := range svc.Middlewares() {
		if !codekit.HasMember(excludeMws, mw.Name) {
			ctx.mws = append(ctx.mws, mw)
		}
	}

	ctx.evs = svc.evHubs
	ctx.defaultEventName = "default"

	return ctx
}

func (ctx *Context) SetHubFn(fn GetHubFn) *Context {
	ctx.hubFn = fn
	return ctx
}

func (ctx *Context) Context() context.Context {
	return ctx.ctx
}

func (ctx *Context) DefaultEvent() (EventHub, error) {
	if ctx.defaultEventName == "" {
		ctx.defaultEventName = "default"
	}

	h, ok := ctx.evs[ctx.defaultEventName]
	if !ok {
		return nil, errors.New("eventhub key is not valid")
	}
	return h, nil
}

func (ctx *Context) EventHubs() map[string]EventHub {
	if ctx.evs == nil {
		ctx.evs = map[string]EventHub{}
	}
	return ctx.evs
}

func (ctx *Context) DefaultHub() (*datahub.Hub, error) {
	if ctx.hubFn != nil {
		h := ctx.hubFn(ctx)
		return h, nil
	}

	if ctx.defaultHubName == "" {
		ctx.defaultHubName = "default"
	}

	h, err := ctx.hubMgr.Get(ctx.defaultHubName, "")
	if err != nil {
		return nil, fmt.Errorf("errors get hub: %s", err.Error())
	}
	return h, nil
}

func (ctx *Context) GetHub(name, group string) (*datahub.Hub, error) {
	return ctx.hubMgr.Get(name, group)
}

func (ctx *Context) Data() *SharedData {
	return ctx.data
}

func (ct *Context) Log() *logger.LogEngine {
	return ct.log
}
