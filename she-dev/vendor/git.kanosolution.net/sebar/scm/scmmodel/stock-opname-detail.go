package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockOpnameDetailRemark string

const (
	StockOpnameDetailRemarkOK    StockOpnameDetailRemark = "OK"    // Gap == 0
	StockOpnameDetailRemarkMinus StockOpnameDetailRemark = "MINUS" // Gap < 0
	StockOpnameDetailRemarkOver  StockOpnameDetailRemark = "OVER"  // Gap > 0
)

type StockOpnameDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	StockOpnameID     string
	ItemID            string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/item/find|_id|Name"`
	SKU               string `label:"SKU"`
	Description       string          `form_multi_row:"2"`
	InventDim         InventDimension `grid:"hide" form_section:"Specification"`
	UnitID            string
	QtyInSystem       float64 `form_label:"Qty In System"`
	QtyActual         float64 `form_label:"Qty Actual"`
	Gap               float64
	Remarks           StockOpnameDetailRemark `form_multi_row:"2"`
	Note              string                  `form_multi_row:"2"`
	NoteStaff         string                  `form_multi_row:"2" form_label:"Staff Note"`
	Created           time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *StockOpnameDetail) TableName() string {
	return "StockOpnameDetails"
}

func (o *StockOpnameDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *StockOpnameDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *StockOpnameDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *StockOpnameDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *StockOpnameDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *StockOpnameDetail) PostSave(dbflex.IConnection) error {
	return nil
}
