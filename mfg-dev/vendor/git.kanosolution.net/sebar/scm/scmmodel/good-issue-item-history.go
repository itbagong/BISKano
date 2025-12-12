package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoodIssueItemHistory struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	GoodIssueID        string
	ItemID             string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/item/find|_id|Name"`
	SKU                string `label:"SKU"`
	Qty                int
	UoM                string `label:"UOM"`
	InventoryDimension InventDimension `form_section:"Specification"`
	Created            time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *GoodIssueItemHistory) TableName() string {
	return "GoodIssueItemHistories"
}

func (o *GoodIssueItemHistory) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *GoodIssueItemHistory) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *GoodIssueItemHistory) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *GoodIssueItemHistory) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *GoodIssueItemHistory) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *GoodIssueItemHistory) PostSave(dbflex.IConnection) error {
	return nil
}
