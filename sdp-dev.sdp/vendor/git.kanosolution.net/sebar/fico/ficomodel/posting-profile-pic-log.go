package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostingProfilePICLog struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id"`
	PostingProfileName string
	PostingProfilePIC  *PostingProfilePIC
	UpdatedBy          string
	Created            time.Time
}

func (o *PostingProfilePICLog) TableName() string {
	return "PostingProfilePICLogs"
}

func (o *PostingProfilePICLog) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PostingProfilePICLog) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PostingProfilePICLog) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PostingProfilePICLog) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PostingProfilePICLog) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	return nil
}

func (o *PostingProfilePICLog) PostSave(dbflex.IConnection) error {
	return nil
}
