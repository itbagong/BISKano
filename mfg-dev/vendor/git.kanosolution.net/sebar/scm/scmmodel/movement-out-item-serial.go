package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementOutItemSerial struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1" form:"hide"`
	MovementOutID     string
	ItemID            string
	BatchID           string    `form_section:"General" form_lookup:"/tenant/itembatch/find|_id"`
	SerialNumberID    string    `form_section:"General" form_lookup:"/tenant/itemserial/find|_id"`
	SKU               string    `form_required:"1" form_section:"General" label:"SKU" form:"hide" form_read_only:"1"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info"`
}

func (o *MovementOutItemSerial) TableName() string {
	return "MovementOutItemSerials"
}

func (o *MovementOutItemSerial) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementOutItemSerial) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementOutItemSerial) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementOutItemSerial) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementOutItemSerial) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementOutItemSerial) PostSave(dbflex.IConnection) error {
	return nil
}
