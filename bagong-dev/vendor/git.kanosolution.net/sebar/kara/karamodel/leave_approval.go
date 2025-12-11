package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveApprovalStatusEnum string

const (
	LeaveRequestIsDraft     LeaveApprovalStatusEnum = "Draft"
	LeaveRequestIsSubmitted LeaveApprovalStatusEnum = "Submitted"
	LeaveRequestIsApproved  LeaveApprovalStatusEnum = "Approved"
	LeaveRequestIsRejected  LeaveApprovalStatusEnum = "Rejected"
)

type LeaveApproval struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	LeaveRequestID    string
	ApproverID        string
	Status            LeaveApprovalStatusEnum
	TrxDate           time.Time
	Reason            string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LeaveApproval) TableName() string {
	return "KaraLeaveApprovals"
}

func (o *LeaveApproval) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LeaveApproval) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LeaveApproval) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LeaveApproval) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LeaveApproval) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LeaveApproval) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *LeaveApproval) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "LeaveRequest_Approver_Index", Fields: []string{"LeaveRequestID", "ApproverID"}},
	}
}
