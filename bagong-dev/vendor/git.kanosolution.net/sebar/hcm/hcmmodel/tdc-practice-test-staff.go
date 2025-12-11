package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentPracticeTestStaffDetail struct {
	AssementName  string
	MaxScore      float64
	AchievedScore float64
	Note          string
}

type TrainingDevelopmentPracticeTestStaff struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	ParticipantID     string
	TemplateID        string
	Details           []TrainingDevelopmentPracticeTestStaffDetail
	TotalScore        float64
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *TrainingDevelopmentPracticeTestStaff) TableName() string {
	return "HCMTrainingDevelopmentPracticeTestStaffs"
}

func (o *TrainingDevelopmentPracticeTestStaff) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentPracticeTestStaff) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentPracticeTestStaff) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentPracticeTestStaff) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentPracticeTestStaff) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentPracticeTestStaff) PostSave(dbflex.IConnection) error {
	return nil
}
