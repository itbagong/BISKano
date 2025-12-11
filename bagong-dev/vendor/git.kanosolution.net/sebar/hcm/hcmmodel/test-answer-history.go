package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MultipleChoiceAnswerHistory struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	QuestionID        string
	EmployeeID        string
	Type              string
	TestTypeID        string
	Answer            int
	Score             int
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *MultipleChoiceAnswerHistory) TableName() string {
	return "HCMTestAnswerHistorys"
}

func (o *MultipleChoiceAnswerHistory) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MultipleChoiceAnswerHistory) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MultipleChoiceAnswerHistory) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MultipleChoiceAnswerHistory) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MultipleChoiceAnswerHistory) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MultipleChoiceAnswerHistory) PostSave(dbflex.IConnection) error {
	return nil
}
