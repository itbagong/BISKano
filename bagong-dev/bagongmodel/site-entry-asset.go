package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryAsset struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string  `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid:"hide"`
	SiteEntryID       string  `grid:"hide"`
	AssetID           string  `grid:"hide"`
	Availability      string  `form_items:"Running|Partial|Standby|Breakdown"`
	Income            float64 `form_read_only:"1"`
	Expense           float64 `form_read_only:"1"`
	Bonus             float64 `grid:"hide" form:"hide"`
	Revenue           float64 `form_read_only:"1"`
	Running           int     `grid:"hide" form:"hide"`
	Standby           int     `grid:"hide" form:"hide"`
	Breakdown         int     `grid:"hide" form:"hide"`
	Notes             string
	ProjectID         string    `grid:"hide" form:"hide"`
	HullNo            string    `grid:"hide" form:"hide"`
	PoliceNo          string    `grid:"hide" form:"hide"`
	UnitType          string    `grid:"hide" form:"hide"`
	CustomerID        string    `grid:"hide" form:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteEntryAsset) TableName() string {
	return "BGSiteEntryAssets"
}

func (o *SiteEntryAsset) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryAsset) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryAsset) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryAsset) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryAsset) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryAsset) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *SiteEntryAsset) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "SiteEntryIDAssetIDIndex", Fields: []string{"SiteEntryID", "AssetID"}},
	}
}
