package kaos

import (
	"sync"
)

// SharedData is shared data
type SharedData struct {
	sync.RWMutex
	data map[string]interface{}
}

// NewSharedData initiate new SharedData
func NewSharedData() *SharedData {
	d := &SharedData{}
	d.data = map[string]interface{}{}
	return d
}

// Keys get all keys
func (s *SharedData) Keys() []string {
	defer s.RUnlock()
	s.RLock()
	keys := make([]string, len(s.data))
	idx := 0
	for k := range s.data {
		keys[idx] = k
		idx++
	}
	return keys
}

// Get data with given key and default value if not exist
func (s *SharedData) Get(key string, def interface{}) interface{} {
	var out interface{}
	var b bool
	hasData := false

	s.RLock()
	if s.data != nil {
		hasData = true
		out, b = s.data[key]
	}
	s.RUnlock()

	if hasData {
		if b {
			return out
		}

		return def
	}

	return def
}

// Set data with given key and value
func (s *SharedData) Set(key string, value interface{}) {
	s.Lock()
	if s.data == nil {
		s.data = map[string]interface{}{}
	}
	s.data[key] = value
	s.Unlock()
}

// Remove data with given key
func (s *SharedData) Remove(key string) {
	s.Lock()
	delete(s.data, key)
	s.Unlock()
}

// Data get copy of all data
func (s *SharedData) Data() map[string]interface{} {
	var res map[string]interface{}

	s.RLock()
	res = s.data
	s.RUnlock()

	return res
}

// Count data
func (s *SharedData) Count() int {
	out := 0
	s.RLock()
	if s.data != nil {
		out = len(s.data)
	}
	s.RUnlock()
	return out
}
