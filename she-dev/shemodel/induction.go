package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// form_section_direction:"row"
type Induction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                       `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_label:"No." form_section_auto_col:"3" `
	InductionDate     time.Time                    `form_section:"General" form_kind:"date" label:"Date"`
	Name              string                       `form_section:"General" label:"Induction Name"`
	JournalTypeID     string                       `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                       `form:"hide" grid:"hide"`
	Company           string                       `grid:"hide" form:"hide" form_lookup:"/tenant/company/find|_id|Name"`
	Category          string                       `label:"Category" form_section:"General" form_items:"New Employee|Mutasi/Rotasi|Pulang Cuti"`
	Type              string                       `label:"Type" form_section:"General" form_items:"Short Induction|Long Induction"`
	Pc                string                       `grid:"hide" form:"hide"`
	CC                string                       `grid:"hide" form:"hide"`
	Site              string                       `grid:"hide" form:"hide"`
	Asset             string                       `grid:"hide" form:"hide"`
	AssesmentStatus   bool                         `grid:"hide" form:"hide"`
	Attendee          []InductionAttendee          `grid:"hide" form:"hide"`
	Materials         []InductionMaterial          `grid:"hide" form:"hide"`
	Assesment         []InductionAssessment        `grid:"hide" form:"hide"`
	Attachments       []tenantcoremodel.Attachment `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension    `form_section:"Dimension"`
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                       `form_section:"Info" form_read_only:"1"`
}

func (o *Induction) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type InductionAttendee struct {
	ID             string `grid:"hide" bson:"_id" json:"_id" `
	NoLine         int    `label:"No" form_read_only:"1"`
	EmploymentType bool   // Internal: true, External: false
	Name           string
	Position       string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Presence       string `form_items:"Present|Not Present"`
}

type InductionMaterial struct {
	ID        string `grid:"hide" bson:"_id" json:"_id" `
	NoLine    int    `label:"No" form_read_only:"1"`
	Category  string `form_items:"SOP|Manual|STD|INK|Form"`
	Material  string
	Document  string `form_read_only:"1"`
	Reference string
}

type InductionAssessment struct {
	ID         string `grid:"hide" bson:"_id" json:"_id" `
	NoLine     int    `label:"No" form_read_only:"1"`
	Attendee   string
	Result     string
	Attachment string `form_read_only:"1"`
	Feedback   string
}

func (o *Induction) TableName() string {
	return "SHEInduction"
}

func (o *Induction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Induction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Induction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Induction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Induction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Status == "" {
		o.Status = string(SHEStatusDraft)
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Induction) PostSave(dbflex.IConnection) error {
	return nil
}
