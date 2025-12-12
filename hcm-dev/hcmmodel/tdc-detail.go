package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopmentDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id"`
	CompanyID         string `grid:"hide" form:"hide"`
	JournalTypeID     string `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID  string `grid:"hide" form:"hide"`
	TrainingCenterID  string
	TrainingDateFrom  time.Time
	TrainingDateTo    time.Time
	ExternalTraining  bool
	AssessmentType    string
	ScheduledTraining bool
	TrainerType       string
	TrainerName       string
	Description       string
	DiscussionScope   string
	ParticipantTarget string
	CostPerPerson     int
	RequiredTool      string
	MaterialClass     string
	PracticeClass     string
	OnlineTraining    bool
	Location          string
	Site              string
	Batch             int
	Status            ficomodel.JournalStatus `form:"hide"`
	Created           time.Time               `grid:"hide" form:"hide"`
	LastUpdate        time.Time               `grid:"hide" form:"hide"`
}

func (o *TrainingDevelopmentDetail) TableName() string {
	return "HCMTrainingDevelopmentDetails"
}

func (o *TrainingDevelopmentDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopmentDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopmentDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopmentDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopmentDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopmentDetail) PostSave(dbflex.IConnection) error {
	return nil
}
