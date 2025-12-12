package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderOutputType string

const (
	WorkOrderOutputTypeWOOutput    WorkOrderOutputType = "WO Output"
	WorkOrderOutputTypeWasteItem   WorkOrderOutputType = "Waste Item"
	WorkOrderOutputTypeWasteLedger WorkOrderOutputType = "Waste Ledger"
)

type WorkOrderSummaryOutput struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string              `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"2"`
	WorkOrderPlanID      string              `form_section:"General"` // wajib ada
	Type                 WorkOrderOutputType `form_items:"WO Output|Waste Item|Waste Ledger"`
	InventoryLedgerAccID string              `label:"Inventory / Ledger Acc"` // TODO: type=WO Output|Waste Item -> ngambil dari /tenant/item/find | type=Waste Ledger -> /fico/ledgeraccount/find
	SKU                  string
	Description          string  // default filled by system (editable)
	Group                string  // type=WO Output|Waste Item otomatis di isi dari item yang dipilih (readonly)
	QtyAmount            float64 `label:"Qty / Amount"`
	AchievedQtyAmount    float64 `label:"Achieved Qty / Amount"`
	UnitID               string
	Created              time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *WorkOrderSummaryOutput) TableName() string {
	return "WorkOrderSummaryOutputs"
}

func (o *WorkOrderSummaryOutput) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderSummaryOutput) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderSummaryOutput) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderSummaryOutput) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderSummaryOutput) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderSummaryOutput) PostSave(dbflex.IConnection) error {
	return nil
}
