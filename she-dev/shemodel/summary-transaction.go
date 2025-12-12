package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SummaryTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form:"hide"`
	Module            STModule
	RefID             string
	CreatedBy         string
	Created           time.Time
}

func (o *SummaryTransaction) TableName() string {
	return "SHESummaryTransactions"
}

func (o *SummaryTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SummaryTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SummaryTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SummaryTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SummaryTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	return nil
}

func (o *SummaryTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
