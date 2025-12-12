package mfgmodel

import (
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderPlanStatusOverall string

const (
	WorkOrderPlanStatusOverallDraft      WorkOrderPlanStatusOverall = "DRAFT"
	WorkOrderPlanStatusOverallInProgress WorkOrderPlanStatusOverall = "IN PROGRESS" // TODO: UI set Status to this when user click Submit button (in save first before postingprofile/post)
	WorkOrderPlanStatusOverallEnd        WorkOrderPlanStatusOverall = "END"
)

type WorkOrderPlan struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                     `bson:"_id" json:"_id" label:"WO No" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"1" form_section_size:"3"`
	JournalTypeID     string                     `form_section:"General" grid:"hide"  form_required:"1" form_lookup:"/mfg/workorder/journal/type/find|_id|_id,Name"`
	RequestorWOName   string                     `form:"hide" form_section:"General" label:"WO Requestor"  form_lookup:"/tenant/employee/find|_id|_id,Name"`
	WoTypeKind        string                     `form_section:"General" grid:"hide" label:"Wo Type Kind"`
	WOName            string                     `form_section:"General" label:"WO Name" form_required:"1"`
	TrxDate           time.Time                  `form_kind:"date" label:"WO Date" form_required:"1" form_section:"General"` // WO Date
	Priority          string                     `form:"hide" form_section:"General" form_items:"Top|Middle|Low"`
	SafetyInstruction []string                   `grid:"hide" form_section:"General" form_multiple:"1" form_items:"Safety helmet|Safety shoes|Ear muff|Ear plug|Mask|Face shield|Reflector vest|Apron|Gloves"`
	BOM               string                     `form:"hide" grid:"hide" form_section:"General" form_label:"BOM" form_lookup:"/mfg/bom/find|_id|_id,Title" form_section_auto_col:"1"`
	WorkDescription   string                     `grid:"hide" form_multi_row:"3" label:"Description" form_section:"General" form_section_auto_col:"1"`
	Status            ficomodel.JournalStatus    `form_section:"General" form_read_only:"1"`
	StatusOverall     WorkOrderPlanStatusOverall `form:"hide" form_section:"General" form_read_only:"1"`
	BreakdownType     string                     `form:"hide" grid:"hide" form_section:"General" label:"Breakdown Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=WOBreakdownType|_id|_id,Name"`
	TrxCreatedDate    time.Time                  `form:"hide" form_section:"General" form_kind:"datetime-local" form_read_only:"1" grid:"hide"`

	WorkRequestID         string     `form:"hide" form_section:"Work Request Information" label:"Reff WR ID" form_read_only:"1" form_section_auto_col:"2" form_section_show_title:"1"`
	RequestorName         string     `form:"hide" grid:"hide" form_section:"Work Request Information" label:"WR Requestor"  form_lookup:"/tenant/employee/find|_id|_id,Name"`
	RequestorDepartment   string     `form:"hide" grid:"hide" form_section:"Work Request Information" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|_id,Name"`
	Asset                 string     `form_section:"Work Request Information" label:"Police No" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|_id,Name"`
	WRDate                time.Time  `grid:"hide" form_kind:"date" form_section:"Work Request Information" label:"WR Date"`
	Schedule              time.Time  `form:"hide" grid:"hide" form_kind:"date" form_section:"Work Request Information" form_read_only:"1" label:"Schedule"`
	StartDownTime         *time.Time `form:"hide" form_kind:"datetime-local" form_section:"Work Request Information" form_read_only:"1" grid:"hide"`
	ExpectedCompletedDate time.Time  `grid:"hide" form:"hide" form_kind:"date" form_read_only:"1" form_section:"Work Request Information"`
	Kilometers            float64    `grid:"hide" form_section:"Work Request Information"` // TODO: default dari Asset (editable)
	UnitType              string     `form_section:"Work Request Information" grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=VTA|_id|_id,Name"`
	Merk                  string     `form_section:"Work Request Information" grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MUA|_id|_id,Name"`
	CaroseryCode          string     `grid:"hide" form_section:"Work Request Information" form_read_only:"1"`
	Year            	  string     `grid:"hide" form_section:"Work Request Information" label:"Year"`
	ACSystem              string     `grid:"hide" form_section:"Work Request Information" label:"AC System"`
	HullNo                string     `form:"hide" form_section:"Work Request Information" form_read_only:"1" label:"Hull No (Lambung)"`  // TODO: default dari Asset -> Detail -> Hull No. (Read Only)
	NoHullCustomer        string     `form:"hide" form_section:"Work Request Information" form_read_only:"1" label:"No Hull Customer"`   // TODO: default dari Asset -> Detail -> Hull No. (Read Only)
	WRDescription         string     `grid:"hide" form_multi_row:"3" label:"Remarks" form_section:"Work Request Information" form_section_auto_col:"1" form_read_only:"1"`
	SunID                 string     `form:"hide" grid:"hide" label:"Sun No" form_read_only:"1" form_section:"Work Request Information"` // TODO: kalo diisi, ngelakuin apa ya?
	WorkRequestType       string     `form:"hide" grid:"hide"`
	
	Dimension tenantcoremodel.Dimension `form_section:"Dimension"`
	InventDim scmmodel.InventDimension  `grid:"hide" form:"hide" form_section:"Dimension"`

	// new additional fields, please move accordingly
	PostingProfileID string        `form:"hide" grid:"hide"`
	CompanyID        string        `form:"hide" grid:"hide" form_lookup:"/tenant/company/find|_id|_id,Name"`
	TrxType          InventTrxType `form:"hide" form_items:"Work Order" grid:"hide"`
	OutputItem       string        `form:"hide" grid:"hide" form_label:"Output Item"`
	Created          time.Time     `form:"hide" grid:"hide" form_kind:"datetime" form_read_only:"1"`
	LastUpdate       time.Time     `form:"hide" grid:"hide" form_kind:"datetime" form_read_only:"1"`
	IsFirsttimeSave  bool          `form:"hide" grid:"hide"`

	// UI Only
	Summary string `grid:"hide" form:"hide" form_section_auto_col:"1"`

	Text string `form:"hide" grid:"hide"` // biar ga error aja karena ada penambahan baru
}

//	func (o *WorkOrderPlan) FormSections() []suim.FormSectionGroup {
//		return []suim.FormSectionGroup{
//			{Sections: []suim.FormSection{
//				{Title: "GeneralWO", ShowTitle: false, AutoCol: 2},
//				{Title: "GeneralWOBOM", ShowTitle: false, AutoCol: 1},
//				{Title: "GeneralWODesc", ShowTitle: false, AutoCol: 1},
//				// {Title: "Summary", ShowTitle: true, AutoCol: 1},
//			}},
//			{Sections: []suim.FormSection{
//				{Title: "GeneralWR", ShowTitle: false, AutoCol: 2},
//			}},
//			{Sections: []suim.FormSection{
//				{Title: "Dimension", ShowTitle: false, AutoCol: 1},
//			}},
//		}
//	}
func (o *WorkOrderPlan) TableName() string {
	return "WorkOrderPlans"
}

func (o *WorkOrderPlan) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderPlan) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderPlan) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderPlan) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderPlan) PreSave(dbflex.IConnection) error {
	o.formatID()

	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()

	if o.IsFirsttimeSave {
		o.TrxCreatedDate = time.Now()
	}
	return nil
}

func (o *WorkOrderPlan) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *WorkOrderPlan) formatID() {
	/*
		Format
		•⁠  ⁠⁠MO Breakdown Maintenance (1)
		•⁠  ⁠MO Backlog Maintenance (2)
		•⁠  ⁠MO Preventive Maintenance (3)
		•⁠  ⁠⁠MO Produksi (4)
		•⁠  ⁠MO Overhoul Maintenance (5)

		Misal : WO (Kode)(BulanTahun)(Numberauto)
		Jadi WO1

		Misal "WO 103240001" untuk WO breakdown
	*/

	typeM := map[string]string{
		"WO_BDMaintenance":      "1",
		"WO_BLMaintenance":      "2",
		"WO_PrevMaintenance":    "3",
		"WO_Production":         "4",
		"WO_PCR/OVHMaintenance": "5",
		"WOGeneral":             "6",
	}

	jt := "X"
	if v, ok := typeM[o.JournalTypeID]; ok {
		jt = v
	}

	o.ID = strings.Replace(o.ID, "[JTYPE]", jt, -1)
	o.ID = strings.Replace(o.ID, "[JTYPEText]", "", -1)
	o.ID = strings.TrimSpace(o.ID)
}
