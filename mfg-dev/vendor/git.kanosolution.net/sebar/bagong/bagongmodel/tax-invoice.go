package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaxInvoice struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	FPNo              string    `form_required:"1" form_section_width:"1"`
	Status            string    `form_items:"Open|Allocated|Submitted"`
	Created           time.Time `grid:"hide" form_kind:"datetime" form_read_only:"1" grid_sortable:"1" form:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
}

func (o *TaxInvoice) TableName() string {
	return "BGTaxInvoices"
}

func (o *TaxInvoice) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}

}

func (o *TaxInvoice) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *TaxInvoice) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TaxInvoice) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TaxInvoice) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TaxInvoice) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *TaxInvoice) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "TaxInvoiceIndex", Fields: []string{"TaxNo"}},
	}
}
