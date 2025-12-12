package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ModuleAssetmovement tenantcoremodel.TrxModule = "ASSETMOVEMENT"
)

var SourceTypeURLMap = map[string]string{
	string(ModuleAssetmovement): "bagong/asset-movement",
}

type AssetMovement struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string              `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_direction:"row" form_section_size:"2"`
	JournalTypeID     string              `form_lookup:"/fico/assetjournaltype/find|_id|_id,Name" form_section:"General"`
	Status            string              `form_read_only:"1" form_section:"General"`
	TrxDate           time.Time           `label:"Transaction date" form_section:"General2"`
	Total             int                 `form_read_only:"1" form_section:"General2"`
	PostingProfileID  string              `form:"hide"`
	Lines             []AssetMovementLine `form:"hide" grid:"hide"`
	Created           time.Time           `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time           `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type AssetMovementLine struct {
	LineNo           int    `grid:"hide"`
	AssetID          string `form_lookup:"/tenant/asset/find|_id|Name"`
	AssetName        string
	SiteFrom         string `form_lookup:"/bagong/sitesetup/find|_id|Name"`
	PCFrom           string `grid:"hide"`
	CustomerFrom     string `form_lookup:"/tenant/customer/find|_id|Name"`
	CustomerFromName string `grid:"hide"`
	ProjectFrom      string `form_lookup:"/sdp/measuringproject/find|_id|ProjectName"`
	SiteTo           string `form_lookup:"/bagong/sitesetup/find|_id|Name"`
	PCTo             string `grid:"hide"`
	CustomerTo       string `form_lookup:"/tenant/customer/find|_id|Name"`
	ProjectTo        string `form_lookup:"/sdp/measuringproject/find|_id|ProjectName"`
	NoHullCustomer   string
	DateFrom         time.Time `form_kind:"date"`
	DateTo           time.Time `form_kind:"date"`
	IsChecked        bool
}

func (o *AssetMovement) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}
func (o *AssetMovement) TableName() string {
	return "BGAssetMovements"
}

func (o *AssetMovement) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetMovement) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetMovement) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetMovement) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetMovement) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == "" {
		o.Status = "DRAFT"
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetMovement) PostSave(dbflex.IConnection) error {
	return nil
}
