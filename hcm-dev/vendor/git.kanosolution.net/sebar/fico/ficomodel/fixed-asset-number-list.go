package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FixedAssetNumberList struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form:"hide" grid:"hide"`
	AssetName          string
	FixedAssetNumberID string
	FixedAssetGrup     string
	GroupCode          string
	Sequence           int
	IsUsed             bool
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *FixedAssetNumberList) TableName() string {
	return "FixedAssetNumberLists"
}

func (o *FixedAssetNumberList) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *FixedAssetNumberList) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *FixedAssetNumberList) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FixedAssetNumberList) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FixedAssetNumberList) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *FixedAssetNumberList) PostSave(dbflex.IConnection) error {
	return nil
}
