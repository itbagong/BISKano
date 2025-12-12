package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemUnitCost struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	ItemID            string
	InventDim         InventDimension
	TrxDate           *time.Time
	UnitCost          float64
}

func (o *ItemUnitCost) TableName() string {
	return "ItemUnitCosts"
}

func (o *ItemUnitCost) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemUnitCost) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemUnitCost) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemUnitCost) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemUnitCost) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *ItemUnitCost) PostSave(dbflex.IConnection) error {
	return nil
}
