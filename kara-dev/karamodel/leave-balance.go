package karamodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LeaveBalance struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `form:"hide" grid:"hide" bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string `form:"hide" grid:"hide" form_lookup:"/iam/user/find|_id|DisplayName,Email"`
	LeaveTypeID       string `form_read_only_edit:"1" form_lookup:"/kara/leave/type/find|_id|Name"`
	Balance           int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LeaveBalance) TableName() string {
	return "LeaveBalances"
}

func (o *LeaveBalance) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LeaveBalance) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LeaveBalance) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LeaveBalance) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LeaveBalance) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	cmd := dbflex.From(o.TableName()).Where(dbflex.Eqs("LeaveTypeID", o.LeaveTypeID, "UserID", o.UserID)).Take(1)
	other := LeaveBalance{}
	conn.Cursor(cmd, nil).Fetch(&other)
	if other.ID != o.ID && other.ID != "" {
		return errors.New("duplicate")
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *LeaveBalance) PostSave(dbflex.IConnection) error {
	return nil
}
