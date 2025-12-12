package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/suim"
)

type User struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	LoginID           string `form_required:"1"`
	DisplayName       string `form_required:"1"`
	Email             string
	Enable            bool
	Status            string `form_required:"1"`
	WalletAddress     string
	Use2FA            bool      `label:"Use 2FA"`
	Created           time.Time `grid:"hide" form_pos:"5001,1" form_read_only:"1"`
	LastUpdated       time.Time `grid:"hide" form_pos:"5001,2" form_read_only:"1"`
}

func (o *User) TableName() string {
	return "RbacUsers"
}

func (o *User) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *User) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{
		{FieldID: "ID", RefTableName: new(UserPassword).TableName(), RefField: "_id", AutoDelete: true},
		{FieldID: "ID", RefTableName: new(RoleMember).TableName(), RefField: "UserID", AutoDelete: true},
	}
}

func (o *User) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *User) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *User) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = sebar.MakeID("", sebar.PrecisionMilli, 32)
	}
	if e := suim.Validate(o); e != nil {
		return e
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdated = time.Now()
	return nil
}

func (o *User) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *User) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "loginid_index", IsUnique: true, Fields: []string{"LoginID"}},
		{Name: "email_index", IsUnique: false, Fields: []string{"Email"}},
		{Name: "wallet_index", IsUnique: false, Fields: []string{"WalletAddress"}},
	}
}
