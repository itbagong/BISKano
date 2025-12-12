package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventoryAdjustmentDetail struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	InventoryAdjustmentID string
	ItemID                string          `form_required:"1" form_section:"Specification" form_lookup:"/tenant/item/find|_id|Name" width:"200px"`
	SKU                   string          `label:"SKU"`
	Description           string          `form_multi_row:"2"`
	InventoryDimension    InventDimension `grid:"hide" form_section:"Specification"`
	UoM                   string          `label:"UOM"`
	ItemName              string          `form_label:"Item"`
	UnitName              string          `form_label:"Unit"`
	AisleName             string          `form_label:"Aisle"`
	SectionName           string          `form_label:"Section"`
	BoxName               string          `form_label:"Box"`
	QtyInSystem           float64         `form_label:"Qty In System"`
	QtyActual             float64         `form_label:"Qty Actual"`
	Gap                   float64
	LineNo                int                     `form:"hide" grid:"hide"`
	Remarks               StockOpnameDetailRemark `form_multi_row:"2"`
	Note                  string                  `form_multi_row:"2" form:"hide"`
	// NoteStaff             string                  `form_multi_row:"2" form_label:"Staff Note"`
	NoteAdjustment string `form_multi_row:"2" form_label:"Note Adjustment" grid:"hide" form:"hide"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *InventoryAdjustmentDetail) TableName() string {
	return "InventoryAdjustmentDetails"
}

func (o *InventoryAdjustmentDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventoryAdjustmentDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventoryAdjustmentDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventoryAdjustmentDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventoryAdjustmentDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventoryAdjustmentDetail) PostSave(dbflex.IConnection) error {
	return nil
}
