package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WistleBlower struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form:"hide"`
	Date              time.Time `form_kind:"date" form_section:"General" form_section_auto_col:"2"`
	ComplainType      string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CPT|_id|Name" grid_label:"Location" form_section:"General"`
	Unit              string    `form_section:"General" label:"Unit"`
	Site              string    `form_section:"General" label:"Site"`
	FillInTheComplain string    `form_section:"General2" form_multi_row:"5" label:"Fill in the Complain"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *WistleBlower) TableName() string {
	return "SHEWistleBlowers"
}

func (o *WistleBlower) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WistleBlower) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WistleBlower) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WistleBlower) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WistleBlower) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WistleBlower) PostSave(dbflex.IConnection) error {
	return nil
}
