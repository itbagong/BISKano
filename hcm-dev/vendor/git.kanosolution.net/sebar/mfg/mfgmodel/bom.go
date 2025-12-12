package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BomOutputType string

const (
	BomOutputTypeItem   BomOutputType = "Item"
	BomOutputTypeLedger BomOutputType = "Ledger"
)

type BoM struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string        `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	Title             string        `form_section:"General1"`
	BoMGroup          string        `form_section:"General2"  label:"BOM Group" form_items:"Production|Service"`
	Description       string        `form_section:"General2" form_multi_row:"2"`
	OutputType        BomOutputType `form_section:"General3" form_items:"Item|Ledger"`
	ItemID            string        `form_section:"General4"`
	SKU               string        `form_section:"General4" label:"SKU" form:"hide"`
	LedgerID          string        `form_section:"General4"`
	IsActive          bool          `form_section:"General3"`
	Created           time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate        time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *BoM) TableName() string {
	return "BoMs"
}

func (o *BoM) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BoM) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BoM) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BoM) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BoM) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BoM) PostSave(dbflex.IConnection) error {
	return nil
}
