package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OnBoarding struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" form:"hide"`
	CandidateID       string `form:"hide"`
	JobVacancyID      string `form:"hide"`
	Checklist         tenantcoremodel.Checklists
	Status            string    `form:"hide"`
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *OnBoarding) TableName() string {
	return "HCMOnBoardings"
}

func (o *OnBoarding) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *OnBoarding) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *OnBoarding) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *OnBoarding) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *OnBoarding) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *OnBoarding) PostSave(dbflex.IConnection) error {
	return nil
}
