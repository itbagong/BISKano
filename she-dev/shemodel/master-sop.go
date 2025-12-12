package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MasterSOP struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string                     `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General"`
	CreatedDate          time.Time                  `form_kind:"date" form_section:"General" grid:"hide"`
	TitleOfDocument      string                     `form_section:"General" grid_pos:"2"`
	PICDocument          string                     `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name" grid_pos:"3" label:"PIC Document"`
	NatureOfChange       SOPNatureOfChange          `grid:"hide" form_required:"1" form_section:"General" form_items:"Pembuatan|Revisi|Obsolete"`
	DocumentNo           string                     `form_section:"General" grid:"hide"`
	PICFacilitator       string                     `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name" grid_pos:"4" label:"PIC Facilitator"`
	DocumentType         SOPDocumentType            `gird:"hide" form_section:"General" form_items:"SOP|Manual|STD|INK|Form" grid:"hide"`
	LatestRevisionStatus string                     `form_section:"General" grid:"hide"`
	CompletionDate       time.Time                  `form_kind:"date" form_section:"General" grid:"hide"`
	DocumentRefno        string                     `form_section:"General" grid:"hide" form_lookup:"/she/mastersopsummary/find|_id|TitleOfDocument"`
	JobPosition          []string                   `form_section:"General" grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	EffectiveDate        time.Time                  `form_kind:"date" form_section:"General" grid_pos:"1"`
	PointsOfChange       string                     `form_multi_row:"5" form_section:"General2" grid_pos:"5"`
	Reasons              string                     `form_multi_row:"5" form_section:"General2" grid_pos:"6"`
	Attachment           tenantcoremodel.Attachment `form_section:"General2" grid_pos:"8"`
	ApprovedBy           string                     `form:"hide" grid_pos:"7"`
	Dimension            tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status               string                     `form_read_only:"1"  form_section:"Info"`
}

func (o *MasterSOP) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "General2", ShowTitle: false, AutoCol: 2},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *MasterSOP) TableName() string {
	return "SHEMasterSOPs"
}

func (o *MasterSOP) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MasterSOP) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MasterSOP) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MasterSOP) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MasterSOP) PreSave(dbflex.IConnection) error {
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

func (o *MasterSOP) PostSave(dbflex.IConnection) error {
	return nil
}
