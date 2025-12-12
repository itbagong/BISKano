package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserPassword struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	Password          string
}

func (o *UserPassword) TableName() string {
	return "RbacUserPasswords"
}

func (o *UserPassword) FK() []*orm.FKConfig {
	return []*orm.FKConfig{
		{FieldID: "ID", RefTableName: new(User).TableName(), RefField: "_id"},
	}
}

func (o *UserPassword) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *UserPassword) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UserPassword) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UserPassword) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *UserPassword) PostSave(dbflex.IConnection) error {
	return nil
}
