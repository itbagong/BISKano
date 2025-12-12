package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MCUItemTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1"`
	TrxDate           time.Time                 `form_section:"General" form_kind:"date"`
	Name              string                    `form_section:"General"`
	Module            string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MDL|_id|_id,Name" grid:"hide"`
	Menu              string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MENU|_id|_id,Name" grid:"hide"`
	IsActive          bool                      `label:"Status" form_section:"General"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Lines             []MCUItemTemplateLine     `form_section:"Detail" grid:"hide" form:"hide"`
	Instruction       Instructions              `form:"hide" grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *MCUItemTemplate) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type MCUItemTemplateLine struct {
	ID                     string             `form:"hide"`
	Number                 string             `form_section:"General"`
	Type                   MCUTemplateType    `form_section:"General" form_items:"List|Range|String"`
	Unit                   string             `form_section:"General"`
	Condition              []MCULineCondition `form_section:"General"`
	Description            string             `form_section:"General" form_multi_row:"3"`
	IsGender               bool               `form_section:"General"`
	Range                  []MCURange         `form_section:"General"`
	Parent                 string             `grid:"hide" form:"hide"`
	IsSelected             bool               `grid:"hide" form:"hide"`
	AssessmentTypeIsNumber bool               `grid:"hide" form:"hide"`
	QuestionTypeIsMost     bool               `grid:"hide" form:"hide"`
	AnswerValue            int                `grid:"hide"`
	Attachment             string             `form:"hide"`
	Level                  int                `form:"hide"`
	Result                 string             `grid:"hide" form_section:"Result"`
	Note                   string             `grid:"hide" form_multi_row:"3" form_section:"Result"`
}

func (o *MCUItemTemplateLine) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 4},
		}},
	}
}

type MCULineCondition struct {
	ID      string
	Name    string
	Value   bool
	Vnumber int
	Letter  string
}

type MCURange struct {
	Name string
	Min  float64
	Max  float64
}

type Instructions struct {
	CalculationMethod string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=TCM|_id|Name"`
	NoOfQuestions     string
	TotalDuration     int    `label:"Total Duration in Minutes"`
	Instruction       string `form_multi_row:"5"`
	Attachment        string `form_space_after:"1"`
}

func (o *Instructions) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
		}},
	}
}

func (o *MCUItemTemplate) TableName() string {
	return "SHEMCUItemTemplates"
}

func (o *MCUItemTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MCUItemTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MCUItemTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MCUItemTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MCUItemTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MCUItemTemplate) PostSave(dbflex.IConnection) error {
	return nil
}
