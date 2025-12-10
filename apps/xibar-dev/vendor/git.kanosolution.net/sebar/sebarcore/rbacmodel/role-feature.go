package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/samber/lo"
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

func (o *RoleFeature) Level() int {
	res := 0

	if o.All {
		return 255
	}

	res += lo.Ternary(o.Read, 1, 0)
	res += lo.Ternary(o.Create, 2, 0)
	res += lo.Ternary(o.Update, 4, 0)
	res += lo.Ternary(o.Delete, 8, 0)
	res += lo.Ternary(o.Special1, 16, 0)
	res += lo.Ternary(o.Special2, 32, 0)
	res += lo.Ternary(o.Posting, 128, 0)

	return res
}
