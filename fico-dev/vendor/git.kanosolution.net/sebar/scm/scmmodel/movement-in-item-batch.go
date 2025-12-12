package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementInItemBatch struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `grid:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1"`
	MovementInID      string `grid:"hide"`
	ItemID            string `grid:"hide"`
	BatchID           string
	SKU               string `grid:"hide" label:"SKU"`
	Qty               int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *MovementInItemBatch) TableName() string {
	return "MovementInItemBatches"
}

func (o *MovementInItemBatch) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementInItemBatch) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementInItemBatch) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementInItemBatch) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementInItemBatch) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementInItemBatch) PostSave(dbflex.IConnection) error {
	return nil
}
