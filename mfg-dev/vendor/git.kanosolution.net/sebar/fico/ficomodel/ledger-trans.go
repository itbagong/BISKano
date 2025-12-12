package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AmountStatus string

const (
	AmountConfirmed AmountStatus = "CONFIRMED"
	AmountReserved  AmountStatus = "RESERVED"
	AmountPlanned   AmountStatus = "PLANNED"
)

type LedgerTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	VoucherNo         string
	Offset            bool
	Account           tenantcoremodel.LedgerAccount
	Expense           *tenantcoremodel.ExpenseType
	CurrencyID        string
	Amount            float64
	Status            AmountStatus
	ReportingAmount   float64
	//Direction         tenantcoremodel.LedgerDirection
	Text              string
	TrxDate           time.Time
	SourceType        tenantcoremodel.TrxModule
	SourceJournalType string
	SourceJournalID   string
	SourceLineNo      int
	SourceTrxType     string
	References        tenantcoremodel.References
	PrepareUserID     string
	PostingUserID     string
	CompanyID         string
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LedgerTransaction) TableName() string {
	return "LedgerTrans"
}

func (o *LedgerTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LedgerTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LedgerTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LedgerTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LedgerTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LedgerTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
