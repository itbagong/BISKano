package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HolidayItem struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	HolidayGroupID    string `form_lookup:"kara/holiday/find|_id|Name"`
	Name              string
	Date              time.Time `form_kind:"date"`
	HolidayType       string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *HolidayItem) TableName() string {
	return "KaraHolidayItem"
}

func (o *HolidayItem) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *HolidayItem) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *HolidayItem) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *HolidayItem) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *HolidayItem) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *HolidayItem) PostSave(dbflex.IConnection) error {
	return nil
}
