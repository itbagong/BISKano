package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomItemDownload struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `json:"_id" bson:"_id"`
	Created           time.Time
	ItemID            string
	LastUpdate        time.Time
	OtherName         string
	SKU               string
	SpecGradeID       string
	SpecID            string
	SpecSizeID        string
	ItemName          string
	SpecGradeName     string
	SpecVariantName   string
}

func (o *CustomItemDownload) TableName() string {
	return "item_sku"
}

func (o *CustomItemDownload) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomItemDownload) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomItemDownload) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomItemDownload) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomItemDownload) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomItemDownload) PostSave(dbflex.IConnection) error {
	return nil
}
