package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentAttendanceDetail struct {
	EmployeeID string
	IsPresent  bool
}

type TrainingDevelopmentAttendance struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" grid:"hide" form:"hide"`
	TrainingCenterID  string `grid:"hide" form:"hide"`
	Date              time.Time
	Time         	  string
	Topic             string
	LocationID        string
	TrainerType       string `grid:"hide"`
	Trainer           string 
	List              []TrainingDevelopmentAttendanceDetail `grid:"hide" form:"hide"`
	Attendace         int `form:"hide"`
	Created           time.Time `grid:"hide" form:"hide"`
	LastUpdate        time.Time `grid:"hide" form:"hide"`
}

func (o *TrainingDevelopmentAttendance) TableName() string {
	return "HCMTrainingDevelopmentAttendances"
}

func (o *TrainingDevelopmentAttendance) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentAttendance) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentAttendance) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentAttendance) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentAttendance) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentAttendance) PostSave(dbflex.IConnection) error {
	return nil
}
