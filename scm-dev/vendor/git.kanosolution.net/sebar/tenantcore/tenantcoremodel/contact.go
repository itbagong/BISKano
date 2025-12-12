package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contact struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" form:"hide" grid:"hide"`
	CustomerID        string `form:"hide" grid:"hide"`
	Name              string `form_required:"1" form_section:"General"`
	Company           string `form_required:"1" form_lookup:"/tenant/company/find|_id|_id,Name"`
	Site              string `form_required:"1" form_lookup:"/bagong/sitesetup/find|_id|Name"`
	Role              string `form_required:"1" form_section:"General"`
	BusinessPhoneNo   string `form_required:"1" form_section:"General"`
	PhoneNumber       string `form_required:"1" form_section:"General"`
	Email             string `form_required:"1" form_section:"General"`
	AsContactPerson	  bool 	 `form_kind:"checkbox" form_section:"General"`
	// IsActive          bool
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Contact) TableName() string {
	return "Contact"
}

func (o *Contact) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Contact) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Contact) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Contact) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Contact) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Contact) PostSave(dbflex.IConnection) error {
	return nil
}
