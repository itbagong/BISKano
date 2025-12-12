package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuestionBankQuestionType struct {
	Name     string
	Duration int
}

type QuestionBank struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	TestName          string
	Module            string
	Menu              string
	QuestionType      []QuestionBankQuestionType
	Status            bool
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *QuestionBank) TableName() string {
	return "HCMQuestionBanks"
}

func (o *QuestionBank) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *QuestionBank) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *QuestionBank) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *QuestionBank) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *QuestionBank) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *QuestionBank) PostSave(dbflex.IConnection) error {
	return nil
}
