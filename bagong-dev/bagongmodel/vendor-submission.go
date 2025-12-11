package bagongmodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JournalLine struct {
	JournalID        string
	LineNo           int `form_read_only:"1"`
	Account          ficomodel.SubledgerAccount
	OffsetAccount    ficomodel.SubledgerAccount
	OffsetTransRefID string
	TagObjectID1     ficomodel.SubledgerAccount
	TagObjectID2     ficomodel.SubledgerAccount
	CurrencyID       string
	LedgerDirection  tenantcoremodel.LedgerDirection
	TrxType          string
	Qty              float64
	UnitID           string
	PriceEach        float64
	Amount           float64
	Text             string
	Taxable          bool
	TaxCodes         []string
	Locked           bool
	References       tenantcoremodel.References
	Dimension        tenantcoremodel.Dimension
}

type VendorSubmissionGrid struct {
	ID              string `json:"_id" bson:"_id"`
	JournalTypeID   string
	TransactionType string
	VendorID        string    `label:"Vendor"`
	TrxDate         time.Time `form_kind:"date"`
	PaymentTermID   string
	CurrencyID      string
	TotalAmount     float64
	Text            string
	Status          string
}

type SubledgerAccount struct {
	AccountType string
	AccountID   string
}

type VendorSubmissionForm struct {
	ID              string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_show_title:"1"`
	CompanyID       string `form_lookup:"/tenant/company/find|_id|Name" form_section_size:"3"`
	JournalTypeID   string `form_lookup:"/fico/vendorjournaltype/find|_id|_id,Name"`
	TransactionType string `form_items:"Vendor Purchase|Credit Note"`
	VendorID        string `form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	CurrencyID      string `form_lookup:"/tenant/currency/find|_id|_id"`
	PaymentTermID   string `form_lookup:"/fico/paymentterm/find|_id|Name"`
	Text            string `form_required:"1" form_section:"General" form_multi_row:"2"`
	DefaultOffset   SubledgerAccount
	CashPayment     bool                      `form_section:"Cash Payment" form_section_show_title:"1"`
	CashBankID      string                    `form_section:"Cash Payment" form_lookup:"/tenant/cashbank/find|_id|_id,Name"`
	TaxCodes        []string                  `form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1"`
	TrxDate         time.Time                 `form_section:"Date" form_kind:"date" form_section_show_title:"1"`
	ExpectedDate    *time.Time                `form_section:"Date" form_kind:"date"`
	DeliveryDate    *time.Time                `form_section:"Date" form_kind:"date" form_space_after:"1"`
	Dimension       tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1"`
	References      codekit.M                 `form_section:"References" form_section_show_title:"1"`
	TotalAmount     float64                   `form_section:"Amount" form_read_only:"1" form_section_auto_col:"2" form_section_show_title:"1"`
	SubtotalAmount  float64                   `form_section:"Amount" form_read_only:"1"`
	DiscountAmount  float64                   `form_section:"Amount" form_read_only:"1"`
	TaxAmount       float64                   `form_section:"Amount" form_read_only:"1"`
	Status          string                    `form_section:"Info" form_read_only:"1" form_section_show_title:"1" form_section_auto_col:"2"`
	InvoiceNo       string                    `form_section:"Info" form_read_only:"1"`
	LedgerVoucherNo string                    `form_section:"Info" form_read_only:"1"`
	Created         time.Time                 `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
	LastUpdate      time.Time                 `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
}

func (o *VendorSubmissionForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Date", ShowTitle: true, AutoCol: 2},
			{Title: "Cash Payment", ShowTitle: true, AutoCol: 1},
			{Title: "References", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: true, AutoCol: 1},
			{Title: "Amount", ShowTitle: true, AutoCol: 2},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

type VendorSubmission struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	JournalTypeID     string `form_lookup:"/fico/vendorjournaltype/find|_id|Name"`
	TransactionType   string
	VendorID          string    `form_lookup:"/tenant/vendor/find|_id|Name"`
	TrxDate           time.Time `form_kind:"date"`
	ExpectedDate      *time.Time
	DeliveryDate      *time.Time
	DefaultOffset     ficomodel.SubledgerAccount
	CashPayment       bool
	CashBankID        string
	Text              string `form_required:"1" form_section:"General"`
	CurrencyID        string `form_lookup:"/tenant/currency/find|_id|_id"`
	Status            string `form_read_only:"1"`
	References        tenantcoremodel.References
	ChecklistTemp     tenantcoremodel.Checklists
	Lines             []JournalLine
	InvoiceID         string `form_read_only:"1"`
	SubtotalAmount    float64
	TaxAmount         float64
	ChargeAmount      float64
	DiscountAmount    float64
	TaxCodes          []string
	Taxes             []ficomodel.Charge
	Charges           []ficomodel.Charge
	TotalAmount       float64
	ReportingAmount   float64
	PaymentTermID     string
	PostingProfileID  string
	AddressAndTax     ficomodel.AddressAndTax
	Errors            string
	CompanyID         string
	Dimension         tenantcoremodel.Dimension
	RecurParam        RecuringParam
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorSubmission) SetDate(t time.Time) {
	o.TrxDate = t
}
func (o *VendorSubmission) GetDate() time.Time {
	return o.TrxDate
}
func (o *VendorSubmission) TableName() string {
	return "BGVendorSubmissionHeaders"
}

func (o *VendorSubmission) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorSubmission) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorSubmission) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorSubmission) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorSubmission) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	if o.Status == "" {
		o.Status = "DRAFT"
	}
	if o.DefaultOffset.AccountType == "" {
		o.DefaultOffset.AccountType = "string(ficomodel.SubledgerAccounting)"
	}
	return nil
}

func (o *VendorSubmission) PreDelete(dbflex.IConnection) error {
	if o.Status != "DRAFT" {
		return errors.New("protected record, could not be deleted")
	}

	return nil
}

func (o *VendorSubmission) PostSave(dbflex.IConnection) error {
	return nil
}
