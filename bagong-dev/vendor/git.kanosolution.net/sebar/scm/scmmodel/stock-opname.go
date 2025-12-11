package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockOpnameStatus string

const (
	StockOpnameStatusDraft            StockOpnameStatus = "Draft"
	StockOpnameStatusSubmitted        StockOpnameStatus = "Submitted"
	StockOpnameStatusOnProgress       StockOpnameStatus = "OnProgress"
	StockOpnameStatusWaitingForReview StockOpnameStatus = "WaitingForReview"
	StockOpnameStatusDone             StockOpnameStatus = "Done"
)

type StockOpname struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	StockOpnameDate    *time.Time                `form_kind:"date" form_section:"General1"`
	InputDate          time.Time                 `form_kind:"date" form_section:"General1"`
	Responsible        string                    `form_section:"General1"`
	PIC                string                    `form_section:"General1"`
	JournalType        string                    `form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Stock_Opname|_id|_id,Name"  form_section:"General1"`
	Company            string                    `form_section:"General2"`
	Note               string                    `form_multi_row:"5" form_section:"General2"`
	Status             StockOpnameStatus         `form_section:"General2" form_read_only:"1"`
	InventoryDimension InventDimension           `grid:"hide" form_section:"Dimension1"`
	FinancialDimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension2"`
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *StockOpname) TableName() string {
	return "StockOpnames"
}

func (o *StockOpname) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *StockOpname) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *StockOpname) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *StockOpname) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *StockOpname) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *StockOpname) PostSave(dbflex.IConnection) error {
	return nil
}
