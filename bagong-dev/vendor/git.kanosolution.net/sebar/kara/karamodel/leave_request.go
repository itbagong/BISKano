package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveRequestCreate struct {
	LeaveTypeID string    `form_lookup:"/kara/leave/type/find|_id|Name" form_section:"General" form_section_auto_col:"2" form_required:"1"`
	UserID      string    `form_lookup:"/iam/user/find|_id|DisplayName"`
	LeaveFrom   time.Time `form_kind:"date" form_required:"1"`
	LeaveTo     time.Time `form_kind:"date" form_required:"1"`
	Name        string    `form_required:"1" form_section:"General" label:"Title"`
	Approvers   []string  `form_lookup:"/kara/profile/find|UserID|Name" form_required:"1"`
}

type LineApprovers struct {
	UserIDs []string
}

type LineApproval struct {
	UserID    string
	Timestamp time.Time
	Op        string
	Reason    string
}

type LeaveRequest struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string                  `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	LeaveTypeID          string                  `form_lookup:"/kara/leave/type/find|_id|Name"`
	UserID               string                  `form_lookup:"/iam/user/find|_id|DisplayName"`
	Name                 string                  `form_required:"1" form_section:"General" label:"Title"`
	LeaveFrom            time.Time               `form_kind:"date"`
	LeaveTo              time.Time               `form_kind:"date"`
	Status               LeaveApprovalStatusEnum `form_read_only:"1"`
	Reason               string                  `form_read_only:"1"`
	Approvers            []LineApprovers         `grid:"hide" form:"hide" form_read_only:"1"`
	Approvals            []LineApproval          `grid:"hide" form:"hide" form_read_only:"1"`
	CurrentApproverIndex *int                    `grid:"hide" form:"hide"`
	CurrentApprovers     []string                `grid:"hide" form:"hide"`
	Created              time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *LeaveRequest) TableName() string {
	return "KaraLeaveRequests"
}

func (o *LeaveRequest) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LeaveRequest) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *LeaveRequest) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LeaveRequest) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LeaveRequest) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == "" {
		o.Status = "Draft"
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LeaveRequest) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *LeaveRequest) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "LeaveType_User_Index", Fields: []string{"LeaveTypeID", "UserID"}},
	}
}
