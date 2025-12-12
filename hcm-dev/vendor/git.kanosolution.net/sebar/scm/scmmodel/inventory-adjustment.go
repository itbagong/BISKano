package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventoryAdjustmentStatus string

const (
	InventoryAdjustmentStatusNeedToReview InventoryAdjustmentStatus = "NeedToReview"
	InventoryAdjustmentStatusDone         InventoryAdjustmentStatus = "Done"
)

type InventoryAdjustment struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"3" form_section:"General1"`
	StockOpnameID     string    `form_section:"General1" form_label:"Stock Opname No."`
	AdjustmentDate    time.Time `form_kind:"date" form_section:"General1" form_label:"Adjustment Date"`
	// JournalType       string                    `form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Inventory_Adjustment|_id|_id,Name" form_section:"General1"`
	CompanyID  string                    `form_section:"General1" grid:"hide" form:"hide"`
	Note       string                    `form_multi_row:"5" form_section:"General2" grid:"hide" form:"hide" `
	Status     InventoryAdjustmentStatus `form_section:"General2" grid:"hide" form:"hide" form_read_only:"1"`
	Dimension  tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension1"`
	InventDim  InventDimension           `grid:"hide" form_section:"Dimension1"`
	Created    time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *InventoryAdjustment) TableName() string {
	return "InventoryAdjustments"
}

func (o *InventoryAdjustment) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventoryAdjustment) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventoryAdjustment) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventoryAdjustment) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventoryAdjustment) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventoryAdjustment) PostSave(dbflex.IConnection) error {
	return nil
}
