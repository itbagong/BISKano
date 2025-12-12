package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TenantUser struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1"`
	UserID            string    `grid:"hide" form_lookup:"admin/user/find|_id|Name,Email"`
	LoginID           string    `form:"hide"`
	TenantID          string    `grid:"hide" form_lookup:"admin/tenant/find|_id|Name"`
	TenantName        string    `form:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1"`
}

func (o *TenantUser) TableName() string {
	return "TenantUsers"
}

func (o *TenantUser) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TenantUser) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TenantUser) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TenantUser) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TenantUser) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TenantUser) PostSave(dbflex.IConnection) error {
	return nil
}
