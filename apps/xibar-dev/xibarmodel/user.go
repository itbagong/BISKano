package xibarmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type User struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	DisplayName       string `form_required:"1" form_section:"General"`
	Email             string
}

func (o *User) TableName() string {
	return "RbacUsers"
}

func (o *User) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *User) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *User) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *User) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *User) PreSave(dbflex.IConnection) error {
	return nil
}

func (o *User) PostSave(dbflex.IConnection) error {
	return nil
}
