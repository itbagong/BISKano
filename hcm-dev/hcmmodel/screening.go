package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Screening struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `form:"hide" bson:"_id" json:"_id"`
	CandidateID       string    `form:"hide"`
	JobVacancyID      string    `form:"hide"`
	Note              string    ` form_multi_row:"5"`
	Status            string    `form:"hide"`
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *Screening) TableName() string {
	return "HCMScreenings"
}

func (o *Screening) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Screening) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Screening) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Screening) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Screening) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Screening) PostSave(dbflex.IConnection) error {
	return nil
}
