package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BGVendor struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"VendorForm1" form_section_size:"3"`
	TenantCoreVendorID    string          `form_section:"VendorForm1"`
	VendorNumber          string          `form_section:"VendorForm1"`
	VendorCategory        string          `form_section:"VendorForm1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=VNC|_id|_id,Name" form:"hide"`
	VendorAddress         string          `form_section:"VendorForm1"`
	City                  string          `form_section:"VendorForm1"`
	Province              string          `form_section:"VendorForm1"`
	PostalCode            string          `form_section:"VendorForm2"`
	Country               string          `form_section:"VendorForm2"`
	VendorTelephone       string          `form_section:"VendorForm2"`
	VendorPersonalContact string          `form_section:"VendorForm2"`
	VendorEmail           string          `form_section:"VendorForm2"`
	VendorWebsite         string          `form_section:"VendorForm2"`
	TaxCodes              []string        `form:"hide" grid:"hide" form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1" form_section:"VendorForm2"`
	Terms                 VendorTerm      `form:"hide" grid:"hide"`
	VendorContacts        []VendorContact `form:"hide" grid:"hide"`
	VendorBank            []VendorBank    `form:"hide" grid:"hide"`
	Notes                 string          `grid:"hide" form_section:"VendorForm3"  form_multi_row:"5"`
	Created               time.Time       `form_read_only:"1" grid:"hide" form_section:"VendorForm3"`
	LastUpdate            time.Time       `form_read_only:"1" grid:"hide"  form_section:"VendorForm3"`
}

type VendorTerm struct {
	Name               string  `form_section:"Term" form_section_size:"2" form_pos:"1"`
	CurrencyID         string  `form:"hide" form_section:"Term" form_lookup:"/tenant/currency/find|_id|Name" form_pos:"2"`
	BeginningBalance   float64 `form_section:"Term" form_pos:"3,1"`
	BalanceByDate      string  `form_section:"Term" form_kind:"date" form_pos:"3,2"`
	DefaultDescription string  `form_section:"Term" form_multi_row:"5"`
	Taxes1             string  `form_section:"Tax" label:"Tax 1" form_lookup:"/fico/taxsetup/find|_id|Name"`
	Taxes2             string  `form_section:"Tax" label:"Tax 2" form_lookup:"/fico/taxsetup/find|_id|Name"`
	BuiltinTaxInvoice  bool    `form_section:"Tax" label:"built-in tax invoice"`
	PKPNumber          string  `form_section:"Tax"`
	TaxType            string  `form_section:"Tax"`
}

type VendorContact struct {
	Name      string
	FirstName string
	JobTitle  string
	Telephone string
}

type VendorBank struct {
	ID              string `bson:"_id" json:"_id"`
	BankName        string
	BankAccountNo   string
	BankAccountName string
	Branch          string
	SwiftCode       string
}

func (o *BGVendor) TableName() string {
	return "BGVendors"
}

func (o *BGVendor) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BGVendor) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BGVendor) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BGVendor) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BGVendor) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BGVendor) PostSave(dbflex.IConnection) error {
	return nil
}
