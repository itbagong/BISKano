package kaos

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ariefdarmawan/datahub"
)

type HubManager struct {
	lock       *sync.RWMutex
	hubs       map[string]*datahub.Hub
	hubBuilder func(key, group string) (*datahub.Hub, error)
}

func NewHubManager(builder func(key, group string) (*datahub.Hub, error)) *HubManager {
	m := new(HubManager)
	m.hubs = make(map[string]*datahub.Hub)
	m.hubBuilder = builder
	m.lock = new(sync.RWMutex)
	return m
}

func (hm *HubManager) SetHubBuilder(builder func(key, group string) (*datahub.Hub, error)) {
	hm.hubBuilder = builder
}

func (hm *HubManager) Keys() []string {
	keys := make([]string, len(hm.hubs))
	index := 0
	for k := range hm.hubs {
		keys[index] = k
		index++
	}
	return keys
}

func (hm *HubManager) Set(k string, group string, h *datahub.Hub) {
	hm.lock.Lock()
	defer hm.lock.Unlock()
	if group != "" {
		k = fmt.Sprintf("%s_%s", group, k)
	}
	hm.hubs[k] = h
}

func (hm *HubManager) Get(k string, group string) (*datahub.Hub, error) {
	var err error
	if group != "" {
		k = fmt.Sprintf("%s_%s", group, k)
	}

	hm.lock.RLock()
	h, has := hm.hubs[k]
	hm.lock.RUnlock()

	if !has {
		if hm.hubBuilder == nil {
			return nil, errors.New("missing: HubBuilder")
		}
		h, err = hm.hubBuilder(k, group)
		if err != nil {
			return nil, err
		}
		hm.Set(k, group, h)
	}
	return h, nil
}

func (hm *HubManager) GetMust(k, g string) *datahub.Hub {
	h, _ := hm.Get(k, g)
	return h
}

func (hm *HubManager) Remove(k, group string) {
	hm.lock.Lock()
	defer hm.lock.Unlock()

	if group != "" {
		k = fmt.Sprintf("%s_%s", group, k)
	}
	h, has := hm.hubs[k]
	if has {
		h.Close()
		delete(hm.hubs, k)
	}
}

func (hm *HubManager) Close() {
	for _, h := range hm.hubs {
		h.Close()
	}
	hm.hubs = make(map[string]*datahub.Hub)
}
