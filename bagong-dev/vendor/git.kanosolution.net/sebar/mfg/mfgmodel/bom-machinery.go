package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Machinery struct {
	MachineCode  string
	MachineName  string
	StandartHour int
	RatePerHour  float64
}

type BoMMachinery struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section:"General" form:"hide"`
	BoMID             string
	Machinery
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *BoMMachinery) TableName() string {
	return "BoMMachineries"
}

func (o *BoMMachinery) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BoMMachinery) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BoMMachinery) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BoMMachinery) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BoMMachinery) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BoMMachinery) PostSave(dbflex.IConnection) error {
	return nil
}
