package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Severity struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	Level             int    `form_required:"1" form_section_width:"1"`
	Value             int
	ParameterName     string
	Criteria          string    `form_required:"1" form_multi_row:"4"`
	CompanyID         string    `grid:"hide" form:"hide"`
	Type              string    `grid:"hide"`
	Created           time.Time `grid:"hide" form_kind:"datetime" form_read_only:"1" grid_sortable:"1" form:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
}

func (o *Severity) TableName() string {
	return "Severity"
}

func (o *Severity) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}

}

func (o *Severity) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *Severity) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Severity) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Severity) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Severity) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Severity) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "SeverityIndex", Fields: []string{"CompanyID"}},
	}
}
