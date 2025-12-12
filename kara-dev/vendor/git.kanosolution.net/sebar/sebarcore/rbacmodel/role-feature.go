package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleFeaturePure struct {
	FeatureID string
	Dimension Dimension
	Create    bool
	Read      bool
	Update    bool
	Delete    bool
	Posting   bool
	Special1  bool
	Special2  bool
	All       bool
}

type RoleFeature struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	RoleID            string
	FeatureID         string
	Create            bool
	Read              bool
	Update            bool
	Delete            bool
	Posting           bool
	Special1          bool
	Special2          bool
	All               bool
}

func (o *RoleFeature) TableName() string {
	return "RbacRoleFeatures"
}

func (o *RoleFeature) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *RoleFeature) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *RoleFeature) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoleFeature) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoleFeature) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *RoleFeature) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *RoleFeature) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "RoleFeatureIndex", Fields: []string{"RoleID", "FeatureID"}, IsUnique: true},
	}
}
