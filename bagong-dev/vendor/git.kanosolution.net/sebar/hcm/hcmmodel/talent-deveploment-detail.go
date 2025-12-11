package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TalentDevelopmentDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id"`
	Department        string    ``
	Position          string    ``
	Grade             string    ``
	Group             string    ``
	SubGroup          string    ``
	Site              string    ``
	PointOfHire       string    ``
	BasicSalary       float64   ``
	Allowance         string    ``
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *TalentDevelopmentDetail) TableName() string {
	return "HCMTalentDevelopmentDetails"
}

func (o *TalentDevelopmentDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TalentDevelopmentDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TalentDevelopmentDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TalentDevelopmentDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TalentDevelopmentDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TalentDevelopmentDetail) PostSave(dbflex.IConnection) error {
	return nil
}
