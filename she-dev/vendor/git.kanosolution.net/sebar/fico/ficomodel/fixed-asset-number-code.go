package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FixedAssetNumberCode struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form:"hide" grid:"hide"`
	GroupCode         string
	LastSequence      int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *FixedAssetNumberCode) TableName() string {
	return "FixedAssetNumberCodes"
}

func (o *FixedAssetNumberCode) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *FixedAssetNumberCode) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *FixedAssetNumberCode) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FixedAssetNumberCode) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FixedAssetNumberCode) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *FixedAssetNumberCode) PostSave(dbflex.IConnection) error {
	return nil
}
