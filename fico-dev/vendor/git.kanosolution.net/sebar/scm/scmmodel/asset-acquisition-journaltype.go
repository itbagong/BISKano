package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetAcquisitionJournalType struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string        `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General"  form_section_size:"3"`
	Name              string        `form_required:"1" form_section:"General"`
	TrxType           InventTrxType `form_read_only:"1" form_items:"Asset Acquisition" form_section:"General2"`
	PostingProfileID  string        `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	Created           time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate        time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *AssetAcquisitionJournalType) TableName() string {
	return "AssetAcquisitionJournalTypes"
}

func (o *AssetAcquisitionJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetAcquisitionJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetAcquisitionJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetAcquisitionJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetAcquisitionJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetAcquisitionJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
