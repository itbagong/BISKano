package sebar

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
)

// Validate struct
func Validate(data interface{}) error {
	if data == nil {
		return fmt.Errorf("params cannot be null or empty")
	}

	if e := validate.Struct(data); e != nil {
		return e
	}

	return nil
}
