package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementOutItemBatch struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1" form:"hide"`
	MovementOutID     string
	ItemID            string
	BatchID           string `form_section:"General" form_lookup:"/tenant/itembatch/find|_id"`
	SKU               string `form_required:"1" form_section:"General" label:"SKU" form_lookup:"/tenant/itemspec/find|SKU|SKU" form_read_only:"1"`
	Qty               int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *MovementOutItemBatch) TableName() string {
	return "MovementOutItemBatches"
}

func (o *MovementOutItemBatch) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementOutItemBatch) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementOutItemBatch) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementOutItemBatch) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementOutItemBatch) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementOutItemBatch) PostSave(dbflex.IConnection) error {
	return nil
}
