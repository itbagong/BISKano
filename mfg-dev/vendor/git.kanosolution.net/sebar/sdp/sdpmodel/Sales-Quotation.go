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

// DiscountTypeSQ for type discount
type DiscountTypeSQ string

const (
	// DiscountFixed for fixed discount (int)
	DiscountFixed DiscountTypeSQ = "fixed"
	// DiscountPercentage for percentage discount (0-100)
	DiscountPercentage DiscountTypeSQ = "percentage"
)

// SalesQuotationForm for form
type SalesQuotationForm struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_show_title:"1" form_disable:"1" form_section_size:"5"  grid:"hide" form:"hide"`

	// LetterHeadAsset string `form_kind:"file" form_section:"Asset" form_section_width:"100%"`
	// FooterAsset     string `form_kind:"file" form_section:"Asset"`
	// LetterHeadFirst bool   `form_kind:"checkbox" form_section:"Asset"`
	// FooterLastPage  bool   `form_kind:"checkbox" form_section:"Asset"`

	OpportunityNo       string    `form_label:"Sales Opportunity Ref No." form_section:"General" form_read_only:"1"`
	QuotationNo         string    `form_label:"Quotation No" form_section:"General" form_read_only:"1"`
	QuotationDate       time.Time `form_kind:"date"  form_label:"Quotation Date" form_section:"General" form_read_only:"1"`
	QuotationName       string    `form_label:"Quotation Name" form_section:"General"`
	SalesPriceBook      string    `form_section:"General" form_label:"Sales Price Book" form_lookup:"/sdp/salespricebook/find|_id|Name"`
	TaxCodes            []string  `form_section:"General" label:"Tax Codes" form_lookup:"/fico/taxsetup/find|_id|Name"`
	HeaderDiscountType  string    `form_section:"General" form_items:"fixed|percent"`
	HeaderDiscountValue float64   `form_section:"General"`
	JournalType         string    `form_section:"General" label:"Journal Type" form_lookup:"/sdp/salesorderjournaltype/find|_id|Name"`
	PostingProfileID    string    `grid:"hide" form_read_only:"1" form_section:"General" label:"Posting Profile"`

	Customer string `form_section:"Quotation For" form_width:"100%" form_section_show_title:"1" form_read_only_edit:"1" form_lookup:"/tenant/customer/find|_id|Name" form_section_width:"30px"`
	Name     string `form_section:"Quotation For" form_read_only:"1"`
	Address  string `form_section:"Quotation For" form_read_only:"1"`
	City     string `form_section:"Quotation For" form_read_only:"1"`
	Province string `form_section:"Quotation For" form_read_only:"1"`
	Country  string `form_section:"Quotation For" form_read_only:"1"`
	Zipcode  string `form_section:"Quotation For" form_read_only:"1"`

	AddressDelivery  string `form_section:"Delivery Address" form_label:"Address" form_width:"100%" form_section_show_title:"1" form_read_only:"1" form_section_width:"30px"`
	CityDelivery     string `form_section:"Delivery Address" form_label:"City" form_read_only:"1"`
	ProvinceDelivery string `form_section:"Delivery Address" form_label:"Province" form_read_only:"1"`
	CountryDelivery  string `form_section:"Delivery Address" form_label:"Country" form_read_only:"1"`
	ZipcodeDelivery  string `form_section:"Delivery Address" form_label:"Zipcode" form_read_only:"1"`

	TotalAmount          int64   `form_section:"Amount" form_kind:"number" form_section_show_title:"1" form_read_only:"1"`
	SubTotalAmount       int64   `form_section:"Amount" form_kind:"number" form_read_only:"1"`
	DiscountAmount       float64 `form_section:"Amount" form_kind:"number" form_read_only:"1" label:"Line Discount Amount"`
	HeaderDiscountAmount float64 `form_section:"Amount" form_kind:"number" form_read_only:"1"`
	TaxAmount            float64 `form_section:"Amount" form_kind:"number" form_read_only:"1"`

	Dimension tenantcoremodel.Dimension `form_section:"Dimension"`

	// InventoryDimension InventoryDimension `form_section:"Specification"`
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Amount"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Amount"`
}

// SalesQuotationGrid for view list table
type SalesQuotationGrid struct {
	OpportunityNo string
	QuotationNo   string
	QuotationDate time.Time
	Customer      string
	QuotationName string
	TotalAmount   int64 `form_kind:"number"`
	Status        ficomodel.JournalStatus
}

// SalesQuotation model for Sales Quotation
type SalesQuotation struct {
	orm.DataModelBase `bson:"-" json:"-"`

	ID string `bson:"_id" json:"_id" key:"1"`

	LetterHeadAsset string
	FooterAsset     string
	LetterHeadFirst bool
	FooterLastPage  bool
	Editor          string

	No             int64
	Rev            int64
	OpportunityNo  string
	QuotationNo    string
	QuotationDate  time.Time
	QuotationName  string
	SalesPriceBook string

	Customer string

	TotalAmount          float64
	SubTotalAmount       int64
	DiscountAmount       float64
	HeaderDiscountAmount float64
	TaxAmount            float64

	CompanyID string

	TaxCodes            []string
	HeaderDiscountType  string
	HeaderDiscountValue float64
	JournalType         string
	PostingProfileID    string

	Lines []struct {
		Asset          string
		Item           string
		Description    string
		Shift          int64
		Qty            uint
		UoM            string
		ContractPeriod int
		UnitPrice      int64
		Amount         int64
		DiscountType   DiscountTypeSQ
		Discount       int
		Taxable        bool
		TaxCodes       []string
		Spesifications []string
	}

	Dimension tenantcoremodel.Dimension
	Status    ficomodel.JournalStatus
	SendEmail bool

	// InventoryDimension InventoryDimension `form_section:"Specification"`
	Created    time.Time
	LastUpdate time.Time
}

func (o *SalesQuotation) TableName() string {
	return "SalesQuotation"
}

func (o *SalesQuotation) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SalesQuotation) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SalesQuotation) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SalesQuotation) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SalesQuotation) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == "" {
		o.Status = ficomodel.JournalStatusDraft
	}

	for index, line := range o.Lines {
		if line.Item != "" {
			o.Lines[index].Asset = ""
		}
		if line.Asset != "" {
			o.Lines[index].Item = ""
		}

		// if line.DiscountType == "" {
		// 	o.Lines[index].DiscountType = DiscountFixed
		// }
	}

	if o.Customer == "" {
		return errors.New("Constumer not null")
	}

	if o.Rev < 0 {
		o.Rev = 0
	}

	if o.No < 0 {
		result := map[string]any{}

		customer := tenantcoremodel.Customer{}
		customerresult := map[string]any{}
		sqlcustomer := dbflex.
			From(customer.TableName()).
			Take(1).
			Where(dbflex.Eq("_id", o.Customer))
		err := conn.Cursor(sqlcustomer, nil).Fetch(customerresult).Close()

		if err != nil {
			return err
		}

		sqldb := dbflex.
			From(o.TableName()).
			Select("No").
			Take(1).
			Where(dbflex.And(dbflex.Eq("Customer", o.Customer), dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)))).
			OrderBy("-No")
		err = conn.Cursor(sqldb, nil).Fetch(result).Close()

		if err != nil && !errors.Is(err, dbflex.ErrEOF) {
			return err
		}

		no := int64(-1)

		if result["No"] != nil {
			no = result["No"].(int64)
		}

		o.No = no + 1
	}

	if o.QuotationDate.IsZero() {
		o.QuotationDate = time.Now()
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *SalesQuotation) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *SalesQuotation) PreDelete(dbflex.IConnection) error {
	if o.Status != ficomodel.JournalStatusDraft {
		return errors.New("protected record, could not be deleted")
	}

	return nil
}
