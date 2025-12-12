package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Inspection struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_label:"No."`
	Inspectors        string                    `form_read_only:"1" form_lookup:"/tenant/employee/find|_id|Name"`
	Name              string                    `form_section:"General"`
	LocationID        string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name" grid_label:"Location" form_required:"1" form_section:"General" label:"Location"`
	LocationDetail    string                    `grid:"hide" form_multi_row:"5" form_section:"General"`
	TemplateID        string                    `form_section:"General" form_lookup:"/she/mcuitemtemplate/find?Menu=SHE-0011|_id|Name" grid:"hide"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Lines             []InspectionLine          `form_section:"Line" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *Inspection) FormSections() []suim.FormSectionGroup {
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

type InspectionLine struct {
	TemplateLine MCUItemTemplateLine
	IsApplicable bool
	Value        int
	Pica         *Pica
	UsePica      bool
	Remark       string
	Attachment   string
}

func (o *Inspection) TableName() string {
	return "SHEInspections"
}

func (o *Inspection) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Inspection) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Inspection) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Inspection) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Inspection) PreSave(dbflex.IConnection) error {
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

func (o *Inspection) PostSave(dbflex.IConnection) error {
	return nil
}
