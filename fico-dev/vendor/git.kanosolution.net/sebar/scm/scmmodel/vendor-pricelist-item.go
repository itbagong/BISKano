package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VendorPriceListItem struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General"`
	VendorPriceListID string    `grid:"hide" form:"hide"`
	VendorID          string    `form_section:"General" form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	ItemID            string    `form_section:"General2" form_lookup:"/tenant/item/find|_id|_id,Name"`
	SKU               string    `form_section:"General2" label:"SKU" `
	Price             float64   `form_section:"General3"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorPriceListItem) TableName() string {
	return "VendorPriceListItems"
}

func (o *VendorPriceListItem) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorPriceListItem) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorPriceListItem) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorPriceListItem) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorPriceListItem) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *VendorPriceListItem) PostSave(dbflex.IConnection) error {
	return nil
}
