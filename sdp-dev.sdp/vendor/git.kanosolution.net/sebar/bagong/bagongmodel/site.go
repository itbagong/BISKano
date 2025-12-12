package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteBenefit struct {
	ficomodel.PayrollBenefit
	IsCash bool
}

type SiteDeduction struct {
	ficomodel.PayrollBenefit
	IsCash bool
}

type Shift struct {
	ShiftID   string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SHFT|_id|Name"`
	StartTime string `form_kind:"time"`
	EndTime   string `form_kind:"time"`
}

type Overtime struct {
	Staff           string `form_section:"Overtime" form_items:"Flat|Perhitungan UU" form_section_auto_col:"2"`
	StaffDivider    int    `form_section:"Overtime" form_label:"Divider"`
	Driver          string `form_section:"Overtime" form_items:"Flat|Perhitungan UU"`
	DriverDivider   int    `form_section:"Overtime" form_label:"Divider"`
	Mechanic        string `form_section:"Overtime" form_items:"Flat|Perhitungan UU"`
	MechanicDivider int    `form_section:"Overtime" form_label:"Divider"`
}

type Site struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string
	Address           string
	Alias             string
	Purpose           string                        `form_items:"BTS|Mining|Trayek|Tourism"`
	Shift             []Shift                       `form:"hide"`
	Overtime          Overtime                      `grid:"hide" form:"hide"`
	Benefits          []SiteBenefit                 `grid:"hide" form:"hide"`
	Deductions        []SiteDeduction               `grid:"hide" form:"hide"`
	Expense           []tenantcoremodel.ExpenseType `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension     `grid:"hide"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Site) TableName() string {
	return "BGSites"
}

func (o *Site) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Site) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Site) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Site) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Site) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Site) PostSave(dbflex.IConnection) error {
	return nil
}
