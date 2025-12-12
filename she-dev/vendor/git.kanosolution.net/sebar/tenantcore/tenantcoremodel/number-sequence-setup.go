package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NumberSequenceSetup struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CompanyID         string `form_lookup:"/tenant/company/find|_id|_id,Name"`
	Kind              string
	Label             string
	NumSeqID          string    `form_lookup:"/tenant/numseq/find|_id|_id,Name"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *NumberSequenceSetup) TableName() string {
	return "NumberSequenceSetups"
}

func (o *NumberSequenceSetup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *NumberSequenceSetup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *NumberSequenceSetup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *NumberSequenceSetup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *NumberSequenceSetup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *NumberSequenceSetup) PostSave(dbflex.IConnection) error {
	return nil
}
