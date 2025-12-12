package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementOutDetail struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" form:"hide"`
	MovementOutID      string          `form:"hide"`
	ItemID             string          `form_required:"1" form_section:"General" form_lookup:"/tenant/item/find|_id|Name"`
	SKU                string          `form_required:"1" form_section:"General" label:"SKU" form_lookup:"/tenant/itemspec/find|SKU|SKU"`
	Description        string          `form_multi_row:"2"`
	Qty                int             `form_section:"General"`
	UoM                string          `form_required:"1" form_section:"General" label:"UOM" form_lookup:"/tenant/unit/find|_id|Name"`
	Remarks            string          `form_multi_row:"2"`
	InventoryDimension InventDimension `grid:"hide" form_section:"Specification"`
	Created            time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *MovementOutDetail) TableName() string {
	return "MovementOutDetails"
}

func (o *MovementOutDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementOutDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementOutDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementOutDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementOutDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementOutDetail) PostSave(dbflex.IConnection) error {
	return nil
}
