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
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_section_size:"3" form_read_only_edit:"1" form_section:"General1" form:"hide"`
	CompanyID          string                    `form_section:"General1" form_lookup:"/tenant/company/find|_id|_id,Name" grid:"hide"  form:"hide"`
	ItemID             string                    `form_required:"1" label:"Item Varian" form_section:"General1" form_lookup:"/tenant/item/find|_id|Name"`
	SKU                string                    `form_section:"General1" label:"SKU" grid:"hide" form:"hide"`
	MinStock           float64                   `form_section:"General1"` // batas minimum qty harus ada di gudang, tidak boleh kurang dari ini
	MaxStock           float64                   `form_section:"General1"` // batas maximum qty harus ada di gudang, tidak boleh lebih dari ini
	SafeStock          float64                   `form_section:"General1"` // nantinya akan dipakai di report
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	FinancialDimension tenantcoremodel.Dimension `form_section:"Dimension"`
	InventoryDimension InventDimension           `form_section:"Dimension" grid:"hide"`
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
