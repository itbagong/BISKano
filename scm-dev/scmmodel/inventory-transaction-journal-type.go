package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionType string

const (
	TransactionTypeInventoryTransfer   TransactionType = "Inventory_Transfer"
	TransactionTypeMovementIn          TransactionType = "Movement_In"
	TransactionTypeMovementOut         TransactionType = "Movement_Out"
	TransactionTypeStockOpname         TransactionType = "Stock_Opname"
	TransactionTypeInventoryAdjustment TransactionType = "Inventory_Adjustment"
)

type InventoryTransactionJournalType struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name                   string          `form_required:"1" form_section:"General"`
	TransactionType        TransactionType `form_required:"1" form_section:"General" form_items:"Inventory_Transfer|Movement_In|Movement_Out|Stock_Opname|Inventory_Adjustment"`
	NumberSequenceID       string
	DefaultOffset          ficomodel.SubledgerAccount `grid:"hide"`
	PostingProfileID       string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ChecklistTemplateID    string                     `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ReferenceTemplateID    string                     `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	LockInventoryDimension bool                       `grid:"hide"`
	LockFinancialDimension bool                       `grid:"hide"`
	LockPostingProfile     bool                       `grid:"hide"`
	Actions                []JournalTypeContext       `grid:"hide" form_section:"Contexts"`
	Previews               []JournalTypeContext       `grid:"hide" form_section:"Contexts"`
	InventoryDimension     InventDimension            `grid:"hide" form_section:"Inventory Dimension"`
	Dimension              tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension"`
	Created                time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate             time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *InventoryTransactionJournalType) TableName() string {
	return "InventoryTransactionJournalType"
}

func (o *InventoryTransactionJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventoryTransactionJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventoryTransactionJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventoryTransactionJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventoryTransactionJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventoryTransactionJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
