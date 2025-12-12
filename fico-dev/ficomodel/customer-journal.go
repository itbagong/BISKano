package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerJournalGrid struct {
	ID string `json:"_id" bson:"_id"  grid_sortable:"1"`
	//JournalTypeID   string
	//TransactionType string
	TrxDate    time.Time `form_kind:"date"  grid_sortable:"1"`
	CustomerID string    `label:"Customer"`
	//PaymentTermID   string
	Text        string
	TotalAmount float64
	Status      string
}

type CustomerJournalForm struct {
	ID               string           `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_show_title:"1"`
	CompanyID        string           `form_lookup:"/tenant/company/find|_id|Name" form_section_size:"3" form:"hide"`
	JournalTypeID    string           `form_lookup:"/fico/customerjournaltype/find|_id|_id,Name"`
	TransactionType  string           `form_items:"Customer Sales|Credit Note|Customer Deposit|Mining Invoice - Rent|Trayek Invoice|BTS Invoice|General Invoice|General Invoice - Tourism|General Invoice - Sparepart|General Invoice - Unit Sales"`
	PostingProfileID string           `form_lookup:"/fico/postingprofile/find|_id|_id,Name"  form_read_only:"1" form_label:"Posting Profile"`
	CustomerID       string           `form_lookup:"/tenant/customer/find|_id|_id,Name"`
	CurrencyID       string           `form_lookup:"/tenant/currency/find|_id|_id" form:"hide"`
	PaymentTermID    string           `form_lookup:"/fico/paymentterm/find|_id|Name"`
	Text             string           `form_section:"General" form_multi_row:"2"`
	DefaultOffset    SubledgerAccount `form:"hide"`
	CashPayment      bool             `form_section:"Cash Payment" form_section_show_title:"1" form:"hide"`
	// InclusiveTax    bool                      `form_section:"Cash Payment" form_section_show_title:"1"`
	CashBankID           string                    `form_section:"Cash Payment" form_lookup:"/tenant/cashbank/find|_id|_id,Name" form:"hide"`
	TaxCodes             []string                  `form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1"`
	TrxDate              time.Time                 `form_section:"Invoice Information" form_kind:"date" form_section_show_title:"1"`
	ExpectedDate         *time.Time                `form_section:"Invoice Information" form_kind:"date"`
	DeliveryDate         *time.Time                `form_section:"Invoice Information" form_kind:"date" form_space_after:"1"`
	TaxInvoiceNo         string                    `form_section:"Invoice Information"`
	TaxInvoiceDate       time.Time                 `form_section:"Invoice Information" form_kind:"date"`
	Dimension            tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1"`
	References           codekit.M                 `form_section:"References"  form:"hide" form_section_show_title:"1"`
	TotalAmount          float64                   `form_section:"Amount" form_read_only:"1" form_section_auto_col:"2" form_section_show_title:"1"`
	SubtotalAmount       float64                   `form_section:"Amount" form_read_only:"1"`
	DiscountAmount       float64                   `form_section:"Amount" form_read_only:"1" label:"Line discount amount"`
	TaxAmount            float64                   `form_section:"Amount" form_read_only:"1"`
	Status               string                    `form_section:"Info" form_read_only:"1" form_section_show_title:"1" form_section_auto_col:"2"`
	HeaderDiscountAmount float64                   `form_section:"Amount" form_read_only:"1"`
	HeaderDiscountType   string                    `form_section:"Invoice Information" form_items:"fixed|percent"`
	HeaderDiscountValue  float64                   `form_section:"Invoice Information"`
	// InvoiceNo       string                    `form_section:"Info" form_read_only:"1"`
	// LedgerVoucherNo string    `form_section:"Info" form_read_only:"1"`
	Created    time.Time `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
	LastUpdate time.Time `form_read_only:"1" form_section:"Info" form_pos:"10001,"`
}

func (o *CustomerJournalForm) FormSections() []suim.FormSectionGroup {
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

type CustomerJournal struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	JournalTypeID        string `form_lookup:"/fico/vendorjournaltype/find|_id|Name"`
	TransactionType      string
	CustomerID           string    `form_lookup:"/tenant/customer/find|_id|Name" form:"hide"`
	TrxDate              time.Time `form_kind:"date"`
	ExpectedDate         *time.Time
	DeliveryDate         *time.Time
	DefaultOffset        SubledgerAccount
	CashPayment          bool `form:"hide"`
	InclusiveTax         bool
	CashBankID           string        `form:"hide"`
	Text                 string        `form_section:"General"`
	CurrencyID           string        `form_lookup:"/tenant/currency/find|_id|_id"`
	Status               JournalStatus `form_read_only:"1"`
	References           tenantcoremodel.References
	ChecklistTemp        tenantcoremodel.Checklists
	Lines                []JournalLine
	InvoiceID            string `form_read_only:"1"`
	SubtotalAmount       float64
	TaxAmount            float64
	ChargeAmount         float64
	DiscountAmount       float64
	TaxCodes             []string
	Taxes                []Charge
	Charges              []Charge
	TotalAmount          float64
	ReportingAmount      float64
	PaymentTermID        string
	PostingProfileID     string
	AddressAndTax        AddressAndTax
	Errors               string
	CompanyID            string
	Dimension            tenantcoremodel.Dimension
	HeaderDiscountAmount float64
	HeaderDiscountType   string
	HeaderDiscountValue  float64
	TaxInvoiceNo         string
	TaxInvoiceDate       time.Time
	Created              time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CustomerJournal) TableName() string {
	return "CustomerJournals"
}

func (o *CustomerJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomerJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomerJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomerJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomerJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomerJournal) PostSave(dbflex.IConnection) error {
	return nil
}
