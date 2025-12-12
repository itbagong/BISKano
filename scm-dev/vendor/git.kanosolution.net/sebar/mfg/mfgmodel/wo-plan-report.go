package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WorkOrderPlanReport doesn't use posting profile
type WorkOrderPlanReport struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"3"`
	WorkOrderPlanID   string `form_section:"General" grid:"hide" label:"Work Order ID" form_read_only:"1"`
	DailyStatus       string `form_section:"General"`

	WorkDate         time.Time `form_kind:"date" label:"Work Date" form_read_only:"1" form_section:"General2"`
	MonitoringStatus string    `form_section:"General2"`
	
	Status WorkOrderPlanStatusOverall `form_read_only:"1" form_section:"General3"` // DRAFT / IN PROGRESS / END
	ComponentCategory string    `form_section:"General3"`

	WorkOrderPlanReportConsumptionID     string `grid:"hide" form:"hide"` // ada isinya -> butuh posting
	WorkOrderPlanReportResourceID        string `grid:"hide" form:"hide"` // ada isinya -> butuh posting
	WorkOrderPlanReportOutputID          string `grid:"hide" form:"hide"` // ada isinya -> butuh posting
	WorkOrderPlanReportConsumptionPPID   string `grid:"hide" form:"hide"` // posting profile id
	WorkOrderPlanReportResourcePPID      string `grid:"hide" form:"hide"` // posting profile id
	WorkOrderPlanReportOutputPPID        string `grid:"hide" form:"hide"` // posting profile id
	WorkOrderPlanReportConsumptionStatus string `grid:"hide" form:"hide"` // status posting profile <- dynamically filled in mw /mfg/workorderplan/report/get or mfglogic/wo-plan-report.go/Get()
	WorkOrderPlanReportResourceStatus    string `grid:"hide" form:"hide"` // status posting profile <- dynamically filled in mw /mfg/workorderplan/report/get or mfglogic/wo-plan-report.go/Get()
	WorkOrderPlanReportOutputStatus      string `grid:"hide" form:"hide"` // status posting profile <- dynamically filled in mw /mfg/workorderplan/report/get or mfglogic/wo-plan-report.go/Get()

	HasRequested bool `grid:"hide" form:"hide"`

	Created    time.Time `form:"hide" grid:"hide" form_kind:"datetime" form_read_only:"1" form_section:"General3"`
	LastUpdate time.Time `form:"hide" grid:"hide" form_kind:"datetime" form_read_only:"1" form_section:"General3"`

	// UI Only
	Consumption string `grid:"hide" form_section:"Consumption" form_section_direction:"row"`
	Resource    string `grid:"hide" form_section:"Resource" form_section_direction:"row"`
	Output      string `grid:"hide" form_section:"Output" form_section_direction:"row"`
}

func (o *WorkOrderPlanReport) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Consumption", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Resource", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Output", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *WorkOrderPlanReport) TableName() string {
	return "WorkOrderPlanReports"
}

func (o *WorkOrderPlanReport) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderPlanReport) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderPlanReport) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderPlanReport) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderPlanReport) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderPlanReport) PostSave(dbflex.IConnection) error {
	return nil
}
