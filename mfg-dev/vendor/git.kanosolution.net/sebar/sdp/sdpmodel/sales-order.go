package sdpmodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StatusSalesOrder for status sales quota
type StatusSalesOrder string

const (
	SalesOrderDraft     StatusSalesOrder = "DRAFT"
	SalesOrderSubmitted StatusSalesOrder = "SUBMITTED"
	SalesOrderApproved  StatusSalesOrder = "APPROVED"
	SalesOrderRejected  StatusSalesOrder = "REJECTED"
	SalesOrderPosted    StatusSalesOrder = "POSTED"
)

type SalesOrderLine struct {
	Asset          string
	Item           string
	Description    string
	Shift          int64
	Qty            uint
	UoM            string
	ContractPeriod int
	UnitPrice      int64
	Amount         int64
	DiscountType   string
	Discount       int
	// Account              SubledgerAccount
	Taxable        bool
	TaxCodes       []string
	Spesifications []string
	StartDate      time.Time
	EndDate        time.Time
	Checklists     tenantcoremodel.Checklists
	References     tenantcoremodel.References
}

type SalesOrderBreakdownCost struct {
	BreakdownCostItem string `form_kind:"text"`
	Amount            int32  `form_kind:"number"`
}
type SalesOrderManPower struct {
	EmployeeType string  `form_kind:"text"`
	Qty          float64 `form_kind:"number"`
}

type SalesCharge struct {
	ID     string
	Amount float64
}

type SalesTax struct {
	ID     string
	Amount float64
}
type SalesOrderLineForm struct {
	Taxable   bool                      `form_section:"Tax" form_section_show_title:"1"`
	TaxCodes  []string                  `form_section:"Tax" form_lookup:"/fico/taxcode/find|_id|Name"`
	Dimension tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1" form_read_only:"1"`
}

type SalesOrderLineGrid struct {
	Asset          string              `form_lookup:"/tenant/asset/find|_id|Name"`
	Item           string              `form_kind:"text" form_lookup:"/tenant/item/find|_id|Name"`
	Shift          uint                `form_kind:"number"`
	Description    string              `form_kind:"text" form_multi_row:"1"`
	ContractPeriod int                 `form_kind:"number"`
	Uom            tenantcoremodel.UoM `form_lookup:"/tenant/unit/find|_id|Name" form_allow_add:"1"`
	StartDate      time.Time           `form_kind:"date"`
	EndDate        time.Time           `form_kind:"date"`
	Qty            uint                `form_kind:"number"`
	UnitPrice      uint64              `form_kind:"number"`
	Amount         uint64              `form_kind:"number" form_read_only:"1"`
	DiscountType   string              `form_kind:"text" form_items:"fixed|percent"`
	Discount       int                 `form_kind:"number"`
	// Account        SubledgerAccount
	Taxable   bool                      `form_kind:"checkbox"`
	Dimension tenantcoremodel.Dimension `form_read_only:"1"`
}

type SalesOrderEditorForm struct {
	LetterHeadAsset string `form_kind:"file" form_section:"Head" form_section_show_title:"1"`
	LetterHeadFirst bool   `form_section:"Head"`
	FooterAsset     string `form_kind:"file" form_section:"Footer" form_section_show_title:"1"`
	FooterLastPage  bool   `form_section:"Footer"`
	Editor          string `form_kind:"html" form_section:"Editor" form_section_show_title:"1"`
}

type SalesOrderGrid struct {
	SalesOpportunityRefNo string                  `label:"Opportunity Ref No"`
	SalesQuotationRefNo   string                  `label:"Quotation Ref No"`
	SalesOrderNo          string                  `label:"Sales Order No"`
	SpkNo                 string                  `label:"SPK/PO/Contract No"`
	SalesOrderDate        time.Time               `label:"Sales Order Date" form_kind:"date"`
	CustomerID            string                  `label:"Customer" form_lookup:"/tenant/customer/find|_id|Name"`
	Name                  string                  `label:"Sales Order Name"`
	TotalAmount           float64                 `label:"Total Amount"`
	Status                ficomodel.JournalStatus `label:"Status"`
}

type SalesOrderLinePreviewGrid struct {
	Item           float64
	Description    float64
	Qty            float64
	Uom            float64
	ContractPeriod float64
	UnitPrice      float64
	Amount         float64
}

type SalesOrder struct {
	orm.DataModelBase `bson:"-" json:"-"`
	No                int64  `form:"hide" grid:"hide"`
	ID                string `form:"hide" grid:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_show_title:"1" form_disable:"1" form_section_size:"5"`
	// ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid:"hide" form:"hide"`

	Name string `form_required:"1" form_section:"General" label:"Sales Order Name" form_section_show_title:"1"`
	// CompanyID        string `grid:"hide" form:"hide"`
	WarehouseID string `grid:"hide" form:"hide"`
	// TaxCodes          []string         `grid:"hide" form:"hide"`
	// JournalTypeID    string        `form_required:"1" form_lookup:"/sdp/purchase/order/journal/type/find|_id|_id,Name" grid:"hide" form_section:"General" grid:"hide"`
	Charges []SalesCharge    `grid:"hide" form:"hide"`
	Lines   []SalesOrderLine `grid:"hide" form:"hide"`

	LetterHeadAsset string `grid:"hide" form:"hide"`
	LetterHeadFirst bool   `grid:"hide" form:"hide"`
	FooterAsset     string `grid:"hide" form:"hide"`
	FooterLastPage  bool   `grid:"hide" form:"hide"`
	Editor          string `grid:"hide" form:"hide"`

	BreakdownCost []SalesOrderBreakdownCost `grid:"hide" form:"hide"`
	ManPower      []SalesOrderManPower      `grid:"hide" form:"hide"`

	SalesOpportunityRefNo string    `form_section:"General" label:"Opportunity Ref No" form_section_show_title:"1" form_read_only:"1"`
	SalesQuotationRefNo   string    `form_section:"General" label:"Quotation Ref No" form_section_show_title:"1" form_read_only:"1"`
	SalesOrderNo          string    `form_section:"General" label:"Sales Order No" form_section_show_title:"1" form_read_only:"1"`
	SpkNo                 string    `label:"SPK/PO/Contract No" form_section:"General" form_section_show_title:"1"`
	SalesOrderDate        time.Time `form_section:"General" label:"Sales Order Date" form_kind:"date"`
	SalesPriceBookID      string    `grid:"hide" form_section:"General" label:"Sales Pricebook" form_section_show_title:"1" form_lookup:"/sdp/salespricebook/find|_id|Name"`
	TaxCodes              []string  `grid:"hide" form_section:"General" label:"Tax Codes" form_section_show_title:"1" form_lookup:"/fico/taxcode/find|_id|Name"`
	HeaderDiscountType    string    `form_section:"General" form_items:"fixed|percent"`
	HeaderDiscountValue   float64   `form_section:"General"`

	JournalTypeID    string `grid:"hide" form_section:"General" label:"Journal Type" form_lookup:"/sdp/salesorderjournaltype/find|_id|Name"`
	PostingProfileID string `grid:"hide" form_read_only:"1" form_section:"General" label:"Posting Profile"`
	CompanyID        string `grid:"hide" form_required:"1" form_section:"General" form_section_show_title:"1" label:"Company ID" form_section_size:"4" form_lookup:"/tenant/company/find|_id|Name"`
	Notes            string `grid:"hide" form_section:"General" form_kind:"text" form_multi_row:"2" label:"Notes"`

	CustomerID string `form_section:"Customer" form_section_show_title:"1" label:"Customer" form_section_size:"4" form_lookup:"/tenant/customer/find|_id|Name"`

	TotalAmount          float64 `form_read_only:"1" form_section:"Amount" label:"Total Amount" form_section_show_title:"1"`
	SubTotalAmount       float64 `grid:"hide" form_read_only:"1" form_section:"Amount" label:"Sub Total Amount" form_section_show_title:"1"`
	DiscountAmount       float64 `grid:"hide" form_read_only:"1" form_section:"Amount" label:"Line Discount Amount" form_section_show_title:"1"`
	HeaderDiscountAmount float64 `grid:"hide" form_read_only:"1" form_section:"Amount" label:"Header Discount Amount" form_section_show_title:"1"`
	TaxAmount            float64 `grid:"hide" form_read_only:"1" form_section:"Amount" label:"Tax Amount" form_section_show_title:"1"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Status    ficomodel.JournalStatus   `form_read_only:"1" form_section:"Dimension" label:"Status"`
	// Status    StatusSalesOrder          `form:"hide" label:"Status"`

	References tenantcoremodel.References `grid:"hide" form:"hide" form_read_only:"1"`
	Checklists tenantcoremodel.Checklists `grid:"hide" form:"hide" form_read_only:"1"`

	Year       int       `grid:"hide" form:"hide"`
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SalesOrder) TableName() string {
	return "SalesOrders"
}

func (o *SalesOrder) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SalesOrder) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SalesOrder) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SalesOrder) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SalesOrder) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.SalesOrderDate.IsZero() {
		o.SalesOrderDate = time.Now()
	}
	if o.Status == "" {
		o.Status = ficomodel.JournalStatusDraft
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SalesOrder) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *SalesOrder) PreDelete(dbflex.IConnection) error {
	if o.Status != ficomodel.JournalStatusDraft {
		return errors.New("protected record, could not be deleted")
	}

	return nil
}
