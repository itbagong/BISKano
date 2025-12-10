package afycoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReferenceTemplateItem struct {
	ID            string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	ReferenceType string `form_items:"date|number|text|lookup|items"`
	Label         string `form_required:"1" form_section:"General"`
	ConfigValue   string
}

// lookup
// contoh : /tenant/employeegroup/find|_id|Name
// items
// contoh : CASH|XXX|AAA

type ReferenceTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string
	Items             []ReferenceTemplateItem `grid:"hide"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ReferenceTemplate) TableName() string {
	return "ReferenceTemplates"
}

func (o *ReferenceTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ReferenceTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ReferenceTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ReferenceTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ReferenceTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ReferenceTemplate) PostSave(dbflex.IConnection) error {
	return nil
}
