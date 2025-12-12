package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaxStatus string

const (
	TaxOpen      TaxStatus = "Open"
	TaxAllocated TaxStatus = "Allocated"
	TaxSubmitted TaxStatus = "Submitted"
)

type TaxTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	SourceType        string
	SourceJournalID   string
	SourceTrxType     string
	SourceAccount     SubledgerAccount
	TaxCode           string
	Amount            float64
	FPNo              string
	FPDate            *time.Time
	Status            TaxStatus
	CompanyID         string
	VoucherNo         string
	InvoiceOperation  TaxInvoiceAmountOperation
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TaxTransaction) TableName() string {
	return "TaxTransactions"
}

func (o *TaxTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TaxTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TaxTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TaxTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TaxTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TaxTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
