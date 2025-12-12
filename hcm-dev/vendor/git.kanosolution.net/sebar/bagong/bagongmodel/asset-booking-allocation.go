package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// todo: some column auto generated
type AssetBookingAllocation struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	AssetBookingID    string
	AssetID           string `form_lookup:"/tenant/asset/find|_id|Name"`
	Utilization       string
	Notes             string
	LinesIdx          int       `grid:"hide" form:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AssetBookingAllocation) TableName() string {
	return "BGAssetBookingAllocations"
}

func (o *AssetBookingAllocation) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetBookingAllocation) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetBookingAllocation) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetBookingAllocation) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetBookingAllocation) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetBookingAllocation) PostSave(dbflex.IConnection) error {
	return nil
}
