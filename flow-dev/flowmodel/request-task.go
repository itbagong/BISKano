package flowmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskStatus string

const (
	TaskSuccess TaskStatus = "SUCCESS"
	TaskFail    TaskStatus = "FAIL"
	TaskRunning TaskStatus = "RUNNING"
	TaskCancel  TaskStatus = "CANCELLED"
)

type UserApproval struct {
	UserID   string
	Approval bool
	Reason   string
	Time     time.Time
}

type RequestTask struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	RequestID         string
	Version           int
	Setup             Task
	Approval          []UserApproval
	Error             string
	Text              string
	Status            TaskStatus
	Output            string
	Assigned          time.Time
	Completed         time.Time
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RequestTask) TableName() string {
	return "RequestTasks"
}

func (o *RequestTask) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RequestTask) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RequestTask) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RequestTask) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RequestTask) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RequestTask) PostSave(dbflex.IConnection) error {
	return nil
}
