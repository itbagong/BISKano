package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderSummaryMaterial struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"2"`
	WorkOrderPlanID   string `form_section:"General"` // wajib ada
	LineNo            int    // agar bisa sync waktu di daily report consumption
	ItemID            string
	SKU               string
	UnitID            string
	Required          float64                  // Qty yang dicatat saat Plan + yang di tambahkan saat Report
	Reserved          float64                  // Qty yang telah berhasil di reserved
	Used              float64                  // Qty yang telah digunakan saat Report (confirmed)
	AvailableStock    float64                  // get from button GetAvailableStock
	Remarks  string `form:"hide"`
	WarehouseLocation string `grid:"hide" form:"hide"`
	InventDim         scmmodel.InventDimension `grid:"hide" form:"hide" form_section:"Dimension"`
	Created           time.Time                `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate        time.Time                `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *WorkOrderSummaryMaterial) TableName() string {
	return "WorkOrderSummaryMaterials"
}

func (o *WorkOrderSummaryMaterial) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderSummaryMaterial) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderSummaryMaterial) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderSummaryMaterial) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderSummaryMaterial) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderSummaryMaterial) PostSave(dbflex.IConnection) error {
	return nil
}
