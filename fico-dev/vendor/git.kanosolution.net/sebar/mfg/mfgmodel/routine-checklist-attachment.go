package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutineChecklistAttachment struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1" grid:"hide"`
	RoutineDetailID    string    `form_section:"General" grid:"hide"`
	RoutineChecklistID string    `form_section:"General" grid:"hide"`
	FileName           string    `form_section:"General"`
	Description        string    `form_section:"General"`
	PIC                string    `form_section:"General"`
	UploadDate         time.Time `form_section:"General" form_kind:"date"`
	URI                string    `form_section:"General" grid_label:"Files"`
	ContentType        string    `form_section:"General" grid:"hide"`
	Size               int       `form_section:"General" grid:"hide"`
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RoutineChecklistAttachment) TableName() string {
	return "RoutineChecklistAttachments"
}

func (o *RoutineChecklistAttachment) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RoutineChecklistAttachment) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RoutineChecklistAttachment) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoutineChecklistAttachment) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoutineChecklistAttachment) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RoutineChecklistAttachment) PostSave(dbflex.IConnection) error {
	return nil
}
