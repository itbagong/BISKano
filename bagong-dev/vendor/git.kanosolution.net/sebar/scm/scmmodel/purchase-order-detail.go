package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseOrderDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General"`
	PurchaseOrderID   string `form_section:"General"`
	LineNo            int
	ItemID            string `form_section:"General"`
	SKU               string `form_section:"General" label:"SKU"`
	InventDim         InventDimension
	Description       string    `form_section:"General"`
	Qty               float64   `form_section:"General"`
	UoM               string    `form_section:"General" label:"UOM"`
	UnitPrice         float64   `form_section:"General"`
	Tax               float64   `form_section:"General"`
	Discount          float64   `form_section:"General"`
	SubTotal          float64   `form_section:"General"`
	Remarks           string    `form_section:"General"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PurchaseOrderDetail) TableName() string {
	return "PurchaseOrderDetails"
}

func (o *PurchaseOrderDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PurchaseOrderDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PurchaseOrderDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PurchaseOrderDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PurchaseOrderDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PurchaseOrderDetail) PostSave(dbflex.IConnection) error {
	return nil
}

type PurchaseOrderDetailUniqueFilterParam struct {
	ItemID             string
	SKU                string `label:"SKU"`
	InventoryDimension InventDimension
}

func (o *PurchaseOrderDetail) UniqueFilter(param PurchaseOrderDetailUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}
