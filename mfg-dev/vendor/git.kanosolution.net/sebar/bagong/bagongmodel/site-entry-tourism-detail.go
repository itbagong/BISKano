package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryTourismDetail struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3" form:"hide"`
	DriverID           string                    `form_section:"General" label:"Driver"`
	DriverID2          string                    `form_section:"General" label:"Driver 2" form_section_direction:"row" form_section_size:"3"`
	Helper             string                    `form_section:"General" form_section_direction:"row" form_section_size:"3"`
	Status             string                    `form_items:"Running|Partial|Standby|Breakdown" form_section:"General3"`
	Rate               float64                   `form_section:"General3"`
	Fine               float64                   `form_section:"General3"`
	TimesOut           string                    `form_kind:"time" form_section:"General2" label:"Time Out"`
	DateOut            time.Time                 `form_kind:"date" form_section:"General2"`
	TimesIn            string                    `form_kind:"time" form_section:"General2" label:"Time In"`
	DateIn             time.Time                 `form_kind:"date" form_section:"General2"`
	SpareAsset         string                    `grid:"hide" form_section:"General3" form_label:"Spare Unit" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|Name"`
	OperationalExpense []SiteExpense             `form:"hide"`
	OtherExpense       []SiteExpense             `form:"hide"`
	CompanyID          string                    `form_section:"General2" form:"hide"`
	Dimension          tenantcoremodel.Dimension `form_section:"Dimension"  form_section_direction:"row" form_section_size:"3"`
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type SiteEntryTourismDetailAsset struct {
	SiteEntryTourismDetail
	SumRevenue float64
	SumIncome  float64
	SumExpense float64
	TrxDate    time.Time
	// TenantAssetID string
	PoliceNum    string
	CustomerName string
}

func (o *SiteEntryTourismDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *SiteEntryTourismDetail) TableName() string {
	return "BGSiteEntryTourismDetails"
}

func (o *SiteEntryTourismDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryTourismDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryTourismDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryTourismDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryTourismDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryTourismDetail) PostSave(dbflex.IConnection) error {
	return nil
}
