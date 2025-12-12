package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Jsa struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `form_section:"General" bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section_size:"4" form_label:"No."`
	PositionInvolved  string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	EquipmentInvolved []string                  `form_section:"General" form_multiple:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=Equipment|_id|Name"`
	JsaDate           time.Time                 `form_section:"General" form_kind:"date" label:"Date"`
	Customer          string                    `form_section:"General" form_lookup:"/tenant/customer/find|_id|Name"`
	Apd               []string                  `form_section:"General" label:"APD" form_multiple:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=APD|_id|Name"`
	JournalTypeID          string `grid:"hide" form_section:"General"  form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID       string `form:"hide" grid:"hide" form_section:"General"`
	Type              string                    `form_section:"General" form_items:"New|Revision"`
	Location          string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail    string                    `form_section:"General" grid:"hide" form_multi_row:"5"`
	TemplateID        string                    `form_section:"General" form:"hide" grid:"hide" form_lookup:"/she/mcuitemtemplate/find?SHE-0010|_id|Name"`
	Lines             []JsaLine                 `form_section:"Line" grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_section:"Info" form_read_only:"1"`
}

func (o *Jsa) FormSections() []suim.FormSectionGroup {
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

type JsaLine struct {
	ID                 string              `grid:"hide"`
	Parent             bool                `grid:"hide" form:"hide"`
	LineNo             int                 `label:"No" form_read_only:"1"`
	StepsTask          string              `label:"Steps Task"`
	HazardAndRiskSteps string              `label:"Hazard And Risk Steps"`
	TemplateLine       MCUItemTemplateLine `grid:"hide" form:"hide"`
	IsApplicable       bool                `grid:"hide" form:"hide"`
	Metode             []string            `grid:"hide" form:"hide"`
	Bobot              int                 `grid:"hide" form:"hide"`
	HazardCode         string              `label:"Hazard Code" form_items:"AA|A|B"`
	Recommendation     string              `label:"Recommendation For Control" form_multi_row:"3"`
	Remark             string              `grid:"hide" form:"hide"`
	Attachment         string              `grid:"hide" form:"hide"`
}

func (o *Jsa) TableName() string {
	return "SHEjsa"
}

func (o *Jsa) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Jsa) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Jsa) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Jsa) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Jsa) PreSave(dbflex.IConnection) error {
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

func (o *Jsa) PostSave(dbflex.IConnection) error {
	return nil
}
