package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ManpowerRequestDetail struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string `bson:"_id" json:"_id"`
	EducationQualification string
	Major                  string
	MinYearExperience      int
	WorkingLocation        string
	MaximumSalary          float64
	Benefit                []string
	Skill                  []string
	Gender                 string
	Age                    int
	MaritalStatus          string
	JobVancancyDeadline    time.Time
	POH                    string
	RequiredCertificate    string
	JobTitle               string
	JobDesription          string
	Requirement            string
	Created                time.Time `form_kind:"datetime" grid:"hide"`
	LastUpdate             time.Time `form_kind:"datetime" grid:"hide"`
}

func (o *ManpowerRequestDetail) TableName() string {
	return "HCMManpowerRequestDetails"
}

func (o *ManpowerRequestDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ManpowerRequestDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ManpowerRequestDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ManpowerRequestDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ManpowerRequestDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ManpowerRequestDetail) PostSave(dbflex.IConnection) error {
	return nil
}
