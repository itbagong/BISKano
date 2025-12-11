package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryAssetBKP struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	SiteID            string
	AssetID           string
	Driver            string
	RevenueType       string    `form_items:"RENTAL|PREMI|SETORAN|PARIWISATA|BTS"`
	TrxDate           time.Time `form_kind:"date"`
	Availability      string    `form_items:"Running|Partial|Standby|Broken"`
	Notes             string
	BrokenHour        float64
	Revenue           float64 `form_read_only:"1"`
	Expense           float64 `form_read_only:"1"`
	ProfitLoss        float64 `form_read_only:"1"`
	RentalRate        float64
	PremiRate         float64
	FullRitase        int
	FullRitaseRate    float64
	CompanyID         string
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteEntryAssetBKP) TableName() string {
	return "BGSiteEntryAssets"
}

func (o *SiteEntryAssetBKP) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryAssetBKP) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryAssetBKP) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryAssetBKP) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryAssetBKP) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryAssetBKP) PostSave(dbflex.IConnection) error {
	return nil
}
