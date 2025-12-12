package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryNonAsset struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string  `form:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" "`
	Revenue           float64 `form:"hide"`
	Expense           float64 `form:"hide"`
	Income            float64 `form:"hide"`
	ExpenseDetail     []SiteExpense
	Created           time.Time `form:"hide" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"  `
	LastUpdate        time.Time `form:"hide" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteEntryNonAsset) TableName() string {
	return "BGSiteEntryNonAssets"
}

func (o *SiteEntryNonAsset) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryNonAsset) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryNonAsset) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryNonAsset) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryNonAsset) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryNonAsset) PostSave(dbflex.IConnection) error {
	return nil
}
