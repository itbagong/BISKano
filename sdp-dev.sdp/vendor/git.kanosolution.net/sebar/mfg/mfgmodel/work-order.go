package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderSource string
type WOStatus string

var (
	FromWorkOrder      WorkOrderSource = "Work Order"
	FromWorkRequest    WorkOrderSource = "Work Request"
	WOStatusDraft      WOStatus        = "DRAFT" // for WO Only, NO Daily Report
	WOStatusInProgress WOStatus        = "IN PROGRESS"
	WOStatusCompleted  WOStatus        = "COMPLETED"
)

type WorkOrder struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                            `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General"  form_section_auto_col:"2"`
	WOName            string                            `form_required:"1" Label:"Name" form_section:"General"`
	WorkRequestID     string                            `form_section:"General"`
	CompanyID         string                            `form_section:"General" form_lookup:"/tenant/company/find|_id|_id,Name" form:"hide"`
	WOType            string                            `grid:"hide" form_section:"General" form_section_direction:"row" label:"WO Type" form_items:"PRODUKSI|PREV MAINTENANCE|BREAKDOWN MAINTENANCE|BACKLOG MAINTENANCE|GENERAL MAINTENANCE|PCR/OCH MAINTENANCE"`
	JournalTypeID     string                            `form_section:"General" form_lookup:"/mfg/workorder/journal/type/find|_id|_id,Name"`
	PostingProfileID  string                            `form_section:"General" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ItemUsage         []scmmodel.InventReceiveIssueLine `grid:"hide" form:"hide"`
	//TODO: ini dirubah ke type data untuk manpower usage dan machine usage
	TrxDate       *time.Time                `form_kind:"date" form_section:"General2"`
	TrxType       scmmodel.InventTrxType    `form_section:"General2" grid:"hide" form:"hide"`
	ManpowerUsage []ManPower                `grid:"hide" form:"hide"`
	MachineUsage  []Machinery               `grid:"hide" form:"hide"`
	Source        WorkOrderSource           `form_section:"General2" form_read_only:"1"`
	SourceType    string                    `form_section:"General2" form_read_only:"1"`
	SourceID      string                    `form_section:"General2"`
	Status        WOStatus                  `form_section:"General2" form_read_only:"1"`
	InventDim     scmmodel.InventDimension  `grid:"hide" form_section:"InventDim"`
	Dimension     tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created       time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate    time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	// Detail
	Name                string     `form_required:"1" label:"Requestor" form_section:"Detail1" form_section_direction:"row" form_lookup:"/tenant/employee/find|_id|Name"`
	RequestorDepartment string     `grid:"hide" form_section:"Detail1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|_id,Name"`
	UnitID              string     `grid:"hide" form:"hide" form_section:"Detail1"`
	StartDownTime       *time.Time `grid:"hide" form_kind:"datetime-local" form_section:"Detail1"`
	TargetFinishTime    *time.Time `grid:"hide" form_kind:"datetime-local" form_section:"Detail1"`

	EquipmentNo string  `grid:"hide" form_required:"1" form_section:"Detail2" form_lookup:"/tenant/asset/find|_id|_id,Name"`
	Kilometers  float64 `grid:"hide" form_required:"1" form_section:"Detail2"`
	AssignedPIC string  `grid:"hide" form_required:"1" form_section:"Detail2" label:"Assigned PIC" form_lookup:"/bagong/employee/gets-filter?Department=DME002|_id|_id,Name"`

	SunID          string     `grid:"hide" form_section:"Detail3"`
	ProductionCode string     `grid:"hide" form_section:"Detail3"`
	PlanStartTime  *time.Time `grid:"hide" form:"hide" form_kind:"datetime-local" form_section:"Detail3"`
	PlanFinishTime *time.Time `grid:"hide" form:"hide" form_kind:"datetime-local" form_section:"Detail3"`

	SafetyInstruction []string          `grid:"hide" form_section:"Detail3" form_multiple:"1" form_items:"Safety helmet|Safety shoes|Ear muff|Ear plug|Mask|Face shield|Reflector vest|Apron|Gloves"`
	FinishDownTime    *time.Time        `grid:"hide" form_kind:"datetime-local" form_section:"Detail4"`
	Description       string            `grid:"hide" form:"hide" form_section:"Detail4" form_multi_row:"3"`
	WorkDescriptions  []WorkDescription `grid:"hide" form_section:"WorkDescriptions"  label:"Work Descriptions" form_section_auto_col:"1"`
}

type WorkOrderGrid struct {
	ID               string                            `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General"  form_section_auto_col:"2"`
	WOName           string                            `form_required:"1" Label:"Name" form_section:"General"`
	WorkRequestID    string                            `form_section:"General"`
	CompanyID        string                            `form_section:"General" form:"hide"`
	WOType           string                            `grid:"hide" form_section:"General" form_section_direction:"row" label:"WO Type" form_items:"PRODUKSI|PREV MAINTENANCE|BREAKDOWN MAINTENANCE|BACKLOG MAINTENANCE|GENERAL MAINTENANCE|PCR/OCH MAINTENANCE"`
	JournalTypeID    string                            `form_section:"General"`
	PostingProfileID string                            `form_section:"General"`
	ItemUsage        []scmmodel.InventReceiveIssueLine `grid:"hide" form:"hide"`
	//TODO: ini dirubah ke type data untuk manpower usage dan machine usage
	TrxDate       *time.Time                `form_kind:"date" form_section:"General2"`
	TrxType       scmmodel.InventTrxType    `form_section:"General2" grid:"hide" form:"hide"`
	ManpowerUsage []ManPower                `grid:"hide" form:"hide"`
	MachineUsage  []Machinery               `grid:"hide" form:"hide"`
	Source        WorkOrderSource           `form_section:"General2" form_read_only:"1"`
	SourceType    string                    `form_section:"General2" form_read_only:"1"`
	SourceID      string                    `form_section:"General2"`
	Status        WOStatus                  `form_section:"General2" form_read_only:"1"`
	InventDim     scmmodel.InventDimension  `grid:"hide" form_section:"InventDim"`
	Dimension     tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created       time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate    time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	// Detail
	Name                string     `form_required:"1" label:"Requestor" form_section:"Detail1" form_section_direction:"row"`
	RequestorDepartment string     `grid:"hide" form_section:"Detail1"`
	UnitID              string     `grid:"hide" form:"hide" form_section:"Detail1"`
	StartDownTime       *time.Time `grid:"hide" form_kind:"datetime-local" form_section:"Detail1"`
	TargetFinishTime    *time.Time `grid:"hide" form_kind:"datetime-local" form_section:"Detail1"`

	EquipmentNo string  `grid:"hide" form_required:"1" form_section:"Detail2"`
	Kilometers  float64 `grid:"hide" form_required:"1" form_section:"Detail2"`
	AssignedPIC string  `grid:"hide" form_required:"1" form_section:"Detail2" label:"Assigned PIC"`

	SunID          string     `grid:"hide" form_section:"Detail3"`
	ProductionCode string     `grid:"hide" form_section:"Detail3"`
	PlanStartTime  *time.Time `grid:"hide" form:"hide" form_kind:"datetime-local" form_section:"Detail3"`
	PlanFinishTime *time.Time `grid:"hide" form:"hide" form_kind:"datetime-local" form_section:"Detail3"`

	SafetyInstruction []string          `grid:"hide" form_section:"Detail3" form_multiple:"1" form_items:"Safety helmet|Safety shoes|Ear muff|Ear plug|Mask|Face shield|Reflector vest|Apron|Gloves"`
	FinishDownTime    *time.Time        `grid:"hide" form_kind:"datetime-local" form_section:"Detail4"`
	Description       string            `grid:"hide" form:"hide" form_section:"Detail4" form_multi_row:"3"`
	WorkDescriptions  []WorkDescription `grid:"hide" form_section:"WorkDescriptions"  label:"Work Descriptions" form_section_auto_col:"1"`
}

func (o *WorkOrder) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "InventDim", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Detail1", ShowTitle: false, AutoCol: 1},
			{Title: "Detail2", ShowTitle: false, AutoCol: 1},
			{Title: "Detail3", ShowTitle: false, AutoCol: 1},
			{Title: "Detail4", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "WorkDescriptions", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type WorkDescription struct {
	WorkDescriptionNo int                   `form:"hide" form_section_auto_col:"2"`
	WorkDescription   string                `form_multi_row:"2" form_section:"General"`
	Type              string                `form_section:"General" form_section_direction:"row"`
	BomName           string                `form_section:"General2"`
	Qty               int                   `form:"hide" form_section:"General2"`
	Note              string                `form:"hide" form_section:"General2"`
	ItemUsage         []WorkDescriptionItem `grid:"hide" form_section:"ItemUsage" form_section_auto_col:"1"`
	AdditionalItem    []WorkDescriptionItem `grid:"hide" form_section:"AdditionalItem" form_section_auto_col:"1"`
	ManpowerUsage     []ManPower            `grid:"hide" form_section:"ManpowerUsage" form_section_auto_col:"1"`
	MachineUsage      []Machinery           `grid:"hide" form:"hide" form_section:"MachineUsage" form_section_auto_col:"1"`
}

type WorkDescriptionItem struct {
	scmmodel.InventReceiveIssueLine
	Requested      bool
	AvailableStock string // filled in MW when get
}

func (o *WorkDescription) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "ItemUsage", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "AdditionalItem", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "ManpowerUsage", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "MachineUsage", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *WorkOrder) TableName() string {
	return "WorkOrders"
}

func (o *WorkOrder) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrder) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrder) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrder) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrder) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrder) PostSave(dbflex.IConnection) error {
	return nil
}
