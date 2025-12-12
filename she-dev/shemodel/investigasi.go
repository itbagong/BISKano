package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type Investigasi struct {
// 	orm.DataModelBase  `bson:"-" json:"-"`
// 	ID                 string                       `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_label:"No."`
// 	InvestigasiDate    time.Time                    `form_kind:"date" label:"Date"`
// 	AccidentDate       time.Time                    `form_kind:"date" label:"Date"`
// 	Classification     string                       `form_lookup:"/tenant/accidentclassification/find|_id|Name"`
// 	Shift              string                       `form_lookup:"/tenant/shift/find|_id|Name"`
// 	Likehood           string                       `form:"hide"`
// 	Severity           string                       `form:"hide"`
// 	ReportingDate      time.Time                    `form_kind:"date" label:"Date"`
// 	Location           string                       `form_lookup:"/tenant/location/find|_id|Name"`
// 	LocationDetail     string                       `form:"hide"`
// 	Chronology         string                       `form:"hide"`
// 	AccidentAttachment []tenantcoremodel.Attachment `grid:"hide" form_section:"General1"`
// 	AccidentType       []AccidentTypeLine           `form_section:"AccidentType" grid:"hide"`
// 	Person             []PersonLine                 `form_section:"Person" grid:"hide"`
// 	MedicalTreatment   []MedicalTreatmentLine       `form_section:"MedicalTreatment" grid:"hide"`
// 	Asset              []AssetLine                  `form_section:"Asset" grid:"hide"`
// 	Environment        []EnvironmentLine            `form_section:"Environment" grid:"hide"`
// 	DirectCause        []DirectCauseLine            `form_section:"DirectCause" grid:"hide"`
// 	BasicCause         []BasicCauseLine             `form_section:"BasicCause" grid:"hide"`
// 	InvestigationTeam  []TeamMemberLine             `form_section:"InvestigationTeam" grid:"hide"`
// 	RiskReduction      []RiskReductionLine          `form_section:"RiskReduction" grid:"hide"`
// 	CheckListDirection []CheckListDirectionLine     `form_section:"CheckListDirection" grid:"hide"`
// 	ExternalReport     []ExternalReportLine         `form_section:"ExternalReport" grid:"hide"`
// 	TindakLanjut       []TindakLanjutLine           `form_section:"TindakLanjut" grid:"hide"`

// 	Attachments []tenantcoremodel.Attachment `grid:"hide" form_section:"General1"`
// 	Status      string                       `form:"hide"`
// 	Dimension   tenantcoremodel.Dimension    `grid:"hide" form_section:"Dimension"`
// 	Created     time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
// 	LastUpdate  time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
// }

type Investigasi struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_label:"Ref No." form_section:"General"  form_section_auto_col:"3" form_section_direction:"row"`
	AccidentDate      time.Time                 `form_section:"General" form_kind:"datetime-local" label:"Date"`
	Classification    string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=AccidentClassification|_id|Name"`
	Shift             string                    `form_section:"General" grid:"hide"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Likehood          string                    `form_section:"General" grid:"hide"`
	Severity          string                    `form_section:"General" grid:"hide"  form:"hide"`
	ReportingDate     time.Time                 `form_section:"General1" grid:"hide" form_kind:"datetime-local" label:"Reporting Date"`
	Location          string                    `form_section:"General1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail    string                    `form_section:"General1" form_multi_row:"5"`
	Level             string                    `form_section:"General" label:"Level"  form:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension"`

	DetailsAccident   `grid:"hide" form:"hide"`
	Involvement       `grid:"hide" form:"hide"`
	DirectCause       []DirectCauseLine `form_section:"DirectCause" grid:"hide" form:"hide"`
	BasicCause        []BasicCauseLine  `form_section:"BasicCause" grid:"hide" form:"hide"`
	InvestigationTeam `form_section:"InvestigationTeam" grid:"hide" form:"hide"`
	RiskReduction     []RiskReductionLine          `form_section:"RiskReduction" grid:"hide" form:"hide"`
	ExternalReport    []ExternalReportLine         `form_section:"ExternalReport" grid:"hide" form:"hide"`
	PICA              []PICA                       `form_section:"PICA" label:"PICA" grid:"hide" form:"hide"`
	Attachments       []tenantcoremodel.Attachment `form_section:"Attachments" grid:"hide" form:"hide" `
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                       `form_section:"Info" form_read_only:"1"`
}

type DetailsAccident struct {
	Chronology         string                       `form_section:"DetailsAccident1" grid:"hide" form_multi_row:"5"  form_section_auto_col:"3" form_section_direction:"row"`
	AccidentAttachment []tenantcoremodel.Attachment `form_section:"DetailsAccident2" grid:"hide"`
	AccidentType       []AccidentTypeLine           `form_section:"DetailsAccident3" grid:"hide"`
}

type Involvement struct {
	Person           []PersonLine           `form_section:"Person" grid:"hide" form_section_direction:"row"`
	MedicalTreatment []MedicalTreatmentLine `form_section:"Medical Treatment" grid:"hide"`
	Asset            []AssetLine            `form_section:"Asset" grid:"hide"`
	Environment      []EnvironmentLine      `form_section:"Environment" grid:"hide"`
}
type InvestigationTeam struct {
	TeamMember         []TeamMemberLine         `form_section:"Team Member" grid:"hide" form_section_direction:"row"`
	CheckListDirection []CheckListDirectionLine `form_section:"CheckList Direction" grid:"hide"`
}

func (o *Investigasi) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *DetailsAccident) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Chronology", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "DetailsAccident1", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Attachment", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "DetailsAccident2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Accident Type", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "DetailsAccident3", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *Involvement) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Person", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Medical Treatment", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Asset", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Environment", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *InvestigationTeam) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Team Member", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "CheckList Direction", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type AccidentTypeLine struct {
	ID           string `grid:"hide" bson:"_id" json:"_id"`
	LineNo       int    `label:"No" form_read_only:"1"`
	Type         string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=AccidentType|_id|Name"`
	Explaination string
}
type PersonLine struct {
	ID             string `grid:"hide" bson:"_id" json:"_id"`
	LineNo         int    `label:"No" form_read_only:"1"`
	Role           string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RoleAccident|_id|Name"`
	Employee       string `form_lookup:"/tenant/employee/find|_id|Name"`
	Company        string `form_lookup:"/tenant/company/find|_id|Name"`
	Job            string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Gender         string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
	Supervisor     string `form_lookup:"/tenant/employee/find|_id|Name"`
	WorkingPeriod  string
	Age            string
	Injured        []InjuredLine
	EstimationCost float32
}

type InjuredLine struct {
	ID       string `grid:"hide" bson:"_id" json:"_id"`
	LineNo   int    `label:"No" form_read_only:"1"`
	BodyPart string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=BodyPart|_id|Name"`
	Side     string `form_items:"Front|Back|Left|Right|Top|Bottom|Inside|Outside"`
	Remark   string `form_multi_row:"3"`
}

type MedicalTreatmentLine struct {
	ID             string `grid:"hide" bson:"_id" json:"_id"`
	LineNo         int    `label:"No" form_read_only:"1"`
	Doctor         string
	Hospital       string
	Treatment      string
	EstimationCost float32
	Remark         string `form_multi_row:"3"`
}
type AssetLine struct {
	ID            string `grid:"hide" bson:"_id" json:"_id"`
	LineNo        int    `label:"No" form_read_only:"1"`
	Asset         string `form_lookup:"/tenant/asset/find|_id|Name"`
	Unit          string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=AUT|_id|Name"`
	Type          string
	Damage        string
	PartEquipment []PartEquipment
	TotalCost     float32
	Remark        string `form_multi_row:"3"`
}
type PartEquipment struct {
	ID                         string `grid:"hide" bson:"_id" json:"_id"`
	LineNo                     int    `label:"No" form_read_only:"1"`
	scmmodel.InventJournalLine `grid:"hide"`
	CompanyAsset               bool
	Estimation                 string `label:"Estimation Cost of Reparation"`
	Remark                     string `form_multi_row:"3"`
}
type EnvironmentLine struct {
	ID          string `grid:"hide" bson:"_id" json:"_id"`
	LineNo      int    `label:"No" form_read_only:"1"`
	Description string
	Damage      string
	Remark      string `form_multi_row:"3"`
}
type DirectCauseLine struct {
	ID          string `grid:"hide" bson:"_id" json:"_id"`
	LineNo      int    `label:"No" form_read_only:"1"`
	DCType      string `label:"Direct Cause Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DirectCause|_id|Name"`
	DCDetail    string `label:"Direct Cause Detail"  form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DirectCauseDetail|_id|Name"`
	SubDCDetail string `label:"Sub-Direct Cause Detail"`
	Description string `label:"Description" form_multi_row:"3"`
}
type BasicCauseLine struct {
	ID          string `grid:"hide" bson:"_id" json:"_id"`
	LineNo      int    `label:"No" form_read_only:"1"`
	BCType      string `label:"Basic Cause Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=BasicCauseType|_id|Name"`
	BCDetail    string `label:"Cause Detail" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=BasicCauseDetail|_id|Name"`
	SubBCDetail string `label:"Sub-Cause Detail"`
	Description string `label:"Description" form_multi_row:"3"`
}
type RiskReductionLine struct {
	ID                     string `grid:"hide" bson:"_id" json:"_id"`
	Parent                 bool   `grid:"hide" form:"hide"`
	LineNo                 int    `grid:"hide" form:"hide" label:"No" form_read_only:"1"`
	SourceId               string `grid:"hide" form:"hide"`
	IdentifiedCause        string `label:"Identified Cause"`
	SubParent              bool   `grid:"hide" form:"hide"`
	ControlNo              string `grid:"hide" form:"hide"`
	KontrolPengendalian    string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RiskReduction|_id|Name"`
	SubKontrolPengendalian string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SubRiskReduction|_id|Name"`
	Remark                 string `form_multi_row:"3"`
}
type TeamMemberLine struct {
	ID        string `grid:"hide" bson:"_id" json:"_id"`
	LineNo    int    `label:"No" form_read_only:"1"`
	Role      string `label:"Role" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=InvestigationTeam|_id|Name"`
	Name      string `form_lookup:"/tenant/employee/find|_id|Name"`
	JobTittle string `label:"Job Tittle" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
}
type CheckListDirectionLine struct {
	ID        string `grid:"hide" bson:"_id" json:"_id"`
	LineNo    int    `label:"No" form_read_only:"1"`
	CheckList string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=ChecklistInvestigation|_id|Name"`
	Remark    string `form_multi_row:"3"`
}
type ExternalReportLine struct {
	ID            string    `grid:"hide" bson:"_id" json:"_id"`
	LineNo        int       `label:"No" form_read_only:"1"`
	Date          time.Time `form_kind:"date"`
	Reporter      string    `form_lookup:"/tenant/employee/find|_id|Name"`
	ThirdParty    string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=ThirdParty|_id|Name"`
	NeedReport    bool
	PICThirdParty string
	Remark        string `form_multi_row:"3"`
}
type PICA struct {
	ID      string `grid:"hide" bson:"_id" json:"_id"`
	Cause   string
	Action  string
	PIC     string    `label:"PIC" form_lookup:"/tenant/employee/find|_id|Name"`
	DueDate time.Time `form_kind:"date"`
	Status  string
}

func (o *Investigasi) TableName() string {
	return "SHEInvestigasi"
}

func (o *Investigasi) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Investigasi) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Investigasi) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Investigasi) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Investigasi) PreSave(dbflex.IConnection) error {
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

func (o *Investigasi) PostSave(dbflex.IConnection) error {
	return nil
}
