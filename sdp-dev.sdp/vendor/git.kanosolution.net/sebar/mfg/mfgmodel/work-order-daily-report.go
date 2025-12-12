package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderDailyResume struct {
	ID              string     `grid:"hide"`
	WorkDescription string     `form_section:"General2"`
	WorkDate        *time.Time `form_kind:"date" form_section:"General3"`
	Consumption     int
	Manpower        int
	Output          int
	Status          WOStatus
}

type WorkOrderDailyReport struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form:"hide"  form_section_auto_col:"2"`
	WorkOrderJournalID string          `form_section:"General" form_section_direction:"row"`
	WorkDescriptionNo  int             `form_section:"General2"  form:"hide"`
	WorkDescription    string          `form_section:"General"`
	WorkDate           *time.Time      `form_kind:"date" form_section:"General2"`
	ItemUsage          []ItemUsage     `grid:"hide" form_section:"ItemUsage"`
	AdditionalItem     []ItemUsage     `grid:"hide" form_section:"AdditionalItem"`
	ManpowerUsage      []ManPowerUsage `grid:"hide" form_section:"ManpowerUsage"`
	Output             []Output        `grid:"hide" form_section:"Output"`
	Status             WOStatus        `form_section:"General2"` // "IN PROGRESS" / "COMPLETED"
	ComponentCategory  string          `form_section:"General2" form:"hide"`
	Created            time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate         time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *WorkOrderDailyReport) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "ItemUsage", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "AdditionalItem", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "ManpowerUsage", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Output", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type ItemUsage struct {
	scmmodel.InventReceiveIssueLine
	Requested      bool
	AvailableStock string // filled in MW when get
	Total          float64
}
type ManPowerUsage struct {
	ManPower
	Employee         string
	ActualStartTime  time.Time `form_kind:"time"`
	ActualFinishTime time.Time `form_kind:"time"`
	WorkingHour      float64
	Total            float64
	Description           string `form_multi_row:"3"`
	Status           string
}

type Output struct {
	Output   string
	ItemID   string
	SKU      string
	Type     string
	Qty      int
	UOM      string
	UnitCost int
}

func (o *WorkOrderDailyReport) TableName() string {
	return "WorkOrderDailyReports"
}

func (o *WorkOrderDailyReport) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderDailyReport) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderDailyReport) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderDailyReport) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderDailyReport) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderDailyReport) PostSave(dbflex.IConnection) error {
	return nil
}
