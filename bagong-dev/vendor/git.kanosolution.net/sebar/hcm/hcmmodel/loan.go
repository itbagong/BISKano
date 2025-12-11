package hcmmodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoanStatus string

const (
	LoanPaid   LoanStatus = "Paid"
	LoanUnpaid LoanStatus = "Unpaid"
)

type Loan struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string    `bson:"_id" json:"_id"`
	CompanyID           string    `grid:"hide" form:"hide"`
	JournalTypeID       string    `grid:"hide" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Loan|_id|Name"`
	PostingProfileID    string    `grid:"hide" form:"hide"`
	EmployeeID          string    `grid:"hide" label:"NIK" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	RequestDate         time.Time `form_kind:"date"`
	LoanApplication     float64
	LoanPurpose         string
	LoanPeriod          int
	Installment         float64
	ApprovedLoan        float64
	ApprovedLoanPeriod  int
	ApprovedInstallment float64
	Notes               string
	AutoDebitStartDate  time.Time                 `form_kind:"date"`
	Status              ficomodel.JournalStatus   `form:"hide"`
	Lines               []LoanLine                `grid:"hide" form:"hide"`
	Dimension           tenantcoremodel.Dimension `grid:"hide"`
	Created             time.Time                 `form_kind:"datetime" grid:"hide"`
	LastUpdate          time.Time                 `form_kind:"datetime" grid:"hide"`
}

type LoanLine struct {
	Date              time.Time `form_kind:"date"`
	Period            int
	InstallmentAmount float64
	RemainingLoan     float64
	Status            LoanStatus
}

func (o *Loan) TableName() string {
	return "HCMLoans"
}

func (o *Loan) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Loan) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Loan) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Loan) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Loan) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Loan) PostSave(dbflex.IConnection) error {
	return nil
}

type LoanForm struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_section:"Employee Information" form_section_direction:"row" form_section_size:"3"`
	JournalTypeID       string                    `form_section:"Employee Information" grid:"hide" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Loan|_id|Name"`
	PostingProfileID    string                    `grid:"hide" form:"hide"`
	RequestDate         time.Time                 `form_section:"Employee Information" form_kind:"date"`
	EmployeeID          string                    `form_required:"1" form_section:"Employee Information" label:"Employee ID" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	EmployeeName        string                    `form_section:"Employee Information" form_read_only:"1"`
	NIK                 string                    `form_section:"Employee Information" form_read_only:"1" label:"NIK"`
	Position            string                    `form_section:"Employee Information2" form_read_only:"1"`
	Department          string                    `form_section:"Employee Information2" form_read_only:"1"`
	WorkLocation        string                    `form_section:"Employee Information2" form_read_only:"1"`
	EmployeeStatus      string                    `form_section:"Employee Information2" form_read_only:"1"`
	MobilePhoneNumber   string                    `form_section:"Employee Information3" form_read_only:"1"`
	PeriodOfEmployement string                    `form_section:"Employee Information3" form_read_only:"1"`
	Salary              string                    `form_section:"Employee Information3" form_read_only:"1"`
	LoanPurpose         string                    `form_section:"Loan Application" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LoanPurpose|_id|_id,Name"`
	LoanPurposeSpecify  string                    `form_section:"Loan Application" label:"Please specify purpose" form:"hide"`
	LoanApplication     float64                   `form_section:"Loan Application" label:"Loan application amount"`
	LoanPeriod          int                       `form_section:"Loan Application" label:"Loan period (month)"`
	Installment         float64                   `form_section:"Loan Application"`
	ApprovedLoan        float64                   `form_section:"Loan Application" label:"Approved loan amount"`
	ApprovedLoanPeriod  int                       `form_section:"Loan Application"`
	ApprovedInstallment float64                   `form_section:"Loan Application"`
	Notes               string                    `form_section:"Loan Application" form_multi_row:"5"`
	Status              string                    `form:"hide"`
	Dimension           tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created             time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate          time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info"`
}

func (o *LoanForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Employee Information", ShowTitle: false, AutoCol: 1},
			{Title: "Employee Information2", ShowTitle: false, AutoCol: 1},
			{Title: "Employee Information3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Loan Application", ShowTitle: true, AutoCol: 3},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *Loan) KxPreSave(ctx *kaos.Context, mdl orm.DataModel) error {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("PreSave: missing connection")
	}

	o = mdl.(*Loan)
	var date time.Time
	if o.RequestDate.Day() >= 24 {
		// add 8 days only for change to next month
		date = o.RequestDate.AddDate(0, 0, 8)
	} else {
		date = o.RequestDate
	}
	o.AutoDebitStartDate = time.Date(date.Year(), date.Month(), 23, 0, 0, 0, 0, date.Location())

	return nil
}
