package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashBank string

const (
	CASH_ONLY CashBank = "Cash Only"
	BANK_ONLY CashBank = "Bank Only"
	CASH_BANK CashBank = "Cash+Bank"

	FREQ_DAY   = "DAY"
	FREQ_MONTH = "MONTH"
	FREQ_WEEK  = "WEEK"
	FREQ_YEAR  = "YEAR"
)

type SubmissionLines struct {
	Text   string
	Amount int
}
type PayrollSubmissionCustom struct {
	ID            string                    `bson:"_id" json:"_id" key:"1" grid_pos:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"3" form_section_direction:"row"`
	TrxDate       time.Time                 `form_kind:"date" form_section:"General" grid_pos:"3" grid_label:"Transaction Date"`
	SiteID        string                    `form:"hide" grid_pos:"2"`
	JournalTypeID string                    `form:"hide" grid:"hide"`
	Dimension     tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Expense       string                    `grid:"hide" form_lookup:"/tenant/expensetype/find|_id|Name"`
	Text          string                    `form_multi_row:"5" grid_pos:"4"`
	TotalAmount   float64                   `form:"hide" grid_pos:"5"`
	Status        string                    `form:"hide" grid_pos:"6"`
	Created       time.Time                 `form_section:"Time Info" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate    time.Time                 `form_section:"Time Info" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PayrollSubmissionCustom) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 2},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

type RecurencePayroll struct {
	StartDate       time.Time
	EndDate         time.Time
	RecurenceAmount int
	RecurenceCount  int
	Frequency       string
}
type BGPayrollSubmission struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Text              string
	SubmissionDate    time.Time
	SiteID            string
	PayrollDetailIDs  []string // store ID of BGPayrollDetail
	JournalType       string
	CashBank          CashBank `form_items:"Cash Only|Bank Only|Cash+Bank"`
	Status            string
	Total             float64
	Recurence         RecurencePayroll `form_items:"DAY|MONTH|WEEK|YEAR"`
	Dimension         tenantcoremodel.Dimension
	ApprovalStatus    string

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *BGPayrollSubmission) TableName() string {
	return "BGPayrollSubmission"
}

func (o *BGPayrollSubmission) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BGPayrollSubmission) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BGPayrollSubmission) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BGPayrollSubmission) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BGPayrollSubmission) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BGPayrollSubmission) PostSave(dbflex.IConnection) error {
	return nil
}
