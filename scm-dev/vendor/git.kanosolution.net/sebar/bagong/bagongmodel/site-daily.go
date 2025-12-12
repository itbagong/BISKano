package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteDaily struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	SiteID            string    `form_lookup:"/tenant/dimension/find?DimensionType=Site|_id|Label"`
	TrxDate           time.Time `form_kind:"date"`
	Notes             string
	Status            string `form_items:"Complete|Pending"`
	Running           int
	Broken            int
	Partial           int
	Standby           int
	Revenue           float64
	Expense           float64
	ProfitLoss        float64
	CompanyID         string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteDaily) TableName() string {
	return "BGSiteDailys"
}

func (o *SiteDaily) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteDaily) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteDaily) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteDaily) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteDaily) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteDaily) PostSave(dbflex.IConnection) error {
	return nil
}
