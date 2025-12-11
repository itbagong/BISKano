package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PhysicalAvailability struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1"  form_section_size:"4" form_section:"General"` // 1-to-1 with RoutineChecklist
	Unit              string `form_section:"General" form_lookup:"/tenant/asset/find|_id|_id,Name"`
	PATarget          float64 `form_section:"General2" label:"PA Target"`
	PAActual          float64 `form_section:"General2" label:"PA Actual" form:"hide"`
	Month             int `form_section:"General3"`
	Year              int `form_section:"General3"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PhysicalAvailability) TableName() string {
	return "PhysicalAvailabilities"
}

func (o *PhysicalAvailability) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PhysicalAvailability) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PhysicalAvailability) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PhysicalAvailability) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PhysicalAvailability) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PhysicalAvailability) PostSave(dbflex.IConnection) error {
	return nil
}
