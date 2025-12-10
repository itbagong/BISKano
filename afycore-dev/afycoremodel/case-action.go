package afycoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InpatientOutpatient string

const (
	InPatient  InpatientOutpatient = "In-Patient"
	OutPatient InpatientOutpatient = "Out-Patient"
)

type CaseActionStatus string
type BillingStatus string

const (
	ActionScheduled        CaseActionStatus = "Scheduled"
	ActionExecuted         CaseActionStatus = "Executed"
	ActionWaitingForResult CaseActionStatus = "Waiting for result"
	ActionDone             CaseActionStatus = "Done"
	ActionCancelled        CaseActionStatus = "Cancelled"
)

const (
	BillingDraft     BillingStatus = "Draft"
	BillingPrepared  BillingStatus = "Prepared"
	BillingPaid      BillingStatus = "Paid"
	BillingCancelled BillingStatus = "Cancelled"
	BillingNone      BillingStatus = "None"
)

type CaseAction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CaseID            string
	PatientID         string
	QueNo             int
	ActionDate        time.Time
	LocationPoliID    string
	ShiftID           string
	MainDoctorID      string
	RecordType        InpatientOutpatient
	ActionID          string
	ActionStatus      CaseActionStatus
	BillingStatus     BillingStatus
	BillingID         string
	Data              codekit.M
	Notes             string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CaseAction) TableName() string {
	return "CaseAction"
}

func (o *CaseAction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CaseAction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CaseAction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CaseAction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CaseAction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CaseAction) PostSave(dbflex.IConnection) error {
	return nil
}
