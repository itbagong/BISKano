package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemMinMax struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_section_size:"3" form_read_only_edit:"1" form_section:"General1"`
	ItemID             string                    `form_required:"1" form_section:"General1" form_lookup:"/tenant/item/find|_id|Name"`
	SKU                string                    `form_section:"General1" label:"SKU" form:"hide"`
	MinStock           int                       `form_section:"General1"`
	MaxStock           int                       `form_section:"General1"`
	SafeStock          int                       `form_section:"General1"`
	InventoryDimension InventDimension           `form_section:"Dimension1" grid:"hide"`
	FinancialDimension tenantcoremodel.Dimension `form_section:"Dimension2" grid:"hide"`
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General1"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General1"`
}

func (o *ItemMinMax) TableName() string {
	return "ItemMinMaxs"
}

func (o *ItemMinMax) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemMinMax) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemMinMax) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemMinMax) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemMinMax) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemMinMax) PostSave(dbflex.IConnection) error {
	return nil
}
