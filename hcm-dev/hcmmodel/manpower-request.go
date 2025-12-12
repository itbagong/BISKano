package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ManpowerStage string

const (
	ManpowerStageScreening     ManpowerStage = "SCREENING"
	ManpowerStagePsychotest    ManpowerStage = "PSYCHOTEST"
	ManpowerStageInterview     ManpowerStage = "INTERVIEW"
	ManpowerStageTechInterview ManpowerStage = "TECHINTERVVIEW"
	ManpowerStageMCU           ManpowerStage = "MCU"
	ManpowerStageTraining      ManpowerStage = "TRAINING"
	ManpowerStageOLPlotting    ManpowerStage = "OLPLOTTING"
	ManpowerStagePKWT          ManpowerStage = "PKWT"
	ManpowerStageOnboarding    ManpowerStage = "ONBOARDING"
)

type ManpowerRequest struct {
	orm.DataModelBase        `bson:"-" json:"-"`
	ID                       string                    `bson:"_id" json:"_id" form_read_only:"1" form_section:"General" form_section_direction:"row" form_section_show_title:"1" `
	CompanyID                string                    `grid:"hide" form:"hide"`
	JournalTypeID            string                    `grid:"hide" form_required:"1" form_section:"General" form_lookup:"/hcm/journaltype/find?TransactionType=Manpower Request|_id|Name"`
	PostingProfileID         string                    `grid:"hide" form:"hide"`
	RequestDate              time.Time                 `form_kind:"datetime"  form_read_only:"1" form_section:"General" `
	RequestorID              string                    `form_lookup:"/tenant/employee/find|_id|Name" form_section:"General"`
	RequestType              string                    `form_items:"MPP|Non-MPP" form_section:"General"  grid:"hide"`
	Name                     string                    `form_section:"General2"`
	EmployementType          string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=EmploymentType|_id|Name" form_section:"General2"  grid:"hide"`
	OnsiteRequiredDate       time.Time                 `form_kind:"date"  form_section:"General2"`
	EmployeeSource           string                    `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=EmployeeSource|_id|Name"  grid:"hide"`
	JobVacancyTitle          string                    `form_section:"General3" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Notes                    string                    `form_multi_row:"4" form_section:"General3"  grid:"hide"`
	Status                   ficomodel.JournalStatus   `form:"hide" form_section:"Dimension"`
	IsClose                  bool                      `form:"hide" grid_label:"Job status"`
	Dimension                tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" `
	Position                 string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name" form_section:"Request Reason"  grid:"hide"`
	Class                    string                    `form_section:"Replacement"  grid:"hide"`
	ReplacedEmployeeName     string                    `form_lookup:"/tenant/employee/find|_id|Name" form_section:"Replacement"  grid:"hide"`
	ReplacementSeason        string                    `form_section:"Replacement"  grid:"hide"`
	EmployeeNumberTotal      int                       `form_section:"Replacement" form_read_only:"1"  grid:"hide"`
	AdditionalNumber         int                       `form_section:"Additional"  grid:"hide"`
	ClosedEmployeeTotal      int                       `grid:"hide" form:"hide"` // total remaining employee when close
	ExistingEmployeeNumber   int                       `form_section:"Additional"  grid:"hide"`
	ReasonAdditionalEmployee string                    `form_multi_row:"5" form_section:"Additional"  grid:"hide"`
	EstimateCostPerMonth     float64                   `form_section:"Additional"  grid:"hide"`
	Stage                    ManpowerStage             `grid:"hide" form:"hide" form_items:"SCREENING|PSYCHOTEST|INTERVIEW|TECHINTERVVIEW|MCU|TRAINING|OLPLOTTING|PKWT|ONBOARDING"`
	Created                  time.Time                 `form_kind:"datetime"  form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate               time.Time                 `form_kind:"datetime"   form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ManpowerRequest) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Request Reason", ShowTitle: true, AutoCol: 1},
		}},
		// {Sections: []suim.FormSection{
		// 	{Title: "Replacement", ShowTitle: true, AutoCol: 4},
		// }},
		// {Sections: []suim.FormSection{
		// 	{Title: "Additional", ShowTitle: true, AutoCol: 4},
		// }},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *ManpowerRequest) TableName() string {
	return "HCMManpowerRequests"
}

func (o *ManpowerRequest) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ManpowerRequest) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ManpowerRequest) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ManpowerRequest) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ManpowerRequest) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ManpowerRequest) PostSave(dbflex.IConnection) error {
	return nil
}
