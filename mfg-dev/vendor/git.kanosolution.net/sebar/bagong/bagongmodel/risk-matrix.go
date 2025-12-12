package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RiskMatrix struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_read_only:"1" grid_sortable:"1" grid_keyword:"1" form_pos:"1"`
	LikelihoodID      string    `form:"hide" grid:"hide"`
	LikelihoodName    string    `grid_keyword:"1" label:"LikelihoodName" form:"hide" grid_pos:"1"`
	SeverityID        string    `form:"hide" grid:"hide"`
	SeverityName      string    `grid_keyword:"1" label:"ConsequenceName" form:"hide" grid_pos:"1"`
	RiskID            string    `form:"hide" grid:"hide"`
	Value             int       `form:"hide"`
	Type              string    `grid:"hide"`
	CompanyID         string    `grid:"hide" form:"hide"`
	Created           time.Time `grid:"hide" form_kind:"datetime" form_read_only:"1" grid_sortable:"1" form:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
}

func (o *RiskMatrix) TableName() string {
	return "RiskMatrixes"
}

func (o *RiskMatrix) FK() []*orm.FKConfig {
	return []*orm.FKConfig{
		{FieldID: "LikelihoodID", RefTableName: new(Likelihood).TableName(), RefField: "_id", Map: codekit.M{"LikelihoodName": "ParameterName"}},
		{FieldID: "SeverityID", RefTableName: new(Severity).TableName(), RefField: "_id", Map: codekit.M{"SeverityName": "ParameterName"}},
		// {FieldID: "RiskRatingID", RefTableName: new(RiskRating).TableName(), RefField: "_id", Map: codekit.M{"RiskRatingCode": "Code", "RiskRatingColor": "Color", "RiskRatingName": "Name"}},
	}

}

func (o *RiskMatrix) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *RiskMatrix) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RiskMatrix) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RiskMatrix) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RiskMatrix) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *RiskMatrix) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "RiskMatrixIndex", Fields: []string{"LikelihoodID", "ConsequenceID", "CompanyID"}},
	}
}
