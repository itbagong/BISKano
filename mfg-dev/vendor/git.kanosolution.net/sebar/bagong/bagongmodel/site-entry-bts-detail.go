package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RitaseFuelUsage struct {
	VendorID    string `form_lookup:"/tenant/vendor/find|_id|Name"`
	Volume      int
	KMStart     int     `label:"Km Start"`
	KMEnd       int     `label:"Km End"`
	KMTempuh    int     `label:"Km Tempuh"`
	Amount      float64 `grid_label:"Price Each"`
	TotalAmount float64 `form_read_only:"1" grid_label:"Amount"`
	Notes       string
}

type RitaseDetail struct {
	Name               string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MRB|_id|_id,Name" label:"Ritase Name"`
	DriverID           string `label:"Driver"`
	Status             string `form_items:"Running|Breakdown|Refueling"`
	ReplacementAssetID string `form_lookup:"/tenant/asset/find?GroupID=UNT|_id|Name" label:"Replacement Asset"`
	KMStart            int    `label:"Km Start"`
	KMEnd              int    `label:"Km End"`
	KMTotal            int    `form_read_only:"1" label:"Km Total"`
}

type SiteEntryBTSDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3" form:"hide"`
	KMStart           int                       `label:"Km Start" form_section:"General"`
	KMEnd             int                       `label:"Km End" form_section:"General2"`
	Status            string                    `form_items:"Running|Partial|Standby|Breakdown" form_section:"General3"`
	SpareAsset        string                    `grid:"hide" form_section:"General" form_label:"Spare Unit" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|Name"`
	RitaseDetail      []RitaseDetail            `form_section:"Ritase"`
	RitaseFuelUsage   []RitaseFuelUsage         `form:"hide"`
	Expense           []SiteExpense             `form:"hide"`
	CompanyID         string                    `form:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type SiteEntryBTSDetailAsset struct {
	SiteEntryBTSDetail
	SumRevenue float64
	SumIncome  float64
	SumExpense float64
	TrxDate    time.Time
	// TenantAssetID string
	PoliceNum    string
	CustomerName string
}

func (o *SiteEntryBTSDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Ritase", ShowTitle: true, AutoCol: 4},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *SiteEntryBTSDetail) TableName() string {
	return "BGSiteEntryBTSDetails"
}

func (o *SiteEntryBTSDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryBTSDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryBTSDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryBTSDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryBTSDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryBTSDetail) PostSave(dbflex.IConnection) error {
	return nil
}
