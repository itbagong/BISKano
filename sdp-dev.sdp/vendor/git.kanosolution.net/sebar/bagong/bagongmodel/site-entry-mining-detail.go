package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryMiningDetail struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1"  form_section:"General" form_section_direction:"row" form_section_size:"3" form:"hide"`
	DriverType           string                    `label:"Driver Type" form_section:"General" form_section_size:"3" form_items:"All|Driver|Non Driver"`
	DriverID             string                    `label:"Driver" form_section:"General" form_section_size:"3"`
	DriverIDReplacement  string                    `label:"Driver Replacement" form_section:"General" form_section_size:"3"`
	DriverID2            string                    `label:"Driver 2" form_section:"General" form_section_size:"3"`
	DriverID2Replacement string                    `label:"Driver 2 Replacement" form_section:"General" form_section_size:"3"`
	Status               string                    `form_items:"Running|Partial|Standby|Standby Ready|Breakdown|Breakdown No Driver" form_section:"General2" form_section_direction:"row"  form_section_size:"3"`
	BreakdownHour        float64                   `form_section:"General2"`
	StandbyHour          float64                   `form_section:"General2"`
	OvertimeHour         float64                   `grid:"hide" form:"hide" form_section:"General2"`
	Note                 string                    `form_section:"General2"`
	StartKM              int                       `grid:"hide" form_section:"General3" form_section_direction:"row" form_label:"Start KM" form_section_size:"3"`
	EndKM                int                       `grid:"hide" form_section:"General3" form_label:"End KM"`
	SpareAsset           string                    `grid:"hide" form_section:"General3" form_label:"Spare Unit" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|Name"`
	RateRental           float64                   `form:"hide" form_section:"Rate" form_section_auto_col:"3" form_section_direction:"row"`
	RateBreakdown        float64                   `form:"hide" form_section:"Rate"`
	RateStandby          float64                   `form:"hide" form_section:"Rate"`
	RateOvertime         float64                   `form:"hide" form_section:"Rate"`
	Expense              []SiteExpense             `form:"hide"`
	Attachment           []SiteAttachment          `form:"hide"`
	CompanyID            string                    `form:"hide"`
	Dimension            tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created              time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type SiteEntryMiningDetailAsset struct {
	SiteEntryMiningDetail
	SumRevenue float64
	SumIncome  float64
	SumExpense float64
	TrxDate    time.Time
	// TenantAssetID string
	PoliceNum    string
	CustomerName string
}

func (o *SiteEntryMiningDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Rate", ShowTitle: false, AutoCol: 4},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *SiteEntryMiningDetail) TableName() string {
	return "BGSiteEntryMiningDetails"
}

func (o *SiteEntryMiningDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryMiningDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryMiningDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryMiningDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryMiningDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryMiningDetail) PostSave(dbflex.IConnection) error {
	return nil
}
