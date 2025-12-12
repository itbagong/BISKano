package crowd

import (
	"reflect"
	"strings"
	"time"
)

type StatCalcInt struct {
	DataEngineBase
	Min, Max, Count, Sum int
	Avg                  float64
}

type StatCalcFloat struct {
	DataEngineBase
	Min, Max, Count, Sum float64
	Avg                  float64
}

type StatCalcString struct {
	Min, Max string
	Count    int
}

type StatCalcDate struct {
	Count    int
	Min, Max time.Time
}

func (s *StatCalcInt) Run(val reflect.Value) error {
	v := int(val.Int())

	s.Count++
	s.Sum += v
	if s.Count == 1 {
		s.Min = v
		s.Max = v
	} else {
		if s.Min > v {
			s.Min = v
		}

		if s.Max < v {
			s.Max = v
		}
	}
	return nil
}

func (s *StatCalcInt) Post() error {
	s.Avg = float64(s.Sum) / float64(s.Count)
	return nil
}

func (s *StatCalcFloat) Run(val reflect.Value) error {
	v := val.Float()

	s.Count++
	s.Sum += v
	if s.Count == 1 {
		s.Min = v
		s.Max = v
	} else {
		if s.Min > v {
			s.Min = v
		}

		if s.Max < v {
			s.Max = v
		}
	}
	return nil
}

func (s *StatCalcFloat) Post() error {
	s.Avg = s.Sum / float64(s.Count)
	return nil
}

func (s *StatCalcString) Run(val reflect.Value) error {
	v := val.String()

	s.Count++
	if s.Count == 1 {
		s.Min = v
		s.Max = v
	} else {
		if strings.Compare(s.Min, v) > 0 {
			s.Min = v
		}

		if strings.Compare(s.Max, v) < 0 {
			s.Max = v
		}
	}
	return nil
}
