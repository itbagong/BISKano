package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitCalendarForm for Form of add / edit form calendar
type UnitCalendarForm struct {
	SORefNo   string `form_section:"Project" form_label:"SORef No." form_section_size:"3" form_lookup:"/sdp/salesorder/find|_id|SalesOrderNo,Name" form_read_only_edit:"1"`
	ProjectID string `form_section:"Project" form_label:"Project ID" form_lookup:"/sdp/measuringproject/find|_id|ProjectName" form_read_only_edit:"1"`

	CustomerName string `form_section:"Customer" form_label:"Customer" form_lookup:"/tenant/customer/find|_id|Name" `
	Remark       string `form_section:"Customer" form_read_only_edit:"1"`

	Dimension tenantcoremodel.Dimension `form_section:"Dimension"`
}

// UnitCalendarLineGrid Form grid for line unit calendar
type UnitCalendarLineGrid struct {
	AssetUnitID []string  `form_lookup:"/tenant/asset/find|_id|Name"`
	Duration    uint32    `form_kind:"number"`
	Uom         string    `form_lookup:"/tenant/unit/find|_id|Name" form_allow_add:"1"`
	StartDate   time.Time `form_kind:"date" form_label:"Start Date"`
	EndDate     time.Time `form_kind:"date" form_label:"End Date" form_read_only:"1"`
	Qty         uint32    `form_kind:"number" form_read_only:"1"`
}

// UnitCalendar model for Unit Calendar
type UnitCalendar struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	SORefNo           string

	Site string

	Lines []struct {
		Index        uint32
		AssetUnitID  string
		IsItem       bool
		StartDate    time.Time
		EndDate      time.Time
		Uom          string
		Duration     uint32
		Qty          uint32
		Descriptions string
	}

	Customer  string
	ProjectID string
	Remark    string

	Dimension tenantcoremodel.Dimension

	// InventoryDimension InventoryDimension `form_section:"Specification"`
	Created    time.Time
	LastUpdate time.Time
}

func (o *UnitCalendar) TableName() string {
	return "UnitCalendar"
}

func (o *UnitCalendar) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *UnitCalendar) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *UnitCalendar) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UnitCalendar) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UnitCalendar) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *UnitCalendar) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *UnitCalendar) PreDelete(dbflex.IConnection) error {
	// if o.Status != UnitCalendarDraft {
	// 	return errors.New("protected record, could not be deleted")
	// }

	return nil
}
