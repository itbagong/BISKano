package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TechnicalInterviewDetail struct {
	Subject     string
	Weight      int
	Description string
	ScoreMlc    float64
	AverageMlc  float64
}

type TechnicalInterviewSection struct {
	Section string
	Weight  int
	Score   float64
	Detail  []TechnicalInterviewDetail
}

type TechnicalInterview struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                      `bson:"_id" json:"_id" form:"hide" form_section:"General" form_section_auto_col:"3" form_section_direction:"row"`
	CandidateID       string                      `form:"hide"  form_section:"General"`
	JobVacancyID      string                      `form:"hide"  form_section:"General"`
	Date              time.Time                   `form_kind:"date" form_label:"Interview Date"  form_section:"General"`
	TemplateID        string                      `form_label:"Template"  form_lookup:"/she/mcuitemtemplate/find|_id|Name"  form_section:"General2"`
	FinalScore        float64                     `form_section:"General3" form_read_only:"1"`
	Grade             string                      `form_section:"General4" form_read_only:"1"`
	Detail            []TechnicalInterviewSection `form_section:"Detail"`
	Status            string                      `form:"hide"`
	Dimension         tenantcoremodel.Dimension   `grid:"hide" form:"hide"`
	Created           time.Time                   `form_kind:"datetime" form:"hide" grid:"hide"`
	LastUpdate        time.Time                   `form_kind:"datetime" form:"hide" grid:"hide"`
}

func (o *TechnicalInterview) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "General4", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Detail", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *TechnicalInterview) TableName() string {
	return "HCMTechnicalInterviews"
}

func (o *TechnicalInterview) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TechnicalInterview) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TechnicalInterview) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TechnicalInterview) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TechnicalInterview) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TechnicalInterview) PostSave(dbflex.IConnection) error {
	return nil
}
