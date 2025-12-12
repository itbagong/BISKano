package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TestScheduleType string
type TDCTestType string
type TestScheduleStatus string

const (
	TestScheduleTypePsychological TestScheduleType = "PSYCHOLOGICAL"
	TestScheduleTypeTDC           TestScheduleType = "TDC"
	TestScheduleTypeTD            TestScheduleType = "TALENTDEVELOPMENT"

	TestScheduleStatusOpen TestScheduleStatus = "OPEN"
	TestScheduleStatusDone TestScheduleStatus = "DONE"

	TDCTestTypePre  TDCTestType = "Pre-Test"
	TDCTestTypePost TDCTestType = "Post-Test"
)

type TestSchedule struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string             `bson:"_id" json:"_id" grid:"hide" form:"hide"`
	TestID            string             `form_read_only:"1"` // man power request id or training id or talent development id
	TestScheduleType  TestScheduleType   `form_items:"PSYCHOLOGICAL|TDC|TALENTDEVELOPMENT"`
	TemplateID        string             `form_lookup:"/she/mcuitemtemplate/find?Module=HCGS&Menu=PSI|_id|Name"`
	Status            TestScheduleStatus `form_read_only:"1"`
	DateFrom          *time.Time         `form_kind:"date"`
	DateTo            *time.Time         `form_kind:"date"`
	TrainerType       string
	TrainerName       string
	TestType          TDCTestType
	Created           time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *TestSchedule) TableName() string {
	return "HCMTestSchedules"
}

func (o *TestSchedule) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TestSchedule) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TestSchedule) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TestSchedule) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TestSchedule) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TestSchedule) PostSave(dbflex.IConnection) error {
	return nil
}
