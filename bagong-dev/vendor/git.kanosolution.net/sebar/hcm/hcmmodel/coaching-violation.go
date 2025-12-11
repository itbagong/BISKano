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

type CoachingViolation struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	RequestorID       string `label:"Requestor name" form_lookup:"/tenant/employee/find|_id|Name"`
	CompanyID         string `grid:"hide" form:"hide"`
	JournalTypeID     string `grid:"hide" form_required:"1"`
	PostingProfileID  string `grid:"hide" form:"hide"`
	RequestDate       time.Time
	StartDate         time.Time
	EndDate           time.Time
	Type              string
	EmployeeID        string `label:"Employee name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	Violation         string
	Investigation     string
	Status            ficomodel.JournalStatus   `form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time                 `form_kind:"datetime" grid:"hide"`
}

func (o *CoachingViolation) TableName() string {
	return "CoachingViolations"
}

func (o *CoachingViolation) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CoachingViolation) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CoachingViolation) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CoachingViolation) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CoachingViolation) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CoachingViolation) PostSave(dbflex.IConnection) error {
	return nil
}

type CoachingViolationForm struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3"`
	CompanyID         string                    `grid:"hide" form:"hide"`
	Type              string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CoachingViolation|_id|Name"`
	JournalTypeID     string                    `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID  string                    `grid:"hide" form:"hide"`
	RequestorID       string                    `form_section:"General" form_required:"1" label:"Requestor name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	RequestDate       time.Time                 `form_section:"General" form_kind:"date"`
	StartDate         time.Time                 `form_section:"General2" form_kind:"date"`
	EndDate           time.Time                 `form_section:"General2" form_kind:"date"`
	EmployeeID        string                    `form_required:"1" form_section:"General3" label:"Employee Name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	EmployeeNIK       string                    `form_section:"General3" label:"NIK" form_read_only:"1"`
	Position          string                    `form_section:"General3" label:"Position" form_read_only:"1"`
	Department        string                    `form_section:"General3" label:"Department" form_read_only:"1"`
	Violation         string                    `form_section:"General4" form_multi_row:"5"`
	Investigation     string                    `form_section:"General4" form_multi_row:"5"`
	Status            string                    `form_section:"General2" form_read_only:"1"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info"`
}

func (o *CoachingViolationForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "General4", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}
