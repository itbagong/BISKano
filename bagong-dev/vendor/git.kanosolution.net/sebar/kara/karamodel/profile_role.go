package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TenantRole string

const (
	TenantManager TenantRole = "LocationManager"
	LocationAdmin TenantRole = "LocationAdmin"
	GlobalAdmin   TenantRole = "GlobalAdmin"
)

type ProfileRole struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string
	RoleID            string
	WorkLocationID    string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ProfileRole) TableName() string {
	return "KaraProfileRoles"
}

func (o *ProfileRole) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ProfileRole) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ProfileRole) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ProfileRole) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ProfileRole) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ProfileRole) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *ProfileRole) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "User_Role_WorkLocation_Index", Fields: []string{"RoleID", "UserID", "WorkLocationID"}},
	}
}
