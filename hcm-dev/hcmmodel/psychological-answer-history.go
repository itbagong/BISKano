package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PsychologicalAnswerHistory struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id"`
	TestID             string // psychological test id or talent development id
	TemplateTestID     string
	QuestionID         string
	AnswerID           string
	Score              int
	IsMostQuestionType bool
	Answer             string
	Created            time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate         time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *PsychologicalAnswerHistory) TableName() string {
	return "HCMPsychologicalAnswerHistories"
}

func (o *PsychologicalAnswerHistory) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PsychologicalAnswerHistory) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PsychologicalAnswerHistory) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PsychologicalAnswerHistory) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PsychologicalAnswerHistory) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PsychologicalAnswerHistory) PostSave(dbflex.IConnection) error {
	return nil
}
