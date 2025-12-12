package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrxType string
type TrxStatus string

const (
	ItemTransactionTypeMovementIn  TrxType = "MovementIn"
	ItemTransactionTypeMovementOut TrxType = "MovementOut"
	ItemTransactionTypeTransfer    TrxType = "Transfer"
	ItemTransactionTypePurchase    TrxType = "Purchase"

	ItemTransactionStatusPlanned   TrxStatus = "Planned"
	ItemTransactionStatusConfirmed TrxStatus = "Confirmed"
)

type ItemTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_pos:"1,1" form_section:"General"`
	ReferenceID       string
	ItemID            string
	SKU               string `label:"SKU"`
	Qty               float64
	TransactionType   TrxType
	Status            TrxStatus
	CreatedBy         string    // TODO: account id dari user
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemTransaction) TableName() string {
	return "ItemTransactions"
}

func (o *ItemTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
