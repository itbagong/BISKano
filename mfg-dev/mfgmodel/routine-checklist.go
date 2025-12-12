package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutineChecklist struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"1" form_section_auto_col:"4" form_section:"General" form:"hide" form_pos:"1,1"` // 1-to-1 with RoutineDetail
	Name               string `form_section:"General" form_use_list:"1" form_lookup:"/tenant/employee/find|_id|Name"`
	Department         string
	Shift              string     `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SHFT|_id|Name"`
	WorkLocation       string     `form_section:"General"`

	KmToday            float64    `form_section:"General"`
	TimeBreakdown      string     `form_kind:"time"`
	Departure          *time.Time `form_section:"General" form_kind:"datetime-local"`
	Arrive             *time.Time `form_section:"General" form_kind:"datetime-local" label:"Arrival"`

	HelperName         string     `form_section:"General"  label:"Helper"`
	BBMLevel           string     `form_section:"General"  label:"BBM Level %"`
	Driver1           string     `form_section:"General"  label:"Driver 1"`
	Driver2           string     `form_section:"General"  label:"Driver 2"`
	Marketing           string     `form_section:"General"  label:"Marketing"`
	
	CategoryIDs        []string   `form:"hide"`
	IsAlreadyRequest   bool       `form:"hide"`
	RoutineTemplateIDs []string   `form:"hide" grid:"hide"`
	Created            time.Time  `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate         time.Time  `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *RoutineChecklist) TableName() string {
	return "RoutineChecklists"
}

func (o *RoutineChecklist) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RoutineChecklist) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RoutineChecklist) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoutineChecklist) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoutineChecklist) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RoutineChecklist) PostSave(dbflex.IConnection) error {
	return nil
}
