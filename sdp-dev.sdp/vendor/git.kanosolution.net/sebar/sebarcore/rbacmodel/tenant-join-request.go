package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TenantJoinEnum string

const (
	TenantJoinInvitation TenantJoinEnum = "INVITATION"
	TenantJoinRequest    TenantJoinEnum = "REQUEST"
)

type TenantJoinExtended struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string
	DisplayName       string
	TenantID          string
	TenantName        string
	Status            string
	Created           time.Time
}

func (o *TenantJoinExtended) TableName() string {
	return "TenantJoinRequests"
}

type TenantJoin struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string
	TenantID          string
	JoinType          TenantJoinEnum `form_items:"INVITATION|REQUEST"`
	Status            string         `form_items:"OPEN|APPROVED|REJECTED"`
	Created           time.Time      `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time      `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TenantJoin) TableName() string {
	return "TenantJoinRequests"
}

func (o *TenantJoin) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TenantJoin) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TenantJoin) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TenantJoin) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TenantJoin) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TenantJoin) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *TenantJoin) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "JoinTypeRequestIndex", Fields: []string{"JoinType", "TenantID", "UserID"}},
		{Name: "UserIndex", Fields: []string{"JoinType", "UserID"}},
	}
}
