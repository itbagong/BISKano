package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type VendorTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"  grid_sortable:"1"`
	CompanyID         string
	SourceType        string
	SourceJournalID   string
	SourceJournalType string
	SourceLineNo      int
	SourceTrxType     string
	TrxDate           time.Time `grid_sortable:"1"`
	Vendor            tenantcoremodel.Vendor
	CurrencyID        string
	Amount            float64
	Text              string
	VoucherNo         string
	Status            AmountStatus
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorTransaction) TableName() string {
	return "VendorTransactions"
}

func (o *VendorTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *VendorTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
