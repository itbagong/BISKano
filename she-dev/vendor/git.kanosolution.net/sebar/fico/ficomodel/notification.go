package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	orm.DataModelBase        `bson:"-" json:"-"`
	CompanyID                string
	ID                       string `bson:"_id" json:"_id"`
	UserSubmitter            string
	UserSubmitterEmail       string
	JournalID                string
	JournalType              string
	PostingProfileApprovalID string
	TrxDate                  time.Time
	TrxType                  string
	Text                     string
	Menu                     string
	UserTo                   string
	UserToEmail              string
	Amount                   float64
	Status                   string
	IsRead                   bool
	URL                      string
	Message                  string
	IsApproval               bool
	Created                  time.Time
	LastUpdate               time.Time
}

func (o *Notification) TableName() string {
	return "Notifications"
}

func (o *Notification) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Notification) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Notification) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Notification) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Notification) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Notification) PostSave(dbflex.IConnection) error {
	return nil
}
