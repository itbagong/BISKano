package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"github.com/google/uuid"
)

type Token struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	App               string `form_required:"1"`
	Kind              string `form_required:"1"`
	Token             string `form_required:"1"`
	Status            string `form_required:"1" form_user_list:"1" items:"Created|Reserved|Claimed|Expired"` //Options: Created, Reserved, Claimed, Expired
	Created           time.Time
	Expiry            time.Time
	ClaimDate         time.Time
	UserID            string
}

func (o *Token) TableName() string {
	return "RbacTokens"
}

func (o *Token) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *Token) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *Token) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Token) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Token) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == "" {
		o.Status = "Created"
	}
	if e := suim.Validate(o); e != nil {
		return e
	}
	return nil
}

func (o *Token) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Token) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "App_Kind_Token_Index", IsUnique: true, Fields: []string{"App", "Kind", "Token", "Status"}},
	}
}
