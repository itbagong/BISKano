package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkOrderJournalType struct {
	orm.DataModelBase           `bson:"-" json:"-"`
	ID                          string                     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	TrxType                     string                     `form_read_only:"1" form_items:"Work Order" form_section:"General2"`
	Name                        string                     `form_required:"1" form_section:"General"`
	PostingProfileID            string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"` // for WO Plan
	PostingProfileIDConsumption string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"` // for WO Report - Consumption
	PostingProfileIDResource    string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"` // for WO Report - Resource
	PostingProfileIDOutput      string                     `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"` // for WO Report - Output
	DefaultOffsiteConsumption   ficomodel.SubledgerAccount `grid:"hide"  label:"Accrual" form_section_show_title:"1"`
	DefaultOffsiteManPower      ficomodel.SubledgerAccount `grid:"hide" form:"hide" label:"Default Offsite ManPower" form_section_show_title:"1"`
	Created                     time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate                  time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *WorkOrderJournalType) TableName() string {
	return "WorkOrderJournalTypes"
}

func (o *WorkOrderJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
