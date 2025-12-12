package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderPlanReportConsumption struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"2"`
	WorkOrderPlanReportID string                    `form_section:"General"`
	WorkOrderPlanID       string                    `form_section:"General"`
	CompanyID             string                    `form_section:"General" form_lookup:"/tenant/company/find|_id|_id,Name" form:"hide"`
	JournalTypeID         string                    `form_section:"General" form_lookup:"/mfg/workorder/journal/type/find|_id|_id,Name"`
	TrxDate               time.Time                 `form_kind:"date" form_section:"General"`
	TrxType               InventTrxType             `form_items:"Work Order" form_section:"General" grid:"hide" form:"hide"`
	Status                ficomodel.JournalStatus   `form:"hide" grid:"hide" form_section:"General" form_read_only:"1"`
	Dimension             tenantcoremodel.Dimension `form:"hide" grid:"hide" form_section:"Dimension"`

	Lines           []WorkOrderMaterialItem `form:"hide" grid:"hide" form_section:"General"`
	AdditionalLines []WorkOrderMaterialItem `form:"hide" grid:"hide" form_section:"General"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`

	Text string // biar ga error aja karena ada penambahan baru
}

type WorkOrderMaterialItem struct {
	Date              time.Time                // di inject by beckend waktu save, ditampilkan hanya di tab Output yang diatas sebelah General
	LineNo            int                      `form:"hide" grid:"hide"` // don't show, for data id only
	ItemID            string                   `label:"Item"`
	SKU               string                   `label:"SKU" grid:"hide"`
	Description       string                   `grid:"hide"`
	UnitID            string                   `label:"UoM"`
	Qty               float64                  // Qty Consumed -> yang nantinya jadi Confirmed
	QtyAvailable      float64                  // filled by backend when /workorderplan/report/get
	UnitCost          float64                  `label:"Unit Price"`
	Total             float64                  // Qty x UnitCost
	Requested         bool                     `grid:"hide"`
	RequestedBy       string                   `label:"Requested By"`
	WarehouseLocation string                   `grid:"hide" form:"hide"`
	IRNo              string                   `grid:"hide"` // TODO: need to be filled when User click "Create Item Request"
	InventDim         scmmodel.InventDimension `grid:"hide" form:"hide" form_section:"Dimension"`
}

func (o *WorkOrderPlanReportConsumption) TableName() string {
	return "WorkOrderPlanReportConsumptions"
}

func (o *WorkOrderPlanReportConsumption) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderPlanReportConsumption) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderPlanReportConsumption) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderPlanReportConsumption) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderPlanReportConsumption) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderPlanReportConsumption) PostSave(dbflex.IConnection) error {
	return nil
}
