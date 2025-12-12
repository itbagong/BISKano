package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureCategory struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_required:"1"`
	Name              string `form_required:"1"`
}

func (o *FeatureCategory) TableName() string {
	return "RbacFeatureCategories"
}

func (o *FeatureCategory) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *FeatureCategory) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *FeatureCategory) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FeatureCategory) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FeatureCategory) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	return nil
}

func (o *FeatureCategory) PostSave(dbflex.IConnection) error {
	return nil
}
