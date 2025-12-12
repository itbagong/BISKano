package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryTrayekDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3" form:"hide"`
	DriverID          string                    `label:"Driver" form_section:"General" `
	DriverID2         string                    `label:"Driver 2" form_section:"General" `
	KondekturID       string                    `label:"Kondektur" form_section:"General2" form_section_size:"3"  `
	KondekturID2      string                    `label:"Kondektur 2" form_section:"General2" form_section_size:"3"  `
	KernetID          string                    `label:"Kernet" form_section:"General3" form_section_size:"3"`
	KernetID2         string                    `label:"Kernet 2" form_section:"General3" form_section_size:"3"`
	Status            string                    `form_section:"General" form_section_auto_col:"3" form_items:"Running|Partial|Standby|Breakdown|Breakdown No Crew"`
	Notes             string                    `form_section:"General2" grid:"hide"`
	SpareAsset        string                    `grid:"hide" form_section:"General3" form_label:"Spare Unit" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|Name"`
	RateDeposit       float64                   `form:"hide" form_section:"General3" label:"Rate Deposit" form_section_show_title:"1"`
	CompanyID         string                    `form:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type SiteEntryTrayekDetailAsset struct {
	SiteEntryTrayekDetail
	SumRevenue float64
	SumIncome  float64
	SumExpense float64
	TrxDate    time.Time
	// TenantAssetID string
	PoliceNum    string
	CustomerName string
}

func (o *SiteEntryTrayekDetail) FormSections() []suim.FormSectionGroup {
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

func (o *SiteEntryTrayekDetail) TableName() string {
	return "BGSiteEntryTrayekDetails"
}

func (o *SiteEntryTrayekDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryTrayekDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryTrayekDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryTrayekDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryTrayekDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryTrayekDetail) PostSave(dbflex.IConnection) error {
	return nil
}
