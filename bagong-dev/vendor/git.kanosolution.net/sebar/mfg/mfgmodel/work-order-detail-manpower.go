package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderDetailManpower struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section:"General" form:"hide"`
	ActivityName      string
	EmployeeQty       int
	StandardHour      float64
	RatePerHour       float64
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *WorkOrderDetailManpower) TableName() string {
	return "WorkOrderDetailManpowers"
}

func (o *WorkOrderDetailManpower) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderDetailManpower) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderDetailManpower) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderDetailManpower) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderDetailManpower) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderDetailManpower) PostSave(dbflex.IConnection) error {
	return nil
}
