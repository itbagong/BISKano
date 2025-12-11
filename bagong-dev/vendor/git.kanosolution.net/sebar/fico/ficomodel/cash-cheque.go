package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChequeType string

const (
	IsCheque ChequeType = "CHEQUE"
	IsBG     ChequeType = "BG"
)

type CashCheque struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CashBankID        string
	BankRefNo         string
	ChequeType        ChequeType
	Amount            float64
	IssueDate         time.Time `form_kind:"date"`
	ReleaseDate       time.Time `form_kind:"date"`
	ClearingDate      time.Time `form_kind:"date"`
	Status            string
	RecipientName     string
	RecipientAddress  string
	RecipientContact  string
	CompanyID         string
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashCheque) TableName() string {
	return "CashCheques"
}

func (o *CashCheque) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashCheque) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashCheque) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashCheque) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashCheque) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashCheque) PostSave(dbflex.IConnection) error {
	return nil
}
