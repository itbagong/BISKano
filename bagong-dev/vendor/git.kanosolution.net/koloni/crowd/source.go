package crowd

import (
	"reflect"
)

type Source interface {
	ElementType() reflect.Type
	Reset()
	Count() int
	Next() (reflect.Value, bool)
	Position() int
}
