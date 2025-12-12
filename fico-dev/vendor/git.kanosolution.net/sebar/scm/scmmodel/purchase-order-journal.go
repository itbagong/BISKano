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

type PurchaseStatus string
type DiscountType string

const (
	PurchaseStatusDraft     PurchaseStatus = "Draft"
	PurchaseStatusSubmitted PurchaseStatus = "Submitted"
	PurchaseStatusPending   PurchaseStatus = "Pending"
	PurchaseStatusApproved  PurchaseStatus = "Approval"
	PurchaseStatusRejected  PurchaseStatus = "Rejected"
	PurchaseStatusClosed    PurchaseStatus = "Completed"

	DiscountTypeFixed   DiscountType = "fixed"
	DiscountTypePercent DiscountType = "percent"
)

type PurchaseOrderJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string       `bson:"_id" json:"_id" key:"1" label:"PO No" form_read_only_edit:"1" form_section_size:"4" form_section:"General" form_section_show_title:"1"`
	Name              string       `form_section:"General" form_required:"1" label:"PO Name"`
	Requestor         string       `form_required:"1" form_section:"General" form_use_list:"1" form_lookup:"/tenant/employee/find|_id|Name"` // TODO: seharusnya diganti Text
	CompanyID         string       `form_section:"General" grid:"hide" form:"hide" form_lookup:"/tenant/company/find|_id|Name"`              //hide or delete                                                  //hide or delete
	PurchaseType      PurchaseType `form_section:"General" grid:"hide" form:"hide" form_items:"STOCK|VIRTUAL|SERVICE|ASSET"`
	ReffNo            []string     `form_section:"General" form_read_only:"1" grid:"hide" label:"Ref No"`
	JournalTypeID     string       `form_required:"1" form_lookup:"/scm/purchase/order/journal/type/find|_id|_id,Name" grid:"hide" form_section:"General"`
	PostingProfileID  string       `form_section:"General" grid:"hide" form:"hide"`
	Note              string       `form_section:"General" grid:"hide" form_multi_row:"3"`

	DocumentDate *time.Time `form_section:"Date" form_kind:"date"`
	TrxDate      time.Time  `form_kind:"date" form_section:"Date" label:"Trx Date"`
	PRDate       *time.Time `form_kind:"date" form_section:"Date" form_section_show_title:"1" grid:"hide" label:"PR date" form:"hide"`
	DueDate      *time.Time `form_kind:"date" form_section:"Date" grid:"hide"`
	PODate       *time.Time `form_kind:"date" form_section:"Date" label:"PO date" form:"hide"`
	DeliveryDate *time.Time `form_kind:"date" form_section:"Date" grid:"hide" form:"hide"`

	VendorID     string `form_section:"Vendor info" form_required:"1" label:"Vendor" form_section_show_title:"1" form_lookup:"/tenant/vendor/find|_id|_id,Name"`
	VendorName   string `form_section:"Vendor info" grid:"hide" form:"hide" `
	VendorRefNo  string `form_section:"Vendor info" grid:"hide"`
	PaymentTerms string `form_section:"Vendor info" grid:"hide" form_lookup:"/fico/paymentterm/find|_id|Name"`

	TaxName         string   `form_section:"Tax" grid:"hide"`
	TaxRegistration string   `form_section:"Tax" grid:"hide" form:"hide"`
	TaxType         string   `form_section:"Tax" grid:"hide" form_section_show_title:"1" form_items:"0|1|2|3|4|5|6|7|8|9"`
	TaxCodes        []string `form_section:"Tax" grid:"hide" form_lookup:"/fico/taxcode/find|_id|Name"` // diisi dari master Vendor
	TaxAddress      string   `form_section:"Tax" grid:"hide" form_multi_row:"5"`

	WarehouseID     string `form_section:"Delivery to" grid:"hide" form:"hide"  label:"Warehouse" form_section_show_title:"1" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	PIC             string `form_section:"Delivery to" grid:"hide" label:"PIC" form_lookup:"/tenant/employee/find|_id|Name"`
	DeliveryName    string `form_section:"Delivery to" grid:"hide"`
	DeliveryAddress string `form_section:"Delivery to" form_multi_row:"5" grid:"hide"`
	BillingName     string `form_section:"Delivery to" grid:"hide"`
	BillingAddress  string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`

	Urgent     bool                    `form_section:"Info" grid:"hide" form:"hide"`
	Priority   string                  `form_section:"Info" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	Status     ficomodel.JournalStatus `form_section:"Info" form_read_only:"1"`
	Created    time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Location  InventDimension           `grid:"hide" form_section:"InventDimension"`

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

	TotalPrint int  `grid:"hide" form:"hide"`
	IsCanceled bool `grid:"hide" form:"hide"`
}

type PurchaseOrderJournalGrid struct {
	ID           string       `bson:"_id" json:"_id" key:"1" label:"PO No" form_read_only_edit:"1" form_section_size:"4" form_section:"General" form_section_show_title:"1"`
	Name         string       `form_section:"General" label:"PO Name"`
	CompanyID    string       `form_section:"General" grid:"hide"`
	TrxDate      time.Time    `form_kind:"date" form_section:"General" label:"Trx Date"`
	PODate       *time.Time   `form_kind:"date" form_section:"Date"  label:"PO Date" grid:"hide"`
	DueDate      *time.Time   `form_kind:"date" form_section:"Date"  label:"Due Date"`
	DocumentDate *time.Time   `form_section:"General" form_kind:"date"`
	ReffNo       []string     `form_section:"General" label:"Ref No"`
	WarehouseID  string       `form_section:"Delivery to" label:"Warehouse" form_section_sPhow_title:"1"`
	VendorID     string       `form_section:"Vendor info" form_section_show_title:"1" label:"Vendor"`
	PurchaseType PurchaseType `form_section:"General" form_items:"Unit|Department|Service|Asset|Stock" grid:"hide"`

	JournalTypeID    string `form_required:"1" grid:"hide" form_section:"General"`
	PostingProfileID string `form_section:"General" grid:"hide"`
	Note             string `form_section:"General" grid:"hide"`

	DeliveryName    string `form_section:"Delivery to" grid:"hide"`
	DeliveryAddress string `form_section:"Delivery to" form_multi_row:"5" grid:"hide"`
	PIC             string `form_section:"Delivery to" grid:"hide" label:"PIC"`
	BillingName     string `form_section:"Delivery to" grid:"hide"`
	BillingAddress  string `form_section:"Delivery to" grid:"hide" form_multi_row:"5"`

	PRDate       *time.Time `form_kind:"date" form_section:"Date" form_section_show_title:"1" grid:"hide" label:"PR date"`
	DeliveryDate *time.Time `form_kind:"date" form_section:"Date" grid:"hide"`

	VendorName   string `form_section:"Vendor info" grid:"hide"`
	VendorRefNo  string `form_section:"Vendor info" grid:"hide"`
	PaymentTerms string `form_section:"Vendor info" grid:"hide"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`

	TaxType         string   `form_section:"Tax" grid:"hide" form_section_show_title:"1" form_items:"0|1|2|3|4|5|6|7|8|9"`
	TaxCodes        []string `form_section:"Tax" grid:"hide"`
	TaxName         string   `form_section:"Tax" grid:"hide"`
	TaxRegistration string   `form_section:"Tax" grid:"hide"`
	TaxAddress      string   `form_section:"Tax" grid:"hide" form_multi_row:"5"`

	Location InventDimension `grid:"hide" form_section:"InventDimension"`

	TotalAmount         float64 `grid:"hide" form:"hide"`
	TotalDiscountAmount float64 `grid:"hide" form:"hide"`
	TotalTaxAmount      float64 `grid:"hide" form:"hide"`
	Freight             float64 `grid:"hide" form:"hide"`
	GrandTotalAmount    float64 `form:"hide" grid_label:"Amount"`

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

func (o *PurchaseOrderJournal) TableName() string {
	return "PurchaseOrders"
}

func (o *PurchaseOrderJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PurchaseOrderJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PurchaseOrderJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PurchaseOrderJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PurchaseOrderJournal) PreSave(dbflex.IConnection) error {
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

func (o *PurchaseOrderJournal) PostSave(dbflex.IConnection) error {
	return nil
}

type PurchaseOrderUniqueFilterParam struct {
	ItemID             string
	SKU                string
	InventoryDimension InventDimension
}

func (o *PurchaseOrderJournal) UniqueFilter(param PurchaseOrderUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}

func (o *PurchaseOrderJournal) formatID() {
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

	o.ID = strings.Replace(o.ID, "[PO_TYPE]", prType, -1)
}
