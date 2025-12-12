package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BoMMaterial struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section:"General" form:"hide"`
	BoMID             string
	ItemID            string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/item/find|_id|Name"`
	SKU               string `label:"SKU"`
	Description       string `form_multi_row:"2"`
	Qty               int
	UoM               string `label:"UOM"`
	UnitPrice         float64
	Total             float64
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *BoMMaterial) TableName() string {
	return "BoMMaterials"
}

func (o *BoMMaterial) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BoMMaterial) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BoMMaterial) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BoMMaterial) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BoMMaterial) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BoMMaterial) PostSave(dbflex.IConnection) error {
	return nil
}
