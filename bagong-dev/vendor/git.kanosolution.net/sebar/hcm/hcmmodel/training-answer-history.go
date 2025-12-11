package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingAnswerHistory struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	TrainingCenterID  string
	ParticipantID     string
	TemplateTestID    string
	QuestionID        string
	AnswerID          string
	Score             int
	Answer            string
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *TrainingAnswerHistory) TableName() string {
	return "HCMTrainingAnswerHistories"
}

func (o *TrainingAnswerHistory) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingAnswerHistory) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingAnswerHistory) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingAnswerHistory) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingAnswerHistory) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingAnswerHistory) PostSave(dbflex.IConnection) error {
	return nil
}
