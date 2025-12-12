package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentPracticeDuration struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	Date              time.Time
	ParticipantID     string
	TrainingCenterID  string
	P2H               bool
	PoliceNo          string
	ActivityName      []string
	StartTime         time.Time
	EndTime           time.Time
	Duration          int
	Note              string
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *TrainingDevelopmentPracticeDuration) TableName() string {
	return "HCMTrainingDevelopmentPracticeDurations"
}

func (o *TrainingDevelopmentPracticeDuration) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentPracticeDuration) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentPracticeDuration) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentPracticeDuration) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentPracticeDuration) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentPracticeDuration) PostSave(dbflex.IConnection) error {
	return nil
}
