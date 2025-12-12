package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DimensionMaster struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	DimensionType     string `form_read_only_edit:"1"`
	Label             string
	ParentDimensionID string
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *DimensionMaster) TableName() string {
	return "DimensionMasters"
}

func (o *DimensionMaster) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *DimensionMaster) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *DimensionMaster) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *DimensionMaster) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *DimensionMaster) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *DimensionMaster) PostSave(dbflex.IConnection) error {
	return nil
}
