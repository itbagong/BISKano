package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostingProfileLog struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	PostingProfile    *PostingProfile
	UpdatedBy         string
	Created           time.Time
}

func (o *PostingProfileLog) TableName() string {
	return "PostingProfileLogs"
}

func (o *PostingProfileLog) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PostingProfileLog) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PostingProfileLog) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PostingProfileLog) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PostingProfileLog) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	return nil
}

func (o *PostingProfileLog) PostSave(dbflex.IConnection) error {
	return nil
}
