package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	Name              string
	OwnByTenant       bool
	TenantID          string `grid:"hide" form_lookup:"admin/tenant/find|_id|Name" form_label_field:"TenantName"`
	TenantName        string `form:"hide"`
	Protected         bool
	Enable            bool
}

func (o *Role) TableName() string {
	return "RbacRoles"
}

func (o *Role) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Role) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
	//return []*orm.ReverseFKConfig{{FieldID: "_id", RefTableName: new(RoleMember).TableName(), RefField: "RoleID", AutoDelete: true}}
}

func (o *Role) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Role) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Role) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *Role) PostSave(dbflex.IConnection) error {
	return nil
}
