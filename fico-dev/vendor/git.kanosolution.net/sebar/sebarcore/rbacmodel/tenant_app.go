package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TenantApp struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_onlyt:"1" form:"hide" grid:"hide" form_section:"General" form_section_auto_col:"2"`
	TenantID          string `form_lookup:"/admin/tenant/find|_id|_id,Name"`
	AppID             string `form_lookup:"/admin/app/find|_id|_id,Name"`
	Enable            bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TenantApp) TableName() string {
	return "TenantApps"
}

func (o *TenantApp) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TenantApp) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TenantApp) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TenantApp) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TenantApp) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TenantApp) PostSave(dbflex.IConnection) error {
	return nil
}
