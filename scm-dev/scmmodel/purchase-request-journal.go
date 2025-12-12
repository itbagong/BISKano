package scmmodel

import (
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseRequestJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string       `bson:"_id" label:"PR No" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General" form_section_show_title:"1" form_pos:"1"`
	Name              string       `form_section:"General" label:"PR Name" form_required:"1"` // TODO: seharusnya diganti Text
	Requestor         string       `form_required:"1" form_section:"General" form_use_list:"1" form_lookup:"/tenant/employee/find|_id|Name"`
	CompanyID         string       `form_section:"General" grid:"hide" form:"hide" form_lookup:"/tenant/company/find|_id|Name"` //hide or delete                          //hide or delete
	PurchaseType      PurchaseType `form_section:"General" grid:"hide" form:"hide" form_items:"STOCK|VIRTUAL|SERVICE|ASSET"`
	ReffNo            []string     `form_section:"General" grid:"hide" form_read_only:"1" form_lookup:"/scm/item/request/find?Status=POSTED|_id|_id,Name" multiple:"1" label:"IR Ref No"`
	POReff            []string     `form_section:"General" grid:"hide" form:"hide" multiple:"1" label:"PO Ref"`
	JournalTypeID     string       `form_required:"1" form_lookup:"/scm/purchase/order/journal/type/find|_id|_id,Name" grid:"hide" form_section:"General"`
	PostingProfileID  string       `form_section:"General" grid:"hide" form:"hide"`
	Note              string       `form_section:"General" grid:"hide" form_multi_row:"3"`

	DocumentDate *time.Time `form_section:"General" form_kind:"date" grid:"hide" form:"hide"`
	TrxDate      time.Time  `form_kind:"date" form_section:"Date" label:"Trx Date"`
	PRDate       *time.Time `form_kind:"date" form_section:"Date" form_section_show_title:"1" label:"Expected Date"`
	ExpectedDate *time.Time `form_kind:"date" form_section:"Date" grid:"hide" form:"hide"`
	References   []string   `form_section:"General" grid:"hide" form:"hide"  multiple:"1" label:"References"`

	VendorID     string `form_section:"Vendor info" label:"Vendor" form_section_show_title:"1" form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	VendorName   string `form_section:"Vendor info" grid:"hide" form:"hide" `
	VendorRefNo  string `form_section:"Vendor info" grid:"hide"`
	PaymentTerms string `form_section:"Vendor info" grid:"hide" form_lookup:"/fico/paymentterm/find|_id|Name"`

	TaxName         string   `form_section:"Tax" grid:"hide"`
	TaxRegistration string   `form_section:"Tax" grid:"hide" form:"hide"`
	TaxType         string   `form_section:"Tax" grid:"hide" form_section_show_title:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=TTY|_id|_id,Name"`
	TaxCodes        []string `form_section:"Tax" grid:"hide" form_lookup:"/fico/taxcode/find|_id|Name"`
	TaxAddress      string   `form_section:"Tax" grid:"hide" form_multi_row:"5"`

	WarehouseID     string `form_section:"Delivery to" grid:"hide" form:"hide" form_section_show_title:"1" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	PIC             string `form_section:"Delivery to" grid:"hide" label:"PIC"`
	DeliveryName    string `form_section:"Delivery to" grid:"hide"`
	DeliveryAddress string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`
	BillingName     string `form_section:"Delivery to" grid:"hide"`
	BillingAddress  string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`

	Urgent     bool                    `form_section:"Info" grid:"hide" form:"hide"`
	Priority   string                  `form_section:"Info" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	Status     ficomodel.JournalStatus `form_section:"Info" form_read_only:"1"`
	Created    time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_pos:"3"`
	Location  InventDimension           `grid:"hide" form_section:"InventDim" form_pos:"4"`

	TotalAmount         float64          `grid:"hide" form:"hide"`
	TotalDiscountAmount float64          `grid:"hide" form:"hide"`
	TotalTaxAmount      float64          `grid:"hide" form:"hide"`
	Freight             float64          `grid:"hide" form:"hide"`
	PPN                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPNTaxCodes
	PPH                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPHTaxCodes
	OtherExpenses       []OtherExpenses  `grid:"hide" form:"hide"`
	Discount            PurchaseDiscount `form_section:"Tax" grid:"hide" form:"hide" form_section_show_title:"1"`
	GrandTotalAmount    float64          `form:"hide" grid_label:"Amount"`

	Lines        []PurchaseJournalLine `grid:"hide" form:"hide"`
	AttachmentID string                `grid:"hide" form:"hide"`
	Text         string                `grid:"hide" form:"hide"` // biar ga error aja karena ada penambahan baru

	TotalPrint int `grid:"hide" form:"hide"`
}

type PurchaseRequestJournalWithLine struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string       `bson:"_id" label:"PR No" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General" form_section_show_title:"1" form_pos:"1"`
	Name              string       `form_section:"General" label:"PR Name" form_required:"1"` // TODO: seharusnya diganti Text
	Requestor         string       `form_required:"1" form_section:"General" form_use_list:"1" form_lookup:"/tenant/employee/find|_id|Name"`
	CompanyID         string       `form_section:"General" grid:"hide" form:"hide" form_lookup:"/tenant/company/find|_id|Name"` //hide or delete                          //hide or delete
	PurchaseType      PurchaseType `form_section:"General" grid:"hide" form:"hide" form_items:"STOCK|VIRTUAL|SERVICE|ASSET"`
	ReffNo            []string     `form_section:"General" grid:"hide" form_read_only:"1" form_lookup:"/scm/item/request/find?Status=POSTED|_id|_id,Name" multiple:"1" label:"IR Ref No"`
	POReff            []string     `form_section:"General" grid:"hide" form:"hide" multiple:"1" label:"PO Ref"`
	JournalTypeID     string       `form_required:"1" form_lookup:"/scm/purchase/order/journal/type/find|_id|_id,Name" grid:"hide" form_section:"General"`
	PostingProfileID  string       `form_section:"General" grid:"hide" form:"hide"`
	Note              string       `form_section:"General" grid:"hide" form_multi_row:"3"`

	DocumentDate *time.Time `form_section:"General" form_kind:"date" grid:"hide" form:"hide"`
	TrxDate      time.Time  `form_kind:"date" form_section:"Date" label:"Trx Date"`
	PRDate       *time.Time `form_kind:"date" form_section:"Date" form_section_show_title:"1" label:"Expected Date"`
	ExpectedDate *time.Time `form_kind:"date" form_section:"Date" grid:"hide" form:"hide"`
	References   []string   `form_section:"General" grid:"hide" form:"hide"  multiple:"1" label:"References"`

	VendorID     string `form_section:"Vendor info" label:"Vendor" form_section_show_title:"1" form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	VendorName   string `form_section:"Vendor info" grid:"hide" form:"hide" `
	VendorRefNo  string `form_section:"Vendor info" grid:"hide"`
	PaymentTerms string `form_section:"Vendor info" grid:"hide" form_lookup:"/fico/paymentterm/find|_id|Name"`

	TaxName         string   `form_section:"Tax" grid:"hide"`
	TaxRegistration string   `form_section:"Tax" grid:"hide" form:"hide"`
	TaxType         string   `form_section:"Tax" grid:"hide" form_section_show_title:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=TTY|_id|_id,Name"`
	TaxCodes        []string `form_section:"Tax" grid:"hide" form_lookup:"/fico/taxcode/find|_id|Name"`
	TaxAddress      string   `form_section:"Tax" grid:"hide" form_multi_row:"5"`

	WarehouseID     string `form_section:"Delivery to" grid:"hide" form:"hide" form_section_show_title:"1" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	PIC             string `form_section:"Delivery to" grid:"hide" label:"PIC"`
	DeliveryName    string `form_section:"Delivery to" grid:"hide"`
	DeliveryAddress string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`
	BillingName     string `form_section:"Delivery to" grid:"hide"`
	BillingAddress  string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`

	Urgent     bool                    `form_section:"Info" grid:"hide" form:"hide"`
	Priority   string                  `form_section:"Info" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	Status     ficomodel.JournalStatus `form_section:"Info" form_read_only:"1"`
	Created    time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_pos:"3"`
	Location  InventDimension           `grid:"hide" form_section:"InventDim" form_pos:"4"`

	TotalAmount         float64          `grid:"hide" form:"hide"`
	TotalDiscountAmount float64          `grid:"hide" form:"hide"`
	TotalTaxAmount      float64          `grid:"hide" form:"hide"`
	Freight             float64          `grid:"hide" form:"hide"`
	PPN                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPNTaxCodes
	PPH                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPHTaxCodes
	OtherExpenses       []OtherExpenses  `grid:"hide" form:"hide"`
	Discount            PurchaseDiscount `form_section:"Tax" grid:"hide" form:"hide" form_section_show_title:"1"`
	GrandTotalAmount    float64          `form:"hide" grid_label:"Amount"`

	Lines        PurchaseJournalLine `grid:"hide" form:"hide"`
	AttachmentID string              `grid:"hide" form:"hide"`
	Text         string              `grid:"hide" form:"hide"` // biar ga error aja karena ada penambahan baru

	TotalPrint int `grid:"hide" form:"hide"`
}

type PurchaseRequestJournalGrid struct {
	ID               string       `bson:"_id" json:"_id" key:"1" label:"PR No" form_read_only_edit:"1" form_section_size:"4" form_section:"General" form_section_show_title:"1" form_pos:"1"`
	Name             string       `form_section:"General" label:"PR Name"`
	TrxDate          time.Time    `form_kind:"date" form_section:"General" label:"Trx Date"`
	PRDate           *time.Time   `form_kind:"date" form_section:"Date" form_section_show_title:"1" form_pos:"2" label:"PR date"`
	ReffNo           []string     `form_section:"General" multiple:"1" label:"IR Ref"`
	POReff           []string     `form_section:"General" multiple:"1" label:"PO Ref"`
	WarehouseID      string       `form_section:"Delivery to" form:"hide" form_section_show_title:"1" label:"Warehouse"`
	VendorID         string       `form_section:"Vendor info" form_section_show_title:"1" label:"Vendor"`
	CompanyID        string       `form_section:"General" grid:"hide" `
	DocumentDate     *time.Time   `form_section:"General" form_kind:"date" grid:"hide"`
	PurchaseType     PurchaseType `form_section:"General" form_items:"Unit|Department|Service|Asset|Stock" grid:"hide"`
	JournalTypeID    string       `form_required:"1" grid:"hide" form_section:"General"`
	PostingProfileID string       `form_section:"General" grid:"hide"`
	Note             string       `form_section:"General" grid:"hide"`

	DeliveryName    string `form_section:"Delivery to" grid:"hide"`
	DeliveryAddress string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`
	PIC             string `form_section:"Delivery to" grid:"hide" label:"PIC" form_lookup:"/tenant/employee/find|_id|Name"`
	BillingName     string `form_section:"Delivery to" grid:"hide"`
	BillingAddress  string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`

	ExpectedDate *time.Time `form_kind:"date" grid:"hide" form_section:"Date"`
	VendorName   string     `form_section:"Vendor info" grid:"hide" `
	VendorRefNo  string     `form_section:"Vendor info" grid:"hide"`
	PaymentTerms string     `form_section:"Vendor info" grid:"hide"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_pos:"3"`

	TaxType         string   `form_section:"Tax" grid:"hide" form_section_show_title:"1" form_items:"0|1|2|3|4|5|6|7|8|9"`
	TaxCodes        []string `form_section:"Tax" grid:"hide"` // diisi dari master Vendor
	TaxName         string   `form_section:"Tax" grid:"hide"`
	TaxRegistration string   `form_section:"Tax" grid:"hide" form_kind:"date"`
	TaxAddress      string   `form_section:"Tax" grid:"hide" form_multi_row:"5"`

	Location InventDimension `grid:"hide" form_section:"InventDim" form_pos:"4"`

	TotalAmount         float64          `grid:"hide" form:"hide"`
	TotalDiscountAmount float64          `grid:"hide" form:"hide"`
	TotalTaxAmount      float64          `grid:"hide" form:"hide"`
	Freight             float64          `grid:"hide" form:"hide"`
	PPN                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPNTaxCodes
	PPH                 float64          `grid:"hide" form:"hide"` // penjumlahan rupiah semua PPHTaxCodes
	OtherExpenses       []OtherExpenses  `grid:"hide" form:"hide"`
	Discount            PurchaseDiscount `form_section:"Tax" grid:"hide" form:"hide" form_section_show_title:"1"`
	GrandTotalAmount    float64          `form:"hide" grid_label:"Amount"`

	Created    time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	Urgent     bool                    `form_section:"Info" grid:"hide"`
	Priority   string                  `form_section:"Info"`
	Approvers  []string                `form:"hide" label:"Next Approval"`
	Status     ficomodel.JournalStatus `form_section:"Info" form_read_only:"1"`

	Lines        []PurchaseJournalLine `grid:"hide" form:"hide"`
	AttachmentID string                `grid:"hide" form:"hide"`

	TotalPrint int `grid:"hide" form:"hide"`
}

type PurchaseJournalLine struct {
	InventJournalLine
	ID                     string                `grid:"hide" form:"hide"`                        // hanya utk identifier saja: dari GR bisa me-refer balik ke PO Line (disimpan di InventTrx.References)
	OffsetAccount          string                `grid:"hide" form:"hide" form_section:"General"` // ini apa ya isinya ???
	Taxable                bool                  `form_section:"General2"`
	DiscountType           DiscountType          `form_section:"General2" form_items:"fixed|percent"`
	DiscountValue          float64               `form_section:"General2"`                                       // fixed: in rupiah (same as DiscountAmount), percent: in percentage
	DiscountAmount         float64               `form_section:"General2" form_read_only:"1" grid_read_only:"1"` // in rupiah
	DiscountGeneral        PurchaseDiscount      `grid:"hide"`
	SubTotalBeforeDiscount float64               `form_section:"General2" grid:"hide"`
	SubTotal               float64               `form_section:"General2" form_read_only:"1" grid_read_only:"1"` // after discount
	PRID                   string                `form_section:"General" grid:"hide" form:"hide"`                // digunakan di PO saja
	TaxCodes               []string              `form_section:"General" grid:"hide" form:"hide"`                // digunakan di PO saja
	PPNTaxCodes            []string              `form_section:"General" grid:"hide" form:"hide"`                // setelah di filter berdasarkan TaxGroupID: TG001
	PPHTaxCodes            []string              `form_section:"General" grid:"hide" form:"hide"`                // setelah di filter berdasarkan TaxGroupID: TG002
	PPN                    float64               `label:"PPN"`                                                   // penjumlahan rupiah semua PPNTaxCodes
	PPH                    float64               `label:"PPh"`                                                   // penjumlahan rupiah semua PPHTaxCodes
	SourceLineNo           int                   `label:"Source Line No"`
	Remarks                string                `form_section:"General" form_multi_row:"2"`      // LineNo yang diambil dari PR Line No
	Item                   *tenantcoremodel.Item `form_section:"General" grid:"hide" form:"hide"` // filled in MW: used for UI
	ReceivedQty            float64               `form_section:"General" form:"hide"`             // filled in MW: used for UI
}

type PurchaseDiscount struct {
	DiscountType   DiscountType `form_section:"General2" form_items:"fixed|percent"`
	DiscountValue  float64      `form_section:"General2"` // fixed: in rupiah (same as DiscountAmount), percent: in percentage
	DiscountAmount float64      `form_section:"General2"`
}

type OtherExpenses struct {
	Expenses string  `grid:"hide" form:"hide"`
	Amount   float64 `grid:"hide" form:"hide"`
}

func (o *PurchaseRequestJournal) TableName() string {
	return "PurchaseRequests"
}

func (o *PurchaseRequestJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PurchaseRequestJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PurchaseRequestJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PurchaseRequestJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PurchaseRequestJournal) PreSave(dbflex.IConnection) error {
	o.formatID()
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PurchaseRequestJournal) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *PurchaseRequestJournal) formatID() {
	typeM := map[PurchaseType]string{
		PurchaseTypeStock:   "01",
		PurchaseTypeVirtual: "02",
		PurchaseTypeService: "03",
		PurchaseTypeAsset:   "04",
	}

	prType := "XX"
	if v, ok := typeM[o.PurchaseType]; ok {
		prType = v
	}

	o.ID = strings.Replace(o.ID, "[PR_TYPE]", prType, -1)
}
