package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostingProfileApprovalLog struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id"`
	PostingProfile    *PostingProfile
	Journal           orm.DataModel
	Action            string
	Created           time.Time
	LastUpdate        time.Time
}

func (o *PostingProfileApprovalLog) TableName() string {
	return "PostingProfileApprovalLogs"
}

func (o *PostingProfileApprovalLog) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PostingProfileApprovalLog) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PostingProfileApprovalLog) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PostingProfileApprovalLog) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PostingProfileApprovalLog) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PostingProfileApprovalLog) PostSave(dbflex.IConnection) error {
	return nil
}
