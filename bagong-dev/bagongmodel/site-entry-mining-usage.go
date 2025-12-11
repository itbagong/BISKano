package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TirePosition struct {
	Position  int
	TireType  string
	SerialNum string
}

type SiteEntryMiningUsage struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string         `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" form:"hide"`
	IsTireChange      bool           `form_section:"Tire1"`
	TirePosition      []TirePosition `form_section:"Tire2"`
	TireChangePlan    time.Time      `form_kind:"date" form_section:"Tire3" form_section_auto_col:"2"`
	TireType          string         `form_section:"Tire3"`
	IsOilChange       bool           `form_section:"Oil" form_section_auto_col:"3"`
	OilUsage          float64        `form_section:"Oil"`
	OilNotes          string         `form_section:"Oil"`
	CompanyID         string         `form:"hide"`
	Created           time.Time      `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time      `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteEntryMiningUsage) TableName() string {
	return "BGSiteEntryMiningUsages"
}

func (o *SiteEntryMiningUsage) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryMiningUsage) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryMiningUsage) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryMiningUsage) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryMiningUsage) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryMiningUsage) PostSave(dbflex.IConnection) error {
	return nil
}
