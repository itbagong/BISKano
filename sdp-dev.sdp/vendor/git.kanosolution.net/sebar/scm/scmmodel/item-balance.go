package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: Set index for ItemID & all fields inside InventoryDimension

type ItemBalance struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General"`
	ItemID            string `form_section:"General" form_read_only_edit:"1" form_pos:"1,2" label:"Item Name" form_label:"Item" form_lookup:"/tenant/item/find|_id|Name"`
	CompanyID         string `form:"hide"`
	BalanceDate       *time.Time
	SKU               string  `form_section:"General" form_read_only_edit:"1" label:"SKU" form_pos:"1,3" form_label:"SKU"`
	Qty               float64 `form_section:"General" form_read_only_edit:"1" form_pos:"2,2"`
	QtyReserved       float64 `form_section:"General" form_read_only_edit:"1" form_pos:"3,2"`
	QtyPlanned        float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	QtyAvail          float64 `form_section:"General" form_read_only_edit:"1" form_pos:"5,2"`
	AmountPhysical    float64
	AmountFinancial   float64
	AmountAdjustment  float64
	InventDim         InventDimension `grid:"hide" form_section:"General" form_pos:"6"`
	Created           time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type ItemBalanceGrid struct {
	ID               string `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General"`
	ItemID           string `form_section:"General" form_read_only_edit:"1" form_pos:"1,2" label:"Item Name" form_label:"Item"`
	CompanyID        string
	BalanceDate      *time.Time
	SKU              string  `form_section:"General" form_read_only_edit:"1" label:"SKU" form_pos:"1,3" form_label:"SKU"`
	Qty              float64 `form_section:"General" form_read_only_edit:"1" form_pos:"2,2"`
	QtyReserved      float64 `form_section:"General" form_read_only_edit:"1" form_pos:"3,2"`
	QtyPlanned       float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	QtyAvail         float64 `form_section:"General" form_read_only_edit:"1" form_pos:"5,2"`
	AmountPhysical   float64
	AmountFinancial  float64
	AmountAdjustment float64
	InventDim        InventDimension `grid:"hide" form_section:"General" form_pos:"6"`
	Created          time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate       time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

// type ItemBalancePerDimension struct {
// 	WarehouseID   string
// 	WarehouseName string
// 	SectionID     string
// 	SectionName   string
// 	Qty           float64
// 	QtyReserved   float64
// 	QtyPlanned    float64
// 	QtyAvail      float64
// }

type ItemBalanceWithName struct {
	ItemBalance        `grid:"hide" form_section:"General" form_pos:"6"`
	WarehouseID        string  `grid:"hide" form_section:"General" form_pos:"6"`
	AisleID            string  `grid:"hide" form_section:"General" form_pos:"6"`
	SectionID          string  `grid:"hide" form_section:"General" form_pos:"6"`
	BoxID              string  `grid:"hide" form_section:"General" form_pos:"6"`
	BatchID            string  `grid:"hide" form_section:"General" form_pos:"6"`
	SerialNumber       string  `grid:"hide" form_section:"General" form_pos:"6"`
	Size               string  `grid:"hide" form_section:"General" form_pos:"6"`
	Grade              string  `grid:"hide" form_section:"General" form_pos:"6"`
	Spec               string  `grid:"hide" form_section:"General" form_pos:"6"`
	VariantID          string  `grid:"hide" form_section:"General" form_pos:"6"`
	Qty                float64 `form_section:"General" form_read_only_edit:"1" form_pos:"2,2"`
	QtyReserved        float64 `form_section:"General" form_read_only_edit:"1" form_pos:"3,2"`
	QtyPlanned         float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	QtyClosing         float64 `grid:"hide" form_section:"General" form_pos:"6"`
	PlannedQtyClosing  float64 `grid:"hide" form_section:"General" form_pos:"6"`
	ReservedQtyClosing float64 `grid:"hide" form_section:"General" form_pos:"6"`
	AmountPhysical     float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	AmountFinancial    float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	AmountAdjustment   float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
	Amount             float64 `form_section:"General" form_read_only_edit:"1" form_pos:"4,2"`
}

func (o *ItemBalance) TableName() string {
	return "ItemBalances"
}

func (o *ItemBalance) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemBalance) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemBalance) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemBalance) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemBalance) Calc() {
	o.QtyAvail = o.Qty + o.QtyPlanned + o.QtyReserved
}

func (o *ItemBalance) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.Calc()
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemBalance) PostSave(dbflex.IConnection) error {
	return nil
}

type ItemBalanceUniqueFilterParam struct {
	ItemID             string
	SKU                string `label:"SKU"`
	InventoryDimension InventDimension
}

func (o *ItemBalance) UniqueFilter(param ItemBalanceUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}
