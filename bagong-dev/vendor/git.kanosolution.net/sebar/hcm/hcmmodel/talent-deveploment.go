package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TalentDevelopmentMapping struct {
	Name              string `form_section:"Summary" form_section_direction:"row" form_section_auto_col:"4" form_section_show_title:"1"`
	NIK               string `form_section:"Summary" form_label:"NIK"`
	Department        string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|Name"`
	Group             string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GME|_id|Name"`
	Level             string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LME|_id|Name"`
	POH               string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PME|_id|Name"`
	EmployeeStatus    string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=ESM|_id|Name"`
	Position          string `form_section:"Summary" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Site              string `form_section:"Summary" form_lookup:"/tenant/dimension/find?DimensionType=Site|_id|Label"`
	WorkingExperience string `form_section:"Summary"`
	SPRecord          string `form_label:"SP Record" form_section:"Summary"`
}
type TalentDevelopment struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	CompanyID         string                    `grid:"hide" form:"hide"`
	JournalTypeID     string                    `grid:"hide" form_required:"1"`
	PostingProfileID  string                    `grid:"hide" form:"hide"`
	EmployeeID        string                    `form_section:"General2" form_label:"Employee Name" form_lookup:"/tenant/employee/find|_id|Name" form_required:"1"`
	SubmissionType    string                    `form_items:"Promotion|Rotation|Demotion|Salary Change|POH" form_section:"General3"`
	Reason            string                    `form_section:"General"`
	Assesment         bool                      `form_section:"General2" form_label:"Tracking Assessment"`
	AssesmentResult   string                    `form_section:"General3"`
	Status            ficomodel.JournalStatus   `form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type TalentDevelopmentForm struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	SubmissionType    string                    `form_items:"Promotion|Rotation|Demotion|Salary Change|POH" form_section:"General"`
	JournalTypeID     string                    `grid:"hide" form_required:"1"`
	PostingProfileID  string                    `grid:"hide" form:"hide"`
	Reason            string                    `form_section:"General2"`
	Status            string                    `form_section:"General2" form_read_only:"1"`
	Assesment         bool                      `form_section:"General2" form_label:"Tracking Assessment"`
	AssesmentResult   string                    `form_section:"General"`
	EmployeeID        string                    `form_section:"General3" form_label:"NIK" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	EmployeeName      string                    `form_required:"1" form_section:"General3" label:"Name" form_read_only:"1"`
	PointOfHire       string                    `form_required:"1" form_section:"General3" label:"Point of Hire" form_read_only:"1"`
	JoinedDate        time.Time                 `form_required:"1" form_section:"General3" label:"Joined Date" form_read_only:"1"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TalentDevelopmentForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *TalentDevelopment) TableName() string {
	return "HCMTalentDevelopments"
}

func (o *TalentDevelopment) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TalentDevelopment) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TalentDevelopment) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TalentDevelopment) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TalentDevelopment) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TalentDevelopment) PostSave(dbflex.IConnection) error {
	return nil
}
