package orm

import (
	"git.kanosolution.net/kano/dbflex"
	"github.com/sebarcode/codekit"
)

// DataModel is an interface that should be implemented to a model
type DataModel interface {
	TableName() string
	GetID(dbflex.IConnection) ([]string, []interface{})
	SetID(...interface{})
	SetObjectID(...interface{}) DataModel
	FK() []*FKConfig
	ReverseFK() []*ReverseFKConfig

	GetFilterID(dbflex.IConnection, ...interface{}) *dbflex.Filter
	GetWhereFilter(codekit.M) *dbflex.Filter
	PreSave(dbflex.IConnection) error
	PostSave(dbflex.IConnection) error
	PreDelete(dbflex.IConnection) error
	PostDelete(dbflex.IConnection) error

	Indexes() []dbflex.DbIndex
	Queries() map[string]Query

	SetThis(DataModel)
	This() DataModel
}
