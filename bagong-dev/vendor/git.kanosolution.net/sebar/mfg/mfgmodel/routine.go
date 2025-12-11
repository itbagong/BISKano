package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: index SiteID & ExecutionDate

type Routine struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section:"General" form:"hide"`
	SiteID            string    `form_section:"General" form_required:"1" form_lookup:"/bagong/sitesetup/find|_id|Name"`
	ExecutionDate     time.Time `form_kind:"date" form_section:"General"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *Routine) TableName() string {
	return "Routines"
}

func (o *Routine) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Routine) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Routine) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Routine) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Routine) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Routine) PostSave(dbflex.IConnection) error {
	return nil
}
