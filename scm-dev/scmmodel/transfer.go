package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransferStatus string

const (
	TransferStatusDraft           TransferStatus = "Draft"
	TransferStatusSubmitted       TransferStatus = "Submitted"
	TransferStatusApproved        TransferStatus = "Approved"
	TransferStatusPartialReceived TransferStatus = "PartialReceived"
	TransferStatusPartialIssued   TransferStatus = "PartialIssued"
	TransferStatusIssued          TransferStatus = "Issued"
	TransferStatusRejected        TransferStatus = "Rejected"
	TransferStatusClosed          TransferStatus = "Closed"
)

type Transfer struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	TrxDate                *time.Time                `form_kind:"date" form_section:"General1"`
	JournalNo              string                    `form_section:"General1"`
	JournalType            string                    `form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Inventory_Transfer|_id|_id,Name"  form_section:"General1"`
	ItemRequestID          string                    `form_section:"General1" form_lookup:"/tenant/item/find|_id|_id,Name"`
	Company                string                    `form_section:"General1"`
	Note                   string                    `form_multi_row:"5" form_section:"General2"`
	ReasonReject           string                    `form_multi_row:"5" form_section:"General2"`
	Status                 TransferStatus            `form_section:"General2" form_read_only:"1"`
	FinancialDimensionFrom tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension1"`
	FinancialDimensionTo   tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension2"`
	InventoryDimensionFrom InventDimension           `grid:"hide" form_section:"Dimension1"`
	InventoryDimensionTo   InventDimension           `grid:"hide" form_section:"Dimension2"`
	RevisionNo             int                       `grid:"hide" form:"hide"` // TODO: increment this when revision
	Created                time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate             time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *Transfer) TableName() string {
	return "Transfers"
}

func (o *Transfer) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Transfer) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Transfer) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Transfer) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Transfer) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Transfer) PostSave(dbflex.IConnection) error {
	return nil
}
