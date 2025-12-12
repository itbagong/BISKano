package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Likelihood struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_read_only:"1" grid_sortable:"1" grid_keyword:"1" form_pos:"1"`
	Level             int       `form_required:"1" form_section:"General"`
	Value             int       `form_section:"General"`
	ParameterName     string    `form_section:"General"`
	Criteria          string    `form_required:"1" form_multi_row:"4"`
	CompanyID         string    `grid:"hide" form:"hide"`
	Created           time.Time `grid:"hide" form_kind:"datetime" form_read_only:"1" grid_sortable:"1" form:"hide"`
	Type              string    `grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
}

func (o *Likelihood) TableName() string {
	return "Likelihoods"
}

func (o *Likelihood) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}

}

func (o *Likelihood) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *Likelihood) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Likelihood) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Likelihood) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Likelihood) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Likelihood) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "LikelihoodIndex", Fields: []string{"CompanyID"}},
	}
}
