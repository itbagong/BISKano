package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	CashBank          tenantcoremodel.CashBank
	TrxDate           time.Time
	SourceType        tenantcoremodel.TrxModule
	SourceJournalType string
	SourceJournalID   string
	SourceLineNo      int
	VoucherNo         string
	TrxType           string
	Amount            float64
	Text              string
	CompanyID         string
	Status            AmountStatus
	Dimension         tenantcoremodel.Dimension
	ChequeGiroID      string
	CashReconID       string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashTransaction) TableName() string {
	return "CashTransactions"
}

func (o *CashTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
