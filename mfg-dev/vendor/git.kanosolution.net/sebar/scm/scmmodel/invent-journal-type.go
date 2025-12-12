package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventJournalType struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string                     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	Name                   string                     `form_required:"1" form_section:"General"`
	PostingProfileID       string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General"`
	TransactionType        InventTrxType              `form_section:"Contexts" form_items:"Movement In|Movement Out|Item Transfer|Stock Opname|Purchase Quotation|Purchase Order|Inventory Receive|Inventory Issuance"`
	NumberSequenceID       string                     `grid:"hide" form_section:"Contexts"`
	DefaultOffset          ficomodel.SubledgerAccount `grid:"hide" form_section:"General"`
	ChecklistTemplateID    string                     `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name" form_section:"General"`
	ReferenceTemplateID    string                     `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name" form_section:"General"`
	LockInventoryDimension bool                       `grid:"hide" form_section:"General"`
	LockFinancialDimension bool                       `grid:"hide" form_section:"General"`
	LockPostingProfile     bool                       `grid:"hide" form_section:"General"`
	Actions                []JournalTypeContext       `form_section:"Contexts"`
	Previews               []JournalTypeContext       `grid:"hide" form_section:"Contexts"`
	InventoryDimension     InventDimension            `grid:"hide" form_section:"Inventory Dimension"`
	Dimension              tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension"`
	Created                time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General"`
	LastUpdate             time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General"`
}

func (o *InventJournalType) TableName() string {
	return "InventJournalTypes"
}

func (o *InventJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
