package crowd

import (
	"errors"
	"reflect"
)

type DataEngine interface {
	Pre() error
	Run(reflect.Value) error
	Post() error
}

type DataEngineBase struct {
}

func (de *DataEngineBase) Pre() error {
	return nil
}

func (de *DataEngineBase) Post() error {
	return nil
}

func (de *DataEngineBase) Run(v reflect.Value) error {
	return errors.New("Run() is not yet implemented")
}
