package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderPlanReportResource struct {
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

	Lines []WorkOrderResourceItem `form:"hide" grid:"hide" form_section:"General"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`

	Text string // biar ga error aja karena ada penambahan baru
}

type WorkOrderResourceItem struct {
	Date         time.Time // di inject by beckend waktu save, ditampilkan hanya di tab Output yang diatas sebelah General
	ExpenseType  string    // ngambil dari Master Expense Type
	ActivityName string    `label:"Activity Name"`
	Employee     string
	WorkingHour  float64
	RatePerHour  float64
	Total        float64 // TODO: isi ini dengan WorkingHour x RatePerHour dari Summary Resource
}

func (o *WorkOrderPlanReportResource) TableName() string {
	return "WorkOrderPlanReportResources"
}

func (o *WorkOrderPlanReportResource) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderPlanReportResource) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderPlanReportResource) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderPlanReportResource) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderPlanReportResource) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderPlanReportResource) PostSave(dbflex.IConnection) error {
	return nil
}
