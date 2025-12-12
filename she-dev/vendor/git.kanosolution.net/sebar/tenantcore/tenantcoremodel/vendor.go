package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vendor struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name                  string          `form_required:"1" form_section:"General"`
	GroupID               string          `form_lookup:"/tenant/vendorgroup/find|_id|_id,Name" form_required:"1"`
	PostingProfileID      string          `form_lookup:"/fico/postingprofile/find|_id|_id,Name" grid:"hide" form:"hide"`
	MainBalanceAccount    string          `form_lookup:"/tenant/ledgeraccount/coa/find|_id|_id,Name"`
	DepositAccount        string          `form_lookup:"/tenant/ledgeraccount/coa/find|_id|_id,Name"`
	PaymentTermID         string          `form_lookup:"/fico/paymentterm/find|_id|Name"`
	TaxType               string          `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=TTY|_id|Name"`
	TaxRegistrationNumber string          `form_section:"Tax" grid:"hide" form_section_auto_col:"2" form_section_show_title:"1"`
	TaxName               string          `form_section:"Tax" grid:"hide"`
	TaxAddress            string          `form_section:"Tax" grid:"hide"`
	TaxRegistrationDate   time.Time       `form_section:"Tax" grid:"hide" form_kind:"date"`
	Setting               CustomerSetting `grid:"hide" form:"hide"`
	Sites                 []string        `form_lookup:"/bagong/sitesetup/find|_id|Name"`
	Dimension             Dimension       `grid:"hide"`
	IsActive              bool
	Created               time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Vendor) TableName() string {
	return "Vendors"
}

func (o *Vendor) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Vendor) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Vendor) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Vendor) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Vendor) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Vendor) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Vendor) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "GroupIndex", Fields: []string{"GroupID"}},
	}
}
