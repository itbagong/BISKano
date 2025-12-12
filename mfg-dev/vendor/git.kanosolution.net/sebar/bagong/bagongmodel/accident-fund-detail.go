package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccidentFundDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form:"hide" grid:"hide"`
	AccidentFundID    string    `grid:"hide" form:"hide" `
	Date              time.Time `form_kind:"date"`
	Mutation          float64
	Notes             string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide"`
}

func (o *AccidentFundDetail) TableName() string {
	return "BGAccidentFundDetails"
}

func (o *AccidentFundDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AccidentFundDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AccidentFundDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AccidentFundDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AccidentFundDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AccidentFundDetail) PostSave(dbflex.IConnection) error {
	return nil
}
