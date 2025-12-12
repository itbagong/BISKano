package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TemporaryIRPRAuditTrail struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string
	IRID                   string
	ReferenceID            interface{}
	Operation              string
	Error                  string
	IsSuccess              bool
	IRCreatedDate          time.Time
	FullfilmentCreatedDate time.Time
	Created                time.Time
	LastUpdate             time.Time
}

func (o *TemporaryIRPRAuditTrail) TableName() string {
	return "TemporaryIRPRAuditTrails"
}

func (o *TemporaryIRPRAuditTrail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TemporaryIRPRAuditTrail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TemporaryIRPRAuditTrail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TemporaryIRPRAuditTrail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TemporaryIRPRAuditTrail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TemporaryIRPRAuditTrail) PostSave(dbflex.IConnection) error {
	return nil
}
