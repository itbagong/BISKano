package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommisioningStatusData string
type TypeDeliveryData string

const (
	CommisioningStatusDataPassed CommisioningStatusData = "Passed"
	CommisioningStatusDataFailed CommisioningStatusData = "Failed"
)

const (
	TypeDeliveryDataLand TypeDeliveryData = "Darat"
	TypeDeliveryDataSea  TypeDeliveryData = "Laut"
)

type ContractChecklistGrid struct {
	SalesOrderRefNo string
	SPKNo           string `grid_label:"SPK No"`
	SalesOrderDate  time.Time
	SalesOrderName  string
	CustomerName    string
	Status          string
}

type ContractChecklistForm struct {
	SalesOrderRefNo  string                    `form_section:"General1" form_section_direction:"row" form_label:"Sales Order Ref No." form_lookup:"/sdp/salesorder/find|_id|SalesOrderNo|Name" form_section_size:"2"`
	SalesOrderDate   time.Time                 `form_section:"General1" form_section_direction:"row" form_label:"Sales Order Date"`
	SalesOrderName   string                    `form_section:"General1" form_section_direction:"row" form_label:"Sales Order Name"`
	SPKNo            string                    `form_section:"General1" form_section_direction:"row" form_label:"SPK/PO/Contract No"`
	CustomerID       string                    `form_section:"General2" form_section_direction:"row" form_label:"Customer" form_lookup:"tenant/customer/find|_id|Name"`
	CustomerName     string                    `form_section:"General2" form_section_direction:"row" form_kind:"text" form_label:"Name"`
	CustomerAddress  string                    `form_section:"General2" form_section_direction:"row" form_label:"Address"`
	CustomerCity     string                    `form_section:"General2" form_section_direction:"row" form_label:"City"`
	CustomerProvince string                    `form_section:"General2" form_section_direction:"row" form_label:"Province"`
	CustomerCountry  string                    `form_section:"General2" form_section_direction:"row" form_label:"Country"`
	CustomerZipcode  int                       `form_section:"General2" form_section_direction:"row" form_label:"Zipcode"`
	Dimension        tenantcoremodel.Dimension `form_section:"General3" form_section_direction:"row"`
}

func (o *ContractChecklistForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type ContractChecklistCompletness struct {
	Text     string
	Selected bool
}

type ContractChecklistBastForm1 struct {
	BastNO    string    `form_section:"section1" form_label:"SJ / BAST NO"`
	BastDate  time.Time `form_section:"section1" form_label:"BAST Date" form_kind:"date"`
	UsageDate time.Time `form_section:"section1" form_label:"Usage Date" form_kind:"date"`
}

type ContractChecklistBastForm2 struct {
	LocationSender    string `form_section:"Data Pengiriman" form_label:"Location"`
	NameSender        string `form_section:"Data Pengiriman" form_label:"Name"`
	CompanySender     string `form_section:"Data Pengiriman" form_label:"Company"`
	AddressSender     string `form_section:"Data Pengiriman" form_label:"address"`
	LocationRecipient string `form_section:"Data Penerima" form_label:"Location"`
	NameRecipient     string `form_section:"Data Penerima" form_label:"Name"`
	CompanyRecipient  string `form_section:"Data Penerima" form_label:"Company"`
	AddressRecipient  string `form_section:"Data Penerima" form_label:"Address"`
	PhoneRecipient    string `form_section:"Data Penerima" form_label:"Phone"`
}

func (o *ContractChecklistBastForm2) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Data Pengiriman", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Data Penerima", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type ContractChecklistBastForm3 struct {
	ItemModel      string  `form_section:"SHIPMENT INFORMATION / DATA UNIT / BARANG" form_label:"Model / Jenis Barang"`
	PoliceNumber   string  `form_section:"SHIPMENT INFORMATION / DATA UNIT / BARANG" form_label:"Nomor Polisi"`
	FrameNumber    string  `form_section:"SHIPMENT INFORMATION / DATA UNIT / BARANG" form_label:"Nomor Rangka"`
	MachineNumber  string  `form_section:"SHIPMENT INFORMATION / DATA UNIT / BARANG" form_label:"Nomor Mesin"`
	ProductionYear int     `form_section:"SHIPMENT INFORMATION / DATA UNIT / BARANG" form_label:"Production Year"`
	SMU            int     `form_section:"Status" form_label:"SMU/Km"`
	City           string  `form_section:"Status" form_label:"City"`
	FuelStatus     float32 `form_section:"Status" form_label:"Fuel Status (%)"`
	ModaPengiriman string  `form_section:"Status" form_label:"Moda Pengiriman" form_items:"Darat|Laut"`
}

func (o *ContractChecklistBastForm3) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "SHIPMENT INFORMATION / DATA UNIT / BARANG", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Status", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type ContractChecklistBastForm5 struct {
	DeliveryDetail string `form_section:"section1" form_multi_row:"2" form_label:"Delivery Detail"`
	ReceiptDetail  string `form_section:"section1" form_multi_row:"2" form_label:"Receipt Detail"`
	Notes          string `form_section:"section1" form_multi_row:"2" form_label:"Notes"`
	Notes2         string `form_section:"section1" form_multi_row:"3" form_label:"Catatan"`
}

type ContractChecklistBast struct {
	BastNO            string           `form_section:"section1" form_label:"SJ / BAST NO"`
	BastDate          time.Time        `form_section:"section1" form_label:"BAST Date"`
	UsageDate         time.Time        `form_section:"section1" form_label:"Usage Date"`
	LocationSender    string           `form_section:"section1" form_label:"Location"`
	NameSender        string           `form_section:"section1" form_label:"Name"`
	CompanySender     string           `form_section:"section1" form_label:"Company"`
	AddressSender     string           `form_section:"section1" form_label:"address"`
	LocationRecipient string           `form_section:"section1" form_label:"Location"`
	NameRecipient     string           `form_section:"section1" form_label:"Name"`
	CompanyRecipient  string           `form_section:"section1" form_label:"Company"`
	AddressRecipient  string           `form_section:"section1" form_label:"Address"`
	PhoneRecipient    string           `form_section:"section1" form_label:"Phone"`
	ItemModel         string           `form_section:"section1" form_label:"Model / Jenis Barang"`
	PoliceNumber      string           `form_section:"section1" form_label:"Nomor Polisi"`
	FrameNumber       string           `form_section:"section1" form_label:"Nomor Rangka"`
	MachineNumber     string           `form_section:"section1" form_label:"Nomor Mesin"`
	ProductionYear    int              `form_section:"section1" form_label:"Production Year"`
	SMU               int              `form_section:"section1" form_label:"SMU/Km"`
	City              string           `form_section:"section1" form_label:"City"`
	FuelStatus        float32          `form_section:"section1" form_label:"Fuel Status (%)"`
	TypeDelivery      TypeDeliveryData `form_section:"section1" form_label:"Moda Pengiriman"`
	DeliveryDetail    string           `form_section:"section1" form_label:"Delivery Detail"`
	ReceiptDetail     string           `form_section:"section1" form_label:"Receipt Detail"`
	Notes             string           `form_section:"section1" form_label:"Notes"`
	Notes2            string           `form_section:"section1" form_label:"Catatan"`
}

type ChecklistWorkOrderSCM struct {
	UnitType           string
	SunID              string
	AssetID            string
	AssetName          string
	UnitStatus         string
	DocumentStatus     string
	DeliveryStatus     string
	CommisioningDate   time.Time
	CommisioningResult string
	CommisioningStatus CommisioningStatusData
	Bast               ContractChecklistBast
}

type ContractChecklist struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" key:"1" form_read_only_edit:"1" form:"hide"`
	SalesOrderRefNo   string
	SalesOrderDate    time.Time
	SalesOrderName    string
	SPKNo             string
	CustomerID        string
	CustomerName      string
	CustomerAddress   string
	CustomerCity      string
	CustomerProvince  string
	CustomerCountry   string
	CustomerZipcode   int
	Checklist         []ChecklistWorkOrderSCM
	Checklists        tenantcoremodel.Checklists
	Created           time.Time `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info"`
}

func (o *ContractChecklist) TableName() string {
	return "ContractChecklists"
}

func (o *ContractChecklist) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ContractChecklist) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ContractChecklist) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ContractChecklist) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ContractChecklist) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ContractChecklist) PostSave(dbflex.IConnection) error {
	return nil
}
