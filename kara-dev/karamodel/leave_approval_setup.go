package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveApprovalSetup struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `form:"hide" grid:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	LeaveTypeID       string    `grid:"hide" form:"hide" form_lookup:"/kara/leave/type/find|_id|_id, Name"`
	Site              string    `form_lookup:"/tenant/dimension/find?DimensionType=Site|_id|Label"`
	ApproverIDs       []string  `form_lookup:"/iam/user/find|_id|DisplayName" label:"Approvers"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LeaveApprovalSetup) TableName() string {
	return "LeaveApprovalSetups"
}

func (o *LeaveApprovalSetup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LeaveApprovalSetup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LeaveApprovalSetup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LeaveApprovalSetup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LeaveApprovalSetup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LeaveApprovalSetup) PostSave(dbflex.IConnection) error {
	return nil
}
