package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseRequestJournalType struct {
	orm.DataModelBase       `bson:"-" json:"-"`
	ID                      string        `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General"  form_section_size:"4"`
	Name                    string        `form_required:"1" form_section:"General"`
	TrxType                 InventTrxType `form_read_only:"1" form_items:"Purchase Order" form_section:"General"`
	PostingProfileID        string        `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	ReceivePostingProfileID string        `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	InvoicePostingProfileID string        `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General3"`
	LockPostingProfileID    string        `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General3"`
	Created                 time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General4"`
	LastUpdate              time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General4"`
}

func (o *PurchaseRequestJournalType) TableName() string {
	return "PurchaseRequestJournalTypes"
}

func (o *PurchaseRequestJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PurchaseRequestJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PurchaseRequestJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PurchaseRequestJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PurchaseRequestJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PurchaseRequestJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
