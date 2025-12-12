package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemRequestFulfillmentType string

const (
	ItemRequestFulfillmentTypeItemTransfer    ItemRequestFulfillmentType = "Item Transfer"
	ItemRequestFulfillmentTypeMovementOut     ItemRequestFulfillmentType = "Movement Out"
	ItemRequestFulfillmentTypePurchaseRequest ItemRequestFulfillmentType = "Purchase Request"
)

type ItemRequestDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General" form_section_show_title:"1" form_read_only_edit:"1" form:"hide" form_section_auto_col:"2"`
	ItemRequestID     string                    `form_section:"General" form_read_only_edit:"1" form:"hide"`
	ItemID            string                    `form_section:"General" form_read_only_edit:"1" label:"Item"`
	SKU               string                    `form_section:"General" form_read_only_edit:"1" label:"SKU"`
	Description       string                    `form_section:"General" form_read_only_edit:"1" grid:"hide" form:"hide"`
	ItemType          string                    `form_section:"General" form_read_only_edit:"1" grid:"hide" form:"hide"` // from Master Item
	UoM               string                    `form_section:"General" label:"UoM" form_read_only_edit:"1" form:"hide"`
	QtyRequested      float64                   `form_section:"General" form_read_only_edit:"1"`
	QtyFulfilled      float64                   `form_section:"General" form_read_only_edit:"1"` // Total penjumlahan QtyFulfilled dari semua DetailLines
	QtyAvailable      float64                   `form_section:"General" form_read_only_edit:"1"`
	QtyRemaining      float64                   `form_section:"General" form_read_only_edit:"1"` // for UI only
	Complete          bool                      `form_section:"General" form_read_only_edit:"1"`
	Remarks           string                    `form_section:"General" form_read_only_edit:"1"`
	WarehouseID       string                    `form_section:"General" form_read_only_edit:"1" form_lookup:"/tenant/warehouse/find|_id|_id,Name" form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form:"hide"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form:"hide"`
	Dimension         tenantcoremodel.Dimension `form:"hide"`
	DetailLines       []ItemRequestDetailLine   `form_section:"Fulfillment Method" form_section_show_title:"1"`
}

type ItemRequestDetailLine struct {
	FulfillmentType ItemRequestFulfillmentType `form_section:"General" form_read_only_edit:"1" form_section_auto_col:"2"`
	QtyFulfilled    float64                    `form_section:"General" form_read_only_edit:"1"`
	UoM             string                     `form_section:"General" form_read_only_edit:"1" label:"UoM"`
	WarehouseID     string                     `form_section:"General" label:"From warehouse"`
	InventDimFrom   InventDimension            `form:"hide" grid:"hide"` // digunakan utk FulfillmentType = Item Transfer
	QtyAvailable    float64                    `form:"hide" grid:"hide" form_section:"General" form_read_only_edit:"1"`
}

func (o *ItemRequestDetail) TableName() string {
	return "ItemRequestDetails"
}

func (o *ItemRequestDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemRequestDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemRequestDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemRequestDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemRequestDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemRequestDetail) PostSave(dbflex.IConnection) error {
	return nil
}

type ItemRequestDetailUniqueFilterParam struct {
	ItemID             string
	SKU                string `label:"SKU"`
	InventoryDimension InventDimension
}

func (o *ItemRequestDetail) UniqueFilter(param ItemRequestDetailUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}
