package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InterviewDetail struct {
	EmployeeID string
	Note       string
}

type Interview struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" form:"hide"`
	CandidateID       string    `form:"hide"`
	JobVacancyID      string    `form:"hide"`
	Date              time.Time `form_label:"Interview Date" form_kind:"date"`
	Notes             []InterviewDetail
	Status            string    `form:"hide"`
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *Interview) TableName() string {
	return "HCMInterviews"
}

func (o *Interview) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Interview) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Interview) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Interview) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Interview) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Interview) PostSave(dbflex.IConnection) error {
	return nil
}
