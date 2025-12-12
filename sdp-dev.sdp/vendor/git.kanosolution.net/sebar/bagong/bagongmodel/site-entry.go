package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntry struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"1" grid:"hide"  form:"hide"`
	SiteID            string    `form_lookup:"/bagong/sitesetup/find|_id|Name" label:"Site ID"`
	Purpose           string    `grid:"hide"  form:"hide"`
	TrxDate           time.Time `form_kind:"date" label:"Transaction Date"`
	Running           int       `form_read_only:"1" form:"hide"`
	Standby           int       `form_read_only:"1" form:"hide"`
	Breakdown         int       `form_read_only:"1" form:"hide"`
	Income            float64   `form_read_only:"1" form:"hide"`
	Expense           float64   `form_read_only:"1" form:"hide"`
	Revenue           float64   `form_read_only:"1" form:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" form:"hide" form_section:"Time Info" form_section_auto_col:"2" label:"Created Date"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *SiteEntry) TableName() string {
	return "BGSiteEntrys"
}

func (o *SiteEntry) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntry) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntry) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntry) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntry) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntry) PostSave(dbflex.IConnection) error {
	return nil
}
