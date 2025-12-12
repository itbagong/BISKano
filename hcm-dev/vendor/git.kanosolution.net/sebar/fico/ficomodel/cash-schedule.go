package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashDirection string
type CashScheduleStatus string

const (
	CashReceive CashDirection = "Receive"
	CashExpense CashDirection = "Expense"

	CashDraft            CashScheduleStatus = "Draft"
	CashScheduled        CashScheduleStatus = "Scheduled"
	CashExecuted         CashScheduleStatus = "Executed"
	CashPartiallySettled CashScheduleStatus = "Partially Settled"
	CashSettled          CashScheduleStatus = "Settled"
)

type CashSchedule struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CompanyID         string
	Direction         CashDirection
	Status            CashScheduleStatus
	Account           SubledgerAccount
	Amount            float64
	CurrencyID        string
	Expected          time.Time
	ExecutionTime     *time.Time
	Text              string
	SourceType        tenantcoremodel.TrxModule
	SourceJournalID   string
	SourceLineNo      int
	VoucherNo         string
	SourceID          string
	BnfType           string
	BnfName           string
	BnfDetail         string
	Settled           float64
	Outstanding       float64
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashSchedule) TableName() string {
	return "CashSchedules"
}

func (o *CashSchedule) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashSchedule) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashSchedule) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashSchedule) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashSchedule) Calc() {
}

func (o *CashSchedule) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
		o.Outstanding = o.Amount
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.Calc()
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashSchedule) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *CashSchedule) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "Oustanding", Fields: []string{"CompanyID", "Outstanding"}},
		{Name: "OustandingByAccount", Fields: []string{"CompanyID", "Account.AccountType", "Account.AccountID", "Outstanding"}},
		{Name: "BySource", Fields: []string{"CompanyID", "SourceType", "SourceJournalID", "Outstanding"}},
	}
}
