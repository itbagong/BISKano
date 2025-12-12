package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementOut struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section_size:"4" form_section:"General1" label:"Journal No"`
	TrxDate            *time.Time                `form_section:"General1" form_kind:"date"`
	JournalType        string                    `form_section:"General1" form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Movement_Out|_id|_id,Name" `
	Departement        string                    `form_section:"General1"`
	Notes              string                    `form_section:"General2" form_multi_row:"3"`
	ReasonReject       string                    `form_section:"General2" form_multi_row:"3"`
	Status             MovementStatus            `form_section:"General2" form_read_only:"1"`
	InventoryDimension InventDimension           `form_section:"Dimension1" grid:"hide"`
	FinancialDimension tenantcoremodel.Dimension `form_section:"Dimension2" grid:"hide"`
	RevisionNo         int                       `grid:"hide" form:"hide"` // TODO: increment this when revision
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *MovementOut) TableName() string {
	return "MovementOuts"
}

func (o *MovementOut) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementOut) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementOut) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementOut) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementOut) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementOut) PostSave(dbflex.IConnection) error {
	return nil
}
