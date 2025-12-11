package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShiftPlan struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	WorkLocationID    string    `form:"hide" grid:"hide"`
	RuleLineID        string    `form_lookup:"kara/ruleline/find|_id|Name"`
	ShiftDate         time.Time `form_kind:"date"`
	UserID            string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ShiftPlan) TableName() string {
	return "KaraShiftPlans"
}

func (o *ShiftPlan) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ShiftPlan) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ShiftPlan) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ShiftPlan) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ShiftPlan) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ShiftPlan) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *ShiftPlan) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "User_RuleLine_WorkLocation_Index", Fields: []string{"RuleLineID", "UserID", "WorkLocationID"}},
	}
}
