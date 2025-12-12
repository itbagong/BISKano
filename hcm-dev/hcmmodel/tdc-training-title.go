package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentTitle struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `form:"hide" bson:"_id" json:"_id"`
	Alias             string `form_read_only_edit:"1"`
	Name              string
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *TrainingDevelopmentTitle) TableName() string {
	return "HCMTrainingDevelopmentTitles"
}

func (o *TrainingDevelopmentTitle) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentTitle) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentTitle) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentTitle) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentTitle) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentTitle) PostSave(dbflex.IConnection) error {
	return nil
}
