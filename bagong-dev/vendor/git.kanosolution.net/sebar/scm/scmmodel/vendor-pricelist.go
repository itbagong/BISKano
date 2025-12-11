package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VendorPriceList struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"3" form_section:"General"`
	VendorID          string    `form_section:"General" form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	ItemID            string    `form_section:"General2" form_lookup:"/tenant/item/find|_id|_id,Name"`
	SKU               string    `form_section:"General2" label:"SKU" form:"hide"`
	Price             float64   `form_section:"General2"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorPriceList) TableName() string {
	return "VendorPriceLists"
}

func (o *VendorPriceList) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorPriceList) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorPriceList) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorPriceList) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorPriceList) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *VendorPriceList) PostSave(dbflex.IConnection) error {
	return nil
}
