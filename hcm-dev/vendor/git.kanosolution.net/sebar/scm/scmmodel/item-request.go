package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemRequestStatus string

const (
	ItemRequestStatusDraft     ItemRequestStatus = "Draft"
	ItemRequestStatusSubmitted ItemRequestStatus = "Submitted"
)

type ItemRequest struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General1" form_section_size:"3" label:"IR No"`
	CompanyID         string `form_section:"General1" grid:"hide" form_lookup:"/tenant/company/find|_id|Name"  form:"hide"`
	Name              string `form_required:"1" form_section:"General1" label:"IR Name"` // TODO: seharusnya diganti Text
	// RequestDate       *time.Time `form_kind:"date" form_section:"General1"`
	// DocumentDate      *time.Time `form_kind:"date" grid:"hide" form:"hide" form_section:"General1"`
	TrxDate          time.Time `form_kind:"date" form_section:"General1" label:"Request Date"`
	WOReff           string    `form_section:"General1" form_read_only:"1" form_use_list:"1" form_lookup:"/mfg/workorder/find?Status=IN PROGRESS|_id|_id" label:"Work Order Reference"`
	JournalTypeID    string    `form_section:"General1" grid:"hide"`
	PostingProfileID string    `form_section:"General1" grid:"hide" form:"hide"` // hide di UI, proses pengisian dari journal type di backend
	Requestor        string    `form_required:"1" form_section:"General2" form_use_list:"1" form_lookup:"/tenant/employee/find|_id|Name"`
	// Department        string                    `form_section:"General2"` // di hilangkan karena sudah di akomodir oleh Dimension CC
	Remarks     string                    `form_section:"General2" form_multi_row:"2" grid:"hide"`
	TrxType     InventTrxType             `form_section:"General2" form:"hide" grid:"hide" form_items:"Movement In|Movement Out|Transfer|Stock Opname"`
	Priority    string                    `form_required:"1" form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	Status      ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	Created     time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate  time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	Dimension   tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	InventDimTo InventDimension           `grid:"hide" form_section:"Dimension"`

	Text        string   `grid:"hide" form:"hide"` // biar ga error aja karena ada penambahan baru
	Approvers   []string `grid:"hide" form:"hide"` // for UI only
	FulfilledBy string   `grid:"hide" form:"hide"` // for UI only
}

type ItemRequestGridUI struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General1" form_section_size:"4" label:"IR No"`
	CompanyID         string `form_section:"General1" grid:"hide" form:"hide"`
	Name              string `form_section:"General1" label:"IR Name"`
	// RequestDate       *time.Time                `form_kind:"date" form_section:"General1"`
	// DocumentDate      *time.Time                `form_kind:"date" grid:"hide" form:"hide" form_section:"General1"`
	TrxDate          time.Time `form_kind:"date" form_section:"General1" label:"Request Date"`
	WOReff           string    `form_section:"General1" form_use_list:"1" label:"WO Ref"`
	JournalTypeID    string    `form_section:"General1" grid:"hide"`
	PostingProfileID string    `form_section:"General1" grid:"hide"`
	Requestor        string    `form_section:"General2" form_use_list:"1"` // ngambil dari Login (userID)
	// Department        string                    `form_section:"General2"`                   // ngambil dari Login (companyID)
	Remarks     string                    `form_section:"General2" form_multi_row:"2" grid:"hide"`
	TrxType     InventTrxType             `form_section:"General2" form:"hide" grid:"hide" form_items:"Movement In|Movement Out|Transfer|Stock Opname"`
	Priority    string                    `form_section:"General2"`
	FulfilledBy string                    `form:"hide" label:"Fulfilled By"`
	Approvers   []string                  `form:"hide" label:"Next Approval"`
	Status      ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	InventDimTo InventDimension           `grid:"hide" form_section:"General3"`
	Dimension   tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created     time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate  time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemRequest) TableName() string {
	return "ItemRequests"
}

func (o *ItemRequest) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemRequest) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemRequest) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemRequest) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemRequest) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemRequest) PostSave(dbflex.IConnection) error {
	return nil
}

type ItemRequestUniqueFilterParam struct {
	ItemID             string
	SKU                string `label:"SKU"`
	InventoryDimension InventDimension
}

func (o *ItemRequest) UniqueFilter(param ItemRequestUniqueFilterParam) *dbflex.Filter {
	return dbflex.And(
		dbflex.Eq("ItemID", param.ItemID),
		dbflex.Eq("SKU", param.SKU),
		dbflex.Eq("InventoryDimension.WarehouseID", param.InventoryDimension.WarehouseID),
		dbflex.Eq("InventoryDimension.AisleID", param.InventoryDimension.AisleID),
		dbflex.Eq("InventoryDimension.SectionID", param.InventoryDimension.SectionID),
		dbflex.Eq("InventoryDimension.BoxID", param.InventoryDimension.BoxID),
	)
}
