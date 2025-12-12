package dbflex

import (
	"errors"
	"sync"
)

type SliceMtx[T any] struct {
	sync.RWMutex
	slice []T

	comparator func(a, b T) bool
}

func (s *SliceMtx[T]) SetComparator(f func(a, b T) bool) {
	s.comparator = f
}

func (s *SliceMtx[T]) Add(item T) {
	s.Lock()
	defer s.Unlock()

	s.slice = append(s.slice, item)
}

func (s *SliceMtx[T]) Remove(index int) {
	s.Lock()
	defer s.Unlock()

	if index < 0 || index >= len(s.slice) {
		return
	}

	s.slice = append(s.slice[:index], s.slice[index+1:]...)
}

func (s *SliceMtx[T]) Get(index int) (res T, err error) {
	s.RLock()
	defer s.RUnlock()

	if index < 0 || index >= len(s.slice) {
		err = errors.New("index out of range")
		return
	}

	res = s.slice[index]
	return
}

func (s *SliceMtx[T]) Count() int {
	s.RLock()
	defer s.RUnlock()

	return len(s.slice)
}

func (s *SliceMtx[T]) RemoveByValue(v T) {
	s.Lock()
	defer s.Unlock()

	if s.comparator == nil {
		return
	}

	buffers := []T{}
	for _, val := range s.slice {
		if !s.comparator(val, v) {
			buffers = append(buffers, val)
		}
	}
	s.slice = buffers
}
