package datahub

import (
	"errors"
	"reflect"

	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/kiva"
)

var (
	fcss = map[string]*kiva.CacheOptions{}
)

func SetObjectCache(name string, obj interface{}) error {
	if obj == nil {
		return errors.New("object is nil")
	}
	rv := reflect.ValueOf(obj)
	rt := rv.Type()
	mtdNum := rt.NumMethod()
	for mi := 0; mi < mtdNum; mi++ {
		mtd := rt.Method(mi)
		if mtd.Name == "CacheSetup" && mtd.Type.NumOut() == 1 && mtd.Type.Out(0).String() == "*kiva.CacheOptions" {
			vFcs := rv.Method(mi).Interface()
			fcs, ok := vFcs.(*kiva.CacheOptions)
			if !ok {
				return errors.New("return of CacheSetup should be *datahub.FieldCaheSetup")
			}
			fcss[name] = fcs
			return nil
		}
	}

	fcss[name] = nil
	return nil
}

func getObjectCacheSetup(name string, obj interface{}) *kiva.CacheOptions {
	fcs, ok := fcss[name]
	if ok {
		return fcs
	}

	e := SetObjectCache(name, obj)
	if e != nil {
		return nil
	}
	return getObjectCacheSetup(name, obj)
}

type InMemOpts struct {
	Provider kiva.MemoryProvider
}

func (h *Hub) Cache(opts *InMemOpts) *Hub {
	if opts == nil {
		h.useCache = false
		h.cacheProvider = nil
		return h
	}

	h.useCache = true
	h.cacheProvider = opts.Provider
	return h
}

func getID(dt orm.DataModel) string {
	_, ids := dt.GetID(nil)
	if len(ids) == 0 {
		return ""
	}
	return ids[0].(string)
}
