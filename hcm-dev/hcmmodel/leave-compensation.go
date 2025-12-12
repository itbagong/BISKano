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

type LeaveCompensation struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	CompanyID         string `grid:"hide" form:"hide"`
	JournalTypeID     string `grid:"hide" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Leave Compensation|_id|Name"`
	PostingProfileID  string `grid:"hide" form:"hide"`
	RequestorID       string `label:"Requestor name"`
	RequestDate       time.Time
	NumbeOfLeave      int
	ApprovedAmount    float64
	Status            ficomodel.JournalStatus   `form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time                 `form_kind:"datetime" grid:"hide"`
}

func (o *LeaveCompensation) TableName() string {
	return "HCMLeaveCompensations"
}

func (o *LeaveCompensation) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LeaveCompensation) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LeaveCompensation) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LeaveCompensation) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LeaveCompensation) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LeaveCompensation) PostSave(dbflex.IConnection) error {
	return nil
}

type LeaveCompensationForm struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"4"`
	RequestorID       string                    `form_section:"General" form_required:"1" label:"Requestor ID" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	JournalTypeID     string                    `form_section:"General" grid:"hide" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Leave Compensation|_id|Name"`
	PostingProfileID  string                    `form_section:"General" grid:"hide" form:"hide"`
	RequestorName     string                    `form_section:"General2" form_read_only:"1"`
	RequestDate       time.Time                 `form_kind:"date" form_section:"General2"`
	NumbeOfLeave      int                       `form_section:"General3"`
	ApprovedAmount    float64                   `form_section:"General3"`
	Status            string                    `form_section:"General2" form_read_only:"1"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info"`
}

func (o *LeaveCompensationForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}
