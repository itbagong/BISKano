package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderSummaryResource struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"2"`
	WorkOrderPlanID   string    `form_section:"General"` // wajib ada
	ExpenseType       string    `label:"Expense Type"`   // ambil dari Master Expense Type
	TargetHour        float64   `label:"Target Hour"`
	UsedHour          float64   `label:"Used Hour"`
	RatePerHour       float64   `label:"Rate per Hour"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *WorkOrderSummaryResource) TableName() string {
	return "WorkOrderSummaryResources"
}

func (o *WorkOrderSummaryResource) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderSummaryResource) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderSummaryResource) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderSummaryResource) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderSummaryResource) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderSummaryResource) PostSave(dbflex.IConnection) error {
	return nil
}
