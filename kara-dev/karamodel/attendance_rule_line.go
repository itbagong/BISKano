package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RuleLine struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	RuleID            string `form:"hide" grid:"hide"`
	WorkStart         string `form_kind:"time"`
	WorkEnd           string `form_kind:"time"`
	CheckinStart      string `form_kind:"time"`
	CheckinEnd        string `form_kind:"time"`
	CheckoutStart     string `form_kind:"time"`
	CheckoutEnd       string `form_kind:"time"`
	PersonPerBlock    int    `form_section:"General"`
	MinimumHour       float32
	Days              []int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RuleLine) TableName() string {
	return "KaraRuleLines"
}

func (o *RuleLine) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RuleLine) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RuleLine) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RuleLine) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RuleLine) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RuleLine) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *RuleLine) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "RuleIndex", Fields: []string{"RuleID"}},
	}
}
