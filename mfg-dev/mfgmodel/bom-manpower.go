package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ManPower struct {
	ExpenseType      string `form_lookup:"/tenant/expensetype/find|_id|_id,Name"`
	ActivityName     string `form_multi_row:"3"`
	EmployeeQuantity int `label:"Employee Qty"`
	StandartHour     int `label:"Standart Hour"`
	RatePerHour      float64
	Description      string `form_multi_row:"3"`
}

type BoMManpower struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section:"General" form:"hide"`
	BoMID             string
	ManPower
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *BoMManpower) TableName() string {
	return "BoMManpowers"
}

func (o *BoMManpower) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BoMManpower) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BoMManpower) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BoMManpower) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BoMManpower) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BoMManpower) PostSave(dbflex.IConnection) error {
	return nil
}
