package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutineChecklistType string
type RoutineChecklistStatus string

const (
	RoutineChecklistStatusNormal  RoutineChecklistStatus = "Normal"
	RoutineChecklistStatusDamaged RoutineChecklistStatus = "Damaged"
)

type RoutineChecklistDetail struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                 `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1" grid:"hide"`
	RoutineChecklistID string                 `form_section:"General" grid:"hide"`
	CategoryID         string                 `grid:"hide"`
	Name               string                 `form_section:"General"`
	Status             RoutineChecklistStatus `form_section:"General"`
	Code               string                 `form_section:"General"`
	Note               string                 `form_section:"General"`
	Created            time.Time              `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate         time.Time              `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RoutineChecklistDetail) TableName() string {
	return "RoutineChecklistDetails"
}

func (o *RoutineChecklistDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RoutineChecklistDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RoutineChecklistDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoutineChecklistDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoutineChecklistDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RoutineChecklistDetail) PostSave(dbflex.IConnection) error {
	return nil
}
