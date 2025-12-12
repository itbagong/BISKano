package sebar

import (
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/samber/lo"
)

func ToDataModels[T orm.DataModel](records []T) []orm.DataModel {
	return lo.Map(records, func(t T, index int) orm.DataModel {
		return t
	})
}

func FromDataModels[T orm.DataModel](models []orm.DataModel, model T) []T {
	return lo.Map(models, func(t orm.DataModel, index int) T {
		return t.(T)
	})
}
