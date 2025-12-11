package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PKWTT struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string    `bson:"_id" json:"_id" form:"hide"`
	CandidateID         string    `form:"hide"`
	JobVacancyID        string    `form:"hide"`
	JoinedDate          time.Time `form_section:"General"  form_kind:"date" form_section_auto_col:"2"`
	ExpiredContractDate time.Time `form_section:"General"  form_kind:"date" form_section_auto_col:"2"`
	Status              string    `form:"hide"`
	Created             time.Time `grid:"hide" form:"hide"`
	LastUpdate          time.Time `grid:"hide" form:"hide"`
}

func (o *PKWTT) TableName() string {
	return "HCMPKWTTs"
}

func (o *PKWTT) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PKWTT) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PKWTT) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PKWTT) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PKWTT) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PKWTT) PostSave(dbflex.IConnection) error {
	return nil
}
