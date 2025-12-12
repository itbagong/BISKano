package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PsychologicalTestDetailStatus string

const (
	PsychologicalTestDetailStatusOpen PsychologicalTestDetailStatus = "OPEN"
	PsychologicalTestDetailStatusDone PsychologicalTestDetailStatus = "DONE"
)

type PsychologicalTestDetail struct {
	TemplateID   string
	TemplateName string
	Detail       interface{}
	Status       PsychologicalTestDetailStatus
}

type PsychologicalTest struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" form:"hide"`
	CandidateID       string `form:"hide"`
	JobVacancyID      string `form:"hide"`
	Details           []PsychologicalTestDetail
	IsTestSent        bool      `form:"hide"`
	Note              string    `form_multi_row:"5"`
	Status            string    `form:"hide"`
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *PsychologicalTest) TableName() string {
	return "HCMPsychologicalTests"
}

func (o *PsychologicalTest) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PsychologicalTest) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PsychologicalTest) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PsychologicalTest) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PsychologicalTest) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PsychologicalTest) PostSave(dbflex.IConnection) error {
	return nil
}
