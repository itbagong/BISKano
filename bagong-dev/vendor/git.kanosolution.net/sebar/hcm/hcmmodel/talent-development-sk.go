package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TalentDevelopmentSKType string

const (
	TalentDevelopmentSKTypeActing    TalentDevelopmentSKType = "ACTING"
	TalentDevelopmentSKTypePermanent TalentDevelopmentSKType = "PERMANENT"
)

type TalentDevelopmentSK struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string `bson:"_id" json:"_id" form:"hide" form_section:"General" form_section_auto_col:"2"`
	TalentDevelopmentID   string `form:"hide" form_section:"General"`
	CompanyID             string `grid:"hide" form:"hide" form_section:"General"`
	JournalTypeID         string `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID      string `grid:"hide" form:"hide" form_section:"General"`
	Notices               []string
	DirectorID            string
	Decides               []string
	EffectivePromotedDate time.Time               `form_kind:"date"`
	Type                  TalentDevelopmentSKType `form:"hide" form_section:"General"`
	Status                ficomodel.JournalStatus `form_section:"General" form_read_only:"1"`
	Created               time.Time               `grid:"hide" form:"hide"`
	LastUpdate            time.Time               `grid:"hide" form:"hide"`
}

type TalentDevelopmentSKActingForm struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string                  `bson:"_id" json:"_id" form_section:"General" form_section_auto_col:"2" form_read_only:"1"`
	TalentDevelopmentID   string                  `form:"hide" form_section:"General"`
	CompanyID             string                  `grid:"hide" form:"hide" form_section:"General"`
	JournalTypeID         string                  `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID      string                  `grid:"hide" form:"hide" form_section:"General"`
	DirectorID            string                  `form_section:"General2" form_lookup:"/tenant/employee/find|_id|Name"`
	DirectorPosition      string                  `form_section:"General2" form_read_only:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	DirectorAddress       string                  `form_section:"General2" form_read_only:"1"`
	Notices               []string                `form_section:"General2"`
	Decides               []string                `form_section:"General2"`
	EffectivePromotedDate time.Time               `form_section:"General" form_kind:"date" form:"hide"`
	Type                  TalentDevelopmentSKType `form:"hide" form_section:"General"`
	Status                ficomodel.JournalStatus `form_section:"General" form_read_only:"1"`
	Name                  string                  `form_read_only:"1" form_section:"Employee" form_section_direction:"row" form_section_auto_col:"4" form_section_show_title:"1"`
	Position              string                  `form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Department            string                  `form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|Name"`
	Grade                 string                  `form_read_only:"1" form_section:"Employee" label:"Grade" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GDE|_id|Name"`
	Level                 string                  `form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LME|_id|Name"`
	Group                 string                  `form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GME|_id|Name"`
	NIK                   string                  `form_read_only:"1" form_section:"Employee" form_label:"NIK"`
	IdentityCardNo        string                  `form_read_only:"1" form_section:"Employee" label:"Identity Card No"`
	PlaceOfBirth          string                  `form_read_only:"1" form_section:"Employee" label:"Place of Birth"`
	DateOfBirth           time.Time               `form_read_only:"1" form_section:"Employee" label:"Date of Birth" form_kind:"date"`
	Gender                string                  `form_read_only:"1" form_section:"Employee" label:"Gender" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
	Religion              string                  `form_read_only:"1" form_section:"Employee" label:"Religion" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RME|_id|Name"`
	Phone                 string                  `form_read_only:"1" form_section:"Employee" label:"Phone Number"`
	Address               string                  `form_read_only:"1" form_section:"Employee" label:"Address"`
	Created               time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate            time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *TalentDevelopmentSKActingForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "Employee", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type TalentDevelopmentSKPermanentForm struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string                  `bson:"_id" json:"_id" form_section:"General" form_section_auto_col:"2" form_read_only:"1"`
	TalentDevelopmentID   string                  `form:"hide" form_section:"General"`
	CompanyID             string                  `grid:"hide" form:"hide" form_section:"General"`
	JournalTypeID         string                  `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID      string                  `grid:"hide" form:"hide" form_section:"General"`
	DirectorID            string                  `form_section:"General2" form:"hide"`
	Notices               []string                `form_section:"General2" form:"hide"`
	Decides               []string                `form_section:"General2" form:"hide"`
	EffectivePromotedDate time.Time               `form_section:"General2" form_kind:"date"`
	Type                  TalentDevelopmentSKType `form:"hide" form_section:"General"`
	Status                ficomodel.JournalStatus `form_section:"General2" form_read_only:"1"`
	Name                  string                  `form_read_only:"1" form_section:"Employee" form_section_direction:"row" form_section_auto_col:"4" form_section_show_title:"1"`
	Position              string                  `form:"hide" form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Department            string                  `form:"hide" form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|Name"`
	Grade                 string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Grade" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GDE|_id|Name"`
	Level                 string                  `form:"hide" form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LME|_id|Name"`
	Group                 string                  `form:"hide" form_read_only:"1" form_section:"Employee" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GME|_id|Name"`
	NIK                   string                  `form_read_only:"1" form_section:"Employee" form_label:"NIK"`
	IdentityCardNo        string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Identity Card No"`
	PlaceOfBirth          string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Place of Birth"`
	DateOfBirth           time.Time               `form:"hide" form_read_only:"1" form_section:"Employee" label:"Date of Birth" form_kind:"date"`
	Gender                string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Gender" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
	Religion              string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Religion" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RME|_id|Name"`
	Phone                 string                  `form:"hide" form_read_only:"1" form_section:"Employee" label:"Phone Number"`
	Address               string                  `form:"hide" form_read_only:"1" form_section:"" label:"Address"`
	POH                   string                  `form_read_only:"1" form_section:"Employee" label:"POH"`
	JoinedDate            time.Time               `form_read_only:"1" form_section:"Employee" form_kind:"date"`
	Detail                string                  `form_read_only:"1" form_section:"Detail"`
	Created               time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate            time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *TalentDevelopmentSKPermanentForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "Employee", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Detail", ShowTitle: false, AutoCol: 1},
		}},
	}
}
func (o *TalentDevelopmentSK) TableName() string {
	return "HCMTalentDevelopmentSKs"
}

func (o *TalentDevelopmentSK) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TalentDevelopmentSK) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TalentDevelopmentSK) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TalentDevelopmentSK) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TalentDevelopmentSK) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TalentDevelopmentSK) PostSave(dbflex.IConnection) error {
	return nil
}
