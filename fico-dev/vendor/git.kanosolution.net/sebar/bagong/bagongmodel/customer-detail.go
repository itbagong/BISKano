package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerDetail struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string            `form:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"Biodata" form_pos:"1"`
	Address                string            `form_section:"Address" label:"Billing Address" form_pos:"1" form_multi_row:"2"`
	City                   string            `form_section:"Address" label:"Billing City" form_pos:"2" form_multi_row:"2" `
	Province               string            `form_section:"Address" label:"Billing Province" form_pos:"3,1"`
	Country                string            `form_section:"Address" label:"Billing Country" form_pos:"3,2"`
	Phone                  string            `form_section:"Address" label:"Billing Phone" form_pos:"4,1"`
	Zipcode                string            `form_section:"Address" label:"Billing Zipcode" form_pos:"4,2"`
	SameAsBillAddr         bool              `form_section:"Address" label:"Same as Billing address" form_pos:"5"`
	DeliveryAddress        string            `form_section:"Address" label:"Delivery Address" form_pos:"6" form_multi_row:"2"`
	DeliveryCity           string            `form_section:"Address" label:"Delivery City" form_pos:"7" form_multi_row:"2" `
	DeliveryProvince       string            `form_section:"Address" label:"Delivery Province" form_pos:"8,1"`
	DeliveryCountry        string            `form_section:"Address" label:"Delivery Country" form_pos:"8,2"`
	DeliveryPhone          string            `form_section:"Address" label:"Delivery Phone" form_pos:"9,1"`
	DeliveryZipcode        string            `form_section:"Address" label:"Delivery Zipcode" form_pos:"9,2"`
	CustomerID             string            `form:"hide" form_section:"Biodata"`
	CustomerNo             string            `form:"hide" form_section:"Biodata" label:"Customer No." form_section_size:"4" form_section_show_title:"1"  form_pos:"2"`
	NPWPNo                 string            `form_section:"Biodata" label:"NPWP No." form_pos:"3"`
	NPWPName               string            `form_section:"Biodata" label:"NPWP Name" form_pos:"4"`
	SameAsBillAddress      bool              `form_section:"Biodata" label:"Same as Billing Address" form_pos:"5,2"`
	SameAsDeliverAddress   bool              `form_section:"Biodata" label:"Same as Delivery Address" form_pos:"5,2"`
	NPWPAddress            string            `form_section:"Biodata" label:"NPWP Address" form_pos:"6" form_multi_row:"2"`
	NPWPCity               string            `form_section:"Biodata" label:"NPWP City" form_pos:"7" form_multi_row:"2"`
	NPWPProvince           string            `form_section:"Biodata" label:"NPWP Province" form_pos:"8,2"`
	NPWPCountry            string            `form_section:"Biodata" label:"NPWP Country" form_pos:"8,2"`
	NPWPPhone              string            `form_section:"Biodata" label:"NPWP Phone" form_pos:"9,2"`
	NPWPZipcode            string            `form_section:"Biodata" label:"NPWP Zipcode" form_pos:"9,2"`
	NIK                    string            `form_section:"Biodata" label:"NIK" form_pos:"10"`
	NPPKP                  string            `form_section:"Biodata" label:"NPPKP" form_pos:"11"`
	TaxAddress             string            `form:"hide" form_section:"Biodata" label:"Tax Address" form_pos:"12" form_multi_row:"2"`
	PersonalContact        string            `form_section:"Biodata" label:"Personal Contact" form_pos:"13,2"`
	Email                  string            `form_section:"Biodata" label:"Email" form_pos:"13,2"`
	BusinessPhoneNo        string            `form_section:"Biodata" label:"Business Phone No." form_pos:"14,2"`
	MobilePhoneNo          string            `form_section:"Biodata" label:"Mobile Phone No." form_pos:"14,2"`
	WebURL                 string            `form_section:"Biodata" label:"Web URL" form_pos:"15,2"`
	Fax                    string            `form_section:"Biodata" form_pos:"15,2"`
	Termin                 string            `form_section:"Termin" label:"Payment Term" form_lookup:"/fico/paymentterm/find|_id|Name" form_section_size:"4" form_section_show_title:"1"`
	DebtLimitNoInvoiceDays string            `form_section:"Termin" label:"Debt Limit Max (Days)" form_section:"Tax"`
	DebtLimitMax           string            `form_section:"Termin" label:"Debt Limit Max (Amount)"  `
	Currency               string            `form:"hide" form_section:"Termin" label:"Currency" form_lookup:"/tenant/currency/find|_id|Name"`
	BalanceStart           float64           `form_section:"Termin" label:"Balance Beginning" form:"hide"`
	Msg                    string            `form_section:"Termin" label:"Message"  `
	IsReportPrinted        bool              `form_section:"Termin" label:"Is Report Printed"  `
	Tax1                   string            `form_section:"Selling" label:"Tax 1" form_section_size:"4" form_section_show_title:"1" form_lookup:"/fico/taxsetup/find|_id|Name"`
	Tax2                   string            `form_section:"Selling" label:"Tax 2" form_lookup:"/fico/taxsetup/find|_id|Name"`
	IsTaxIncluded          bool              `form_section:"Selling" label:"Is Tax Included"  `
	TaxType                string            `form_section:"Selling" label:"Tax Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=TTY|_id|Name"`
	TaxCodes               []string          `form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1" form_section:"Selling" form:"hide"`
	CustomerType           string            `form:"hide" form_section:"Selling" label:"Customer Type"  `
	SellerPriceLevel       string            `form:"hide" form_section:"Selling" label:"Seller Price Level"  `
	DefaultPriceBook       []string          `form_section:"Selling" label:"Default Price Book" form:"hide"  `
	DefaultSellingDiscount float64           `form:"hide" form_section:"Selling" label:"Default Selling Discount"  `
	Contacts               []CustomerContact `form:"hide"  form_section:"Selling"` // bingung UI nya gimana, kalo dimunculin malah error
	Note                   string            `form_section:"Note" label:"Note" form_section_size:"4" form_section_show_title:"1" form_multi_row:"2"`
	AdditionalNote         string            `form:"hide" form_section:"Note" label:"Additional Note" form_multi_row:"2"`
	Created                time.Time         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"  `
	LastUpdate             time.Time         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CustomerDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Address", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Biodata", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Termin", ShowTitle: true, AutoCol: 1},
			{Title: "Note", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Selling", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type CustomerContact struct {
	Name      string
	FirstName string
	JobTitle  string
	Phone     string
}

func (o *CustomerDetail) TableName() string {
	return "BGCustomerDetails"
}

func (o *CustomerDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomerDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomerDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomerDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomerDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomerDetail) PostSave(dbflex.IConnection) error {
	return nil
}
