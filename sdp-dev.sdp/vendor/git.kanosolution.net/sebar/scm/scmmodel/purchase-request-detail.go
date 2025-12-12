package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseRequestDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General"`
	PurchaseRequestID string
	LineNo            int
	ItemID            string
	SKU               string `label:"SKU"`
	InventDim         InventDimension
	Qty               float64
	UoM               string `label:"UOM"`
	UnitPrice         float64
	Tax               float64
	Discount          float64
	SubTotal          float64
	Remarks           string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PurchaseRequestDetail) TableName() string {
	return "PurchaseRequestDetails"
}

func (o *PurchaseRequestDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PurchaseRequestDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PurchaseRequestDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PurchaseRequestDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PurchaseRequestDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PurchaseRequestDetail) PostSave(dbflex.IConnection) error {
	return nil
}

type PurchaseRequestDetailUniqueFilterParam struct {
	ItemID             string
	SKU                string
	InventoryDimension InventDimension
}

func (o *PurchaseRequestDetail) UniqueFilter(param PurchaseRequestDetailUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}
