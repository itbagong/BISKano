package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pica struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                       `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" label:"PICA Number" form_section_auto_col:"2" form_section_direction:"row"`
	SourceModule       STModule                     `grid:"hide" form_read_only:"1" form_section:"General"`
	DueDate            time.Time                    `form_kind:"date" form_section:"General"`
	SourceNumber       string                       `form_read_only:"1" form_section:"General2"`
	EmployeeID         string                       `form_lookup:"/tenant/employee/find|_id|Name" form_section:"General2" label:"Responsible Person (PIC)" form_required:"1" form_section_auto_col:"2"`
	FindingDescription string                       `form_multi_row:"5"  form_section:"General2"`
	Dimension          tenantcoremodel.Dimension    `form_section:"Dimension"`
	ActionDate         time.Time                    `form_kind:"date" form_section:"TakeAction" form_section_show_title:"1"`
	Evidence           []tenantcoremodel.Attachment `form_section:"TakeAction" form_section_auto_col:"2"`
	Comment            string                       `form_multi_row:"5" form_section:"TakeAction"`
	JournalTypeID      string                       `form:"hide" grid:"hide"`
	PostingProfileID   string                       `form:"hide" grid:"hide"`
	Created            time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate         time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy          string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy       string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status             string                       `form_section:"Info"`
}

type ItemPica struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                       `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" label:"PICA Number" form_section_auto_col:"2" form_section_direction:"row"`
	SourceModule       STModule                     `grid:"hide" form_read_only:"1" form_section:"General"`
	DueDate            time.Time                    `form_kind:"date" form_section:"General"`
	SourceNumber       string                       `form_read_only:"1" form_section:"General2"`
	EmployeeID         string                       `form_lookup:"/tenant/employee/find|_id|Name" form_section:"General2" label:"Responsible Person (PIC)" form_required:"1" form_section_auto_col:"2"`
	FindingDescription string                       `form_multi_row:"5"  form_section:"General2"`
	Dimension          tenantcoremodel.Dimension    `form_section:"Dimension"`
	ActionDate         time.Time                    `form_kind:"date" form_section:"TakeAction" form_section_show_title:"1"`
	Evidence           []tenantcoremodel.Attachment `form_section:"TakeAction" form_section_auto_col:"2"`
	Comment            string                       `form_multi_row:"5" form_section:"TakeAction"`
	Created            time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate         time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy          string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy       string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status             string                       `form_section:"Info"`
}

func (o *Pica) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "TakeAction", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *ItemPica) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
		}},
	}
}
func (o *Pica) TableName() string {
	return "SHEPicas"
}

func (o *Pica) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Pica) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Pica) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Pica) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Pica) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Pica) PostSave(dbflex.IConnection) error {
	return nil
}
