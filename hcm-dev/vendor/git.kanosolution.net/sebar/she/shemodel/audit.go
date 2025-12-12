package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Audit struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_label:"No."`
	Category          AuditCategory             `form_section:"General" form_items:"SMK3|SMKPAU|SMKP"`
	Name              string                    `form_section:"General"`
	TemplateID        string                    `form_section:"General" form_lookup:"/she/mcuitemtemplate/find?Menu=SHE-0013|_id|Name" grid:"hide"`
	Auditor           string                    `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name" form_space_after:"1"`
	SMK3              []SMK3                    `form_section:"Line" grid:"hide"`
	SMKP              []SMKP                    `form_section:"Line" grid:"hide"`
	SMKPAU            []SMKPAU                  `form_section:"Line" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status             string                       `form_section:"Info" form_read_only:"1"`
}

func (o *Audit) FormSections() []suim.FormSectionGroup {
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

type SMK3 struct {
	TemplateLine MCUItemTemplateLine `grid:"hide"`
	No           string
	ElementAudit string
	Point        int
	Score        int
	ACH          float64
	Pica         *Pica
	UsePica      bool `grid:"hide"`
}

type SMKP struct {
	TemplateLine      MCUItemTemplateLine `grid:"hide"`
	No                string
	ElementAudit      string
	ElementValue      float64
	MaxElementValue   int
	AuditElementValue float64
	ScoreAudit        int
	Note              string
}

type SMKPAU struct {
	TemplateLine       MCUItemTemplateLine `grid:"hide"`
	No                 string
	ElementAudit       string
	SupportingDocument string
	AttachedDocument   string
	Point              bool
	Target             int
	ActualValue        int
	ACH                float64
	Note               string
}

func (o *Audit) TableName() string {
	return "SHEAudits"
}

func (o *Audit) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Audit) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Audit) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Audit) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Audit) PreSave(dbflex.IConnection) error {
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

func (o *Audit) PostSave(dbflex.IConnection) error {
	return nil
}
