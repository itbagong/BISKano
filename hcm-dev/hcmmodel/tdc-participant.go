package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentParticipantStatus string

const (
	TrainingDevelopmentParticipantStatusOpen TrainingDevelopmentParticipantStatus = "OPEN"
	TrainingDevelopmentParticipantStatusDone TrainingDevelopmentParticipantStatus = "DONE"
)

type TrainingDevelopmentParticipant struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                                 `bson:"_id" json:"_id" grid:"hide" form:"hide" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	TrainingCenterID  string                                 `form_section:"General"`
	EmployeeID        string                                 `form_section:"General2" label:"Employee"`
	ManpowerRequestID string                                 `form_section:"General3"`
	TestDetails       []TrainingDevelopmentParticipantDetail `grid:"hide" form:"hide"`
	Created           time.Time                              `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                              `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type TrainingDevelopmentParticipantDetail struct {
	Stage        TDCTestType
	TemplateID   string
	TemplateName string
	Score        int
	Status       TrainingDevelopmentParticipantStatus
}

func (o *TrainingDevelopmentParticipant) TableName() string {
	return "HCMTrainingDevelopmentParticipants"
}

func (o *TrainingDevelopmentParticipant) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentParticipant) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentParticipant) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentParticipant) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentParticipant) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentParticipant) PostSave(dbflex.IConnection) error {
	return nil
}
