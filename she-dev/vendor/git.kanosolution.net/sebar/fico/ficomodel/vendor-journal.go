package ficomodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VendorJournalLineReference struct {
	VendorInvoiceNo string
	TaxCodes        []string
	ChargeCodes     []string
}

type VendorJournalGrid struct {
	ID              string `json:"_id" bson:"_id"  grid_sortable:"1"`
	JournalTypeID   string `grid:"hide"`
	TransactionType string
	VendorID        string    `label:"Vendor"`
	TrxDate         time.Time `form_kind:"date"  grid_sortable:"1"`
	PaymentTermID   string    `grid:"hide"`
	Text            string
	Status          string
}

type AddressAndTax struct {
	DeliveryName    string `form_section:"Address" form_section_size:"3" form_section_show_title:"1"`
	DeliveryAddress string `form_multi_row:"5" form_section:"Address"`
	BillingName     string `form_section:"Address"`
	BillingAddress  string `form_multi_row:"5" form_section:"Address"`
	TaxType         string `form_section:"Tax" form_items:"0|1|2|3|4|5|6|7|8|9"`
	TaxName         string `form_section:"Tax" form_section_show_title:"1"`
	TaxRegistration string `form_section:"Tax"`
	TaxAddress      string `form_section:"Tax" form_multi_row:"5"`
	BankName        string `form_section:"Bank Detail" form_section_show_title:"1"`
	BankAccountNo   string `form_section:"Bank Detail" label:"Bank account No."`
	BankAccountName string `form_section:"Bank Detail"`
	BankNotes       string `form_section:"Bank Detail" label:"Notes" form_multi_row:"5"`
}

type VendorJournalForm struct {
	ID                   string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_show_title:"1"`
	JournalTypeID        string                    `form_lookup:"/fico/vendorjournaltype/find|_id|_id,Name"`
	VendorID             string                    `form_lookup:"tenant/vendor/find|_id|_id,Name"`
	CurrencyID           string                    `form_lookup:"/tenant/currency/find|_id|_id" form_read_only:"1"`
	PaymentTermID        string                    `form_lookup:"/fico/paymentterm/find|_id|Name"`
	Text                 string                    `form_required:"1" form_section:"General" form_multi_row:"2"`
	CashPayment          bool                      `form_section:"Cash Payment" form_section_show_title:"1" form:"hide"`
	TransactionType      string                    `form_items:"Vendor Purchase|Credit Note|Good Receive|Site Entry Expense|Employee Expense|General Submission" form_required:"1"`
	CashBankID           string                    `form_section:"Cash Payment" form_lookup:"/tenant/cashbank/find|_id|_id,Name" form:"hide"`
	TaxCodes             []string                  `form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1"`
	TrxDate              time.Time                 `form_section:"Invoice Information" form_kind:"date" form_section_show_title:"1" form_space_after:"1"`
	InvoiceNo            string                    `form_section:"Invoice Information"`
	ExpectedDate         time.Time                 `form_section:"Invoice Information" form_kind:"date" form_read_only:"1"`
	DeliveryDate         time.Time                 `form_section:"Invoice Information" form_kind:"date" form:"hide"`
	TaxInvoiceNo         string                    `form_section:"Invoice Information"`
	TaxInvoiceDate       time.Time                 `form_section:"Invoice Information" form_kind:"date"`
	HeaderDiscountType   string                    `form_section:"Invoice Information" form_items:"fixed|percent"`
	HeaderDiscountValue  float64                   `form_section:"Invoice Information"`
	Dimension            tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1"`
	References           codekit.M                 `form_section:"References" form_section_show_title:"1" form:"hide"`
	TotalAmount          float64                   `form_section:"Amount" label:"Grand total amount" form_read_only:"1" form_section_auto_col:"2" form_section_show_title:"1"`
	SubtotalAmount       float64                   `form_section:"Amount" form_read_only:"1"`
	PriceTotalAmount     float64                   `form_section:"Amount" form_read_only:"1"`
	TaxAmount            float64                   `form_section:"Amount" form_read_only:"1"`
	DiscountAmount       float64                   `form_section:"Amount" label:"Line discount amount" form_read_only:"1"`
	PPNAmount            float64                   `form_section:"Amount" label:"PPN Amount" form_read_only:"1"`
	HeaderDiscountAmount float64                   `form_section:"Amount" form_read_only:"1"`
	PPHAmount            float64                   `form_section:"Amount" label:"PPh Amount" form_read_only:"1"`
	Status               string                    `form_section:"Info" form_read_only:"1" form_section_show_title:"1" form_section_auto_col:"2"`
	LedgerVoucherNo      string                    `form_section:"Info" form_read_only:"1"`
	Created              time.Time                 `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
	LastUpdate           time.Time                 `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
}

func (o *VendorJournalForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Invoice Information", ShowTitle: true, AutoCol: 2},
			// {Title: "Cash Payment", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: true, AutoCol: 1},
			{Title: "Amount", ShowTitle: true, AutoCol: 2},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

type VendorJournal struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string    `bson:"_id" json:"_id" key:"1" form_read_onlyt:"1" form_section:"General" form_section_auto_col:"2"`
	JournalTypeID        string    `form_lookup:"/fico/vendorjournaltype/find|_id|Name" grid:"hide"`
	TransactionType      string    `grid:"hide" form:"hide"`
	VendorID             string    `form_lookup:"/tenant/vendor/find|_id|Name"`
	TrxDate              time.Time `form_kind:"date"`
	ExpectedDate         *time.Time
	DeliveryDate         *time.Time
	CashPayment          bool
	InclusiveTax         bool
	CashBankID           string
	Text                 string        `form_required:"1" form_section:"General"`
	CurrencyID           string        `form_lookup:"/tenant/currency/find|_id|_id"`
	Status               JournalStatus `form_read_only:"1"`
	References           tenantcoremodel.References
	ChecklistTemp        tenantcoremodel.Checklists
	Lines                []JournalLine
	InvoiceID            string `form_read_only:"1"`
	SubtotalAmount       float64
	PriceTotalAmount     float64
	TaxAmount            float64
	PPNAmount            float64 `label:"PPN Amount"`
	PPHAmount            float64 `label:"PPh Amount"`
	ChargeAmount         float64
	DiscountAmount       float64
	HeaderDiscountAmount float64
	TaxCodes             []string
	Taxes                []Charge
	Charges              []Charge
	TotalAmount          float64
	ReportingAmount      float64
	PaymentTermID        string `grid:"hide"`
	PostingProfileID     string
	AddressAndTax        AddressAndTax
	Errors               string
	CompanyID            string
	Dimension            tenantcoremodel.Dimension
	InvoiceNo            string
	TaxInvoiceNo         string
	TaxInvoiceDate       time.Time
	HeaderDiscountType   string
	HeaderDiscountValue  float64
	LedgerVoucherNo      string
	Created              time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorJournal) TableName() string {
	return "VendorJournalHeaders"
}

func (o *VendorJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorJournal) PreSave(dbflex.IConnection) error {
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
	return nil
}

func (o *VendorJournal) PreDelete(dbflex.IConnection) error {
	if o.Status != "DRAFT" {
		return errors.New("protected record, could not be deleted")
	}

	return nil
}

func (o *VendorJournal) PostSave(dbflex.IConnection) error {
	return nil
}
