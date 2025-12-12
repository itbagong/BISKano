package flowmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RequestStatus string

const (
	RequestDraft   RequestStatus = "DRAFT"
	RequestRunning RequestStatus = "RUNNING"
	RequestSuccess RequestStatus = "SUCCESS"
	RequestFail    RequestStatus = "FAIL"
	RequestCancel  RequestStatus = "CANCEL"
)

type Request struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	Reason            string
	Template          *FlowTemplate
	Data              codekit.M
	Version           int
	Status            RequestStatus
	CompanyID         string
	CreatedBy         string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Request) TableName() string {
	return "Requests"
}

func (o *Request) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Request) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Request) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Request) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Request) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Request) PostSave(dbflex.IConnection) error {
	return nil
}
