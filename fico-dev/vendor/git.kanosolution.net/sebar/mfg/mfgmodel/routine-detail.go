package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutineDetailStatusCondition string

const (
	RoutineDetailStatusConditionNotChecked  RoutineDetailStatusCondition = "NotCheckedYet"
	RoutineDetailStatusConditionRepaired    RoutineDetailStatusCondition = "Need Repaire"
	RoutineDetailStatusConditionRunningWell RoutineDetailStatusCondition = "RunningWell"
)

type RoutineDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                       `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1"` // 1-to-1 with RoutineChecklist
	RoutineID         string                       `form_section:"General"`
	AssetID           string                       `form_section:"General"`
	AssetType         string                       `form_section:"General"` // for lookup to Template
	DriveType         string                       `form_section:"General"` // for lookup to Template
	CustomerID        string                       `form_section:"General"` // for lookup to Template
	Dimension         tenantcoremodel.Dimension    `form_section:"General"` // for lookup to Template
	StatusCondition   RoutineDetailStatusCondition `form_section:"General"`
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RoutineDetail) TableName() string {
	return "RoutineDetails"
}

func (o *RoutineDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RoutineDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RoutineDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoutineDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoutineDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RoutineDetail) PostSave(dbflex.IConnection) error {
	return nil
}
