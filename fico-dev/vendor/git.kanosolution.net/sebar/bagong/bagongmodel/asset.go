package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetOtherInfo struct {
	AssetDuration      int        `form_section:"Other Info" form_section_auto_col:"3" form_section_show_title:"1"`
	DepreciationPeriod *time.Time `form_kind:"date" form_section:"Other Info"`
	User               string     `form_section:"Other Info" form_lookup:"/iam/user/find-by|_id|LoginID" label:"User"`
	CustomerID         string     `form_section:"Other Info" form_lookup:"/tenant/customer/find|_id|Name" grid:"hide"`
	NoHullBagong       string     `form_section:"Other Info"`
	NoHullCustomer     string     `form_section:"Other Info"`
}
type AssetRegisterInfo struct {
	LineNo            int
	ID                string     `form:"hide" grid:"hide"`
	PrevPoliceNum     string     `form_section:"Vehicle Info" label:"Previous Police No."`
	AnnualExpDate     *time.Time `form_kind:"date" form_section:"Register Info" form_section_direction:"row" form_section_size:"3" form_section_show_title:"1" label:"Annual Tax Expiration Date"   `
	YearsExpDate      *time.Time `form_kind:"date" form_section:"Register Info" label:"5 Years Tax Expiration Date"  `
	KirTestNum        string     `form_section:"Register Info" label:"Kir Test No." form_section_direction:"row" form_section_size:"3" `
	KirExpDate        *time.Time `form_kind:"date" form_section:"Register Info" label:"6 Months Kir Expiration Date"  `
	CommissioningDate *time.Time `form_kind:"date" form_section:"Register Info" label:"Commissioning Date"`
	STNK              bool       `form_section:"Register Info" label:"STNK"`
	Tax               bool       `form_section:"Register Info" label:"Tax" grid:"hide"`
	BPKB              bool       `form_section:"Register Info" label:"BPKB"`
	Kir               bool       `form_section:"Register Info" label:"KIR"`
}
type AssetUnit struct {
	Purpose        string              `form_section:"Vehicle Info" form_section_direction:"row" form_section_show_title:"1"  form_section_size:"4" form_items:"BTS|Mining|Trayek|Tourism"`
	TrayekID       string              `form_section:"Vehicle Info" label:"Trayek" form_lookup:"/bagong/trayek/find|_id|Name" `
	Owner          string              `form_section:"Vehicle Info"  `
	Coridor        string              `form_section:"Vehicle Info" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=BCU|_id|_id,Name"`
	UnitType       string              `form_section:"Vehicle Info" label:"Vehicle Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=VTA|_id|_id,Name"`
	PoliceNum      string              `form_section:"Vehicle Info" label:"Police No."`
	HullNum        string              `form_section:"Vehicle Info" label:"Hull No. (Lambung)" form_read_only_edit:"1"`
	Merk           string              `form_section:"Vehicle Info" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MUA|_id|_id,Name"`
	Color          string              `form_section:"Vehicle Info2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CUA|_id|_id,Name"`
	PurchaseCode   string              `form_section:"Vehicle Info2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=APC|_id|_id,Name" label:"Purchase Code" form_section_direction:"row" form_section_size:"3" form_section_show_title:"1"  `
	BPKBNO         string              `form_section:"Vehicle Info2" label:"BPKB No."`
	ProductionYear int                 `form_section:"Vehicle Info2" label:"Production Year"`
	CaroseriCode   string              `form_section:"Vehicle Info2" label:"Caroseri Code"`
	UnitCondition  string              `form_section:"Vehicle Info2" label:"Unit Condition" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=AUC|_id|_id,Name"`
	TrayekCode     string              `form_section:"Vehicle Info2" label:"Kode Trayek"`
	TrayekName     string              `form_section:"Vehicle Info2" label:"Nama Trayek" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=NTR|_id|Name"`
	Notes          string              `form_section:"Vehicle Info2" form_multi_row:"5"`
	Seat           string              `form_section:"Machine Info" form_section_direction:"row" form_section_size:"3" form_section_show_title:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SUA|_id|_id,Name"`
	MachineNum     string              `form_section:"Machine Info" label:"Machine No."  `
	ChassisNum     string              `form_section:"Machine Info" label:"Chassis No."  `
	RegisterInfo   []AssetRegisterInfo `form_section:"Register Info"`
	OtherInfo      AssetOtherInfo      `form:"hide"`
}

func (o *AssetUnit) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Vehicle Info", ShowTitle: true, AutoCol: 1},
			{Title: "Vehicle Info2", ShowTitle: true, AutoCol: 1},
			// {Title: "Vehicle Info3", ShowTitle: true, AutoCol: 1},
			{Title: "Machine Info", ShowTitle: true, AutoCol: 1},
		}}, {Sections: []suim.FormSection{
			{Title: "Register Info", ShowTitle: true, AutoCol: 1},
		}}, {Sections: []suim.FormSection{
			{Title: "date", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type AssetProperty struct {
	Owner          string `form_section:"Property Info" form_section_auto_col:"3" form_section_show_title:"1"`
	AssetType      string `form_section:"Property Info" label:"Type"`
	LandArea       string `form_section:"Property Info" label:"Land Area (m2)"`
	BuildingArea   string `form_section:"Property Info" label:"Building Area (m3)"`
	Address        string `form_section:"Property Info"`
	CertificateNum string `form_section:"Property Info" label:"Certificate No."`
	Notes          string `form_section:"Property Info" form_multi_row:"5"`
}

type AssetElectronic struct {
	Owner               string     `form_section:"Electronic Info" form_section_auto_col:"3" form_section_show_title:"1" form_pos:"1,1"`
	PurchaseDate        string     `form_section:"Electronic Info" form_pos:"1,2"`
	Merk                string     `form_section:"Electronic Info" form_pos:"2,1"`
	WarrantyExpDate     *time.Time `form_section:"Electronic Info" label:"Warranty Expiration Date" form_kind:"date" form_pos:"2,2"`
	Seri                string     `form_section:"Electronic Info" form_pos:"3,1"`
	WarrantyExpDateDays int        `form_section:"Electronic Info" label:"Warranty Expiration Date (in Days)" form_pos:"3,2"`
	SerialNum           string     `form_section:"Electronic Info" label:"Serial No./IMEI/Product Code" form_pos:"4,1"`
	Processor           string     `form_section:"Specification" form_pos:"1,1" form_section_show_title:"1"`
	VGA                 string     `form_section:"Specification" form_pos:"1,2" label:"VGA"`
	HDD                 string     `form_section:"Specification" form_pos:"2,1" label:"HDD"`
	OS                  string     `form_section:"Specification" form_pos:"2,2" label:"Operating System"`
	RAM                 string     `form_section:"Specification" form_pos:"3,1" label:"RAM"`
	SoftwareInstalled   string     `form_section:"Specification" form_pos:"3,2"`
	Notes               string     `form_section:"Specification" form_multi_row:"5" form_pos:"4"`
}

type Depreciation struct {
	AssetDuration      int        `form_section:"General" form_section_size:"3"`
	DepreciationPeriod string     `form_section:"General2" form_section_size:"3" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DEPP|_id|Name" form_section:"Depreciation"`
	AcquisitionDate    *time.Time `form_section:"General3" form_section_size:"3" form_kind:"date" form_section:"Depreciation"`
	DepreciationDate   *time.Time `form_section:"General4" form_section_size:"3" form_kind:"date" form_section:"Depreciation"`
	User               string     `form:"hide" form_section:"General" form_section_size:"3" form_lookup:"/iam/user/find-by|_id|LoginID" label:"User"`
	ResidualAmount     float64    `form_section:"General" form_section:"Depreciation"`
}

func (o *Depreciation) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "General2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "General3", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "General4", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type DepreciationActivity struct {
	ID                 string    `bson:"_id" json:"_id"`
	Date               time.Time `form_section:"DepreciationActivity" form_section_auto_col:"3" form_section_show_title:"1"`
	Activity           string    `form_section:"DepreciationActivity"`
	DepreciationAmount float64   `form_section:"DepreciationActivity"`
	AdjustmentAmount   float64   `form_section:"DepreciationActivity"`
	NetBookValue       float64   `form_section:"DepreciationActivity"`
	JournalID          string    `form_section:"DepreciationActivity"`
}

type UserInfo struct {
	ProjectID           string     `form_lookup:"/sdp/measuringproject/find|_id|ProjectName" form_read_only:"1"`
	SONumber            string     `label:"SO Number"`
	SOStartDate         *time.Time `form_kind:"date" form_section:"Other Info" label:"SO Start Date"`
	SOEndDate           *time.Time `form_kind:"date" form_section:"Other Info" label:"SO End Date"`
	AssetDateFrom       time.Time  `form_kind:"date" form_read_only:"1"`
	AssetDateTo         time.Time  `form_kind:"date" form_read_only:"1"`
	AssetDateFromString string     `grid:"hide"`
	AssetDateToString   string     `grid:"hide"`
	SiteID              string     `form_lookup:"/bagong/sitesetup/find|_id|Name" form_read_only:"1"`
	UserID              string     `form_lookup:"/tenant/employee/find|_id|Name"`
	CustomerID          string     `label:"Customer" form_lookup:"/tenant/customer/find|_id|Name" form_read_only:"1"`
	NoHullCustomer      string
	Description         string
}

type Asset struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"1" form:"hide"`
	Name                 string `grid:"hide" form:"hide"`
	IsActive             bool
	DetailUnit           AssetUnit                  `grid:"hide" form:"hide"`
	DetailProperty       AssetProperty              `grid:"hide" form:"hide"`
	DetailElectronic     AssetElectronic            `grid:"hide" form:"hide"`
	References           []codekit.M                `grid:"hide" form:"hide"`
	Depreciation         Depreciation               `grid:"hide" form:"hide"`
	UserInfo             []UserInfo                 `grid:"hide" form:"hide"`
	DepreciationActivity []DepreciationActivity     `grid:"hide" form:"hide"`
	ChecklistTemp        tenantcoremodel.Checklists `grid:"hide" form:"hide"`
	// for filter asset
	LatestCustomer string                    `form:"hide"`
	Dimension      tenantcoremodel.Dimension `form:"hide"`
	//-- di isi fields yang spesifik untuk bagong
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2" form:"hide"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form:"hide"`
}

func (o *Asset) TableName() string {
	return "BGAssets"
}

func (o *Asset) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Asset) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Asset) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Asset) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Asset) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Asset) PostSave(dbflex.IConnection) error {
	return nil
}
