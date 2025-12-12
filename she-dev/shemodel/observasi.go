package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Observasi struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_space_after:"1"`
	Observer          string                    `form_lookup:"/tenant/employee/find|_id|Name"`
	Observee          string                    `form_lookup:"/tenant/customer/find|_id|Name"`
	Name              string                    `form_multi_row:"3"`
	TemplateID        string                    `form_section:"General" form_lookup:"/she/mcuitemtemplate/find|_id|Name"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Lines             []ObservasiLine           `form_section:"Line" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *Observasi) FormSections() []suim.FormSectionGroup {
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

type ObservasiLine struct {
	TemplateLine MCUItemTemplateLine
	IsApplicable bool
	Deviation    string
	HazardCode   string
	HarzardLevel string
	Result       bool
	Remark       string
	Attachment   string
}

func (o *Observasi) TableName() string {
	return "SHEObservasi"
}

func (o *Observasi) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Observasi) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Observasi) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Observasi) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Observasi) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == "" {
		o.Status = string(SHEStatusDraft)
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Observasi) PostSave(dbflex.IConnection) error {
	return nil
}
