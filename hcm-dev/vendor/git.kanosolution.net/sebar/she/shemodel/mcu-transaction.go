package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MCUTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form:"hide"`
	Date              time.Time                 `form_section:"General" form_kind:"date"`
	Category          MCUCategory               `form_section:"General" form_items:"Candidate|Employee"`
	Name              string                    `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name"`
	Gender            string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
	Age               int                       `form_section:"General"`
	Position          string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Customer          string                    `form_section:"General2" form_lookup:"/tenant/customer/find|_id|Name"`
	DoctorParamedic   string                    `form_section:"General2" label:"Doctor/Paramedic"`
	Provider          string                    `label:"Hospital/Provider" form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MPR|_id|Name"`
	Purpose           string                    `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MPU|_id|Name"`
	MCUPackage        string                    `form_section:"General2" form_lookup:"/she/mcumasterpackage/find|_id|PackageName" label:"MCU Package"`
	AdditionalItem    []string                  `form_section:"General2" form_lookup:"/she/mcutransaction/get-mcu-item-last-child|ID|Description" grid:"hide" form_use_list:"1"`
	FollowUp          []MCUFollowUp             `form_section:"General2" grid:"hide" form:"hide"`
	IsCheck           bool                      `form_section:"General" form:"hide" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	AssessmentResult  AssessmentResult          `form_section:"Assessment Result" grid:"hide" form:"hide"`
	MCUResult         MCUResult                 `form_section:"MCU Result" grid:"hide" form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *MCUTransaction) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "General2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Info", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type AssessmentResult struct {
	RegisterNo        string    `form_section:"General"`
	RegisterDate      time.Time `form_kind:"date" form_section:"General"`
	ExaminationDate   time.Time `form_kind:"date" form_section:"General"`
	ResponsiblePerson string    `form_section:"General"`
	UrineCollection   time.Time `form_kind:"date" form_section:"General"`
	BloodCollection   time.Time `form_kind:"date" form_section:"General"`
	Assessment        string    `form_multi_row:"5" form_section:"General2"`
	Therapy           string    `form_multi_row:"5" form_section:"General2"`
	Diagnose          string    `form_multi_row:"5" form_section:"General2"`
	Recomendation     string    `form_multi_row:"5" form_section:"General2"`
}

func (o *AssessmentResult) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "General2", ShowTitle: false, AutoCol: 2},
		}},
	}
}

type MCUResult struct {
	VisitResult       MCUVisitResult         `form_items:"Fit|UnFit" form_section:"General" form_pos:"1"`
	DetailPemeriksaan []MCUMasterPackageLine `form_section:"General" form_pos:"2"`
	Notes             string                 `form_multi_row:"5" form_section:"General" form_pos:"3,1"`
	DocumentMCU       string                 `form_section:"General" form_pos:"3,2" label:"Documents"`
	AdditionalItem    []MCUResultDetailFrom  `form_section:"General" form_pos:"4"`
}

func (o *MCUResult) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type MCUFollowUp struct {
	ID                string                `form:"hide"`
	DetailPemeriksaan []MCUResultDetailFrom `form_pos:"1"`
	Notes             string                `form_multi_row:"5" form_pos:"2,1"`
	Document          string                `form_pos:"2,2"`
	DoctorParamedic   string                `form:"hide"`
	AdditionalItem    []string              `form:"hide"`
}

type MCUResultItem struct {
	DetailPemeriksaan []MCUMasterPackageLine
	Document          string
	Notes             string `form_multi_row:"5" `
}

type MCUResultDetailFrom struct {
	ID           string `form:"hide" form_section_direction:"row"`
	Number       string `form:"hide"`
	Description  string `form_multi_row:"3" form_read_only:"1"`
	Result       string
	Type         MCUTemplateType    `form:"hide"`
	IsGender     bool               `form:"hide"`
	NilaiRujukan string             `form_read_only:"1"`
	Unit         string             `form_read_only:"1"`
	Range        []MCURange         `form:"hide"`
	Condition    []MCULineCondition `form:"hide"`
	Parent       string             `form:"hide"`
	IsSelected   bool               `form:"hide"`
	Note         string             `grid:"hide" form_multi_row:"2"`
}

func (o *MCUResultDetailFrom) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 5},
		}},
	}
}

func (o *MCUTransaction) TableName() string {
	return "SHEMCUTransactions"
}

func (o *MCUTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MCUTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MCUTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MCUTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MCUTransaction) PreSave(dbflex.IConnection) error {
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

func (o *MCUTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
