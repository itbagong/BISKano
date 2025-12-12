package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaxCalcMethod string
type TaxInvoiceAmountOperation string
type TaxRounding string

const (
	TaxCalcHeader TaxCalcMethod = "Header"
	TaxCalcLine   TaxCalcMethod = "Line"

	TaxIncreaseAmount TaxInvoiceAmountOperation = "Increase"
	TaxDecreaseAmount TaxInvoiceAmountOperation = "Decrease"

	TaxRoundingUp   TaxRounding = "Up"
	TaxRoundingDown TaxRounding = "Down"
	TaxRoundingAuto TaxRounding = "Auto"
)

type TaxSetup struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string        `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string        `form_required:"1" form_section:"General"`
	TaxGroup          string        `form_lookup:"/fico/taxgroup/find|_id|_id,Name"`
	CalcMethod        TaxCalcMethod `form_items:"Header|Line"`
	IncludeInInvoice  bool
	InvoiceOperation  TaxInvoiceAmountOperation `form_items:"Increase|Decrease" form_label:"Tax Object" label:"Tax Object"`
	Rate              float64
	Rounding          TaxRounding `form_items:"Up|Down|Auto"`
	Decimal           int
	LedgerAccountID   string                    `form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	Dimension         tenantcoremodel.Dimension `grid:"hide"`
	Modules           []string                  `form_items:"Sales|Purchase" form_required:"1" form_label:"Modules" label:"Modules"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TaxSetup) TableName() string {
	return "TaxSetups"
}

func (o *TaxSetup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TaxSetup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TaxSetup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TaxSetup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TaxSetup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TaxSetup) PostSave(dbflex.IConnection) error {
	return nil
}
