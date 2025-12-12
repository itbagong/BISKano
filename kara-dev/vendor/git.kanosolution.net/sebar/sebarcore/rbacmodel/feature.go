package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Feature struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_required:"1"`
	Name              string `form_required:"1"`
	Ref1              string `grid:"hide"`
	Ref2              string `grid:"hide"`
	Ref3              string `grid:"hide"`
	NeedDimension     bool   `label:"Need Dim"`
	FeatureCategoryID string `form_lookup:"admin/featurecategory/find|_id|Name"`
}

func (o *Feature) TableName() string {
	return "RbacFeatures"
}

func (o *Feature) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *Feature) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *Feature) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Feature) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Feature) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *Feature) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Feature) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "CategoryIndex", Fields: []string{"FeatureCategoryID"}},
	}
}
