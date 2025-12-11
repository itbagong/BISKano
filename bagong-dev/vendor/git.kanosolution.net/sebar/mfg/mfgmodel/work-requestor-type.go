package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkRequestorJournalType struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name              string `form_required:"1" form_section:"General"`
	PostingProfileID  string `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	TrxType           string `form_read_only:"1" form_items:"Work Order" form_section:"General2"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *WorkRequestorJournalType) TableName() string {
	return "WorkRequestorJournalTypes"
}

func (o *WorkRequestorJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkRequestorJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkRequestorJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkRequestorJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkRequestorJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkRequestorJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
