package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventTrx struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CompanyID         string
	Item              tenantcoremodel.Item
	SKU               string `label:"SKU"`
	InventDim         InventDimension
	Dimension         tenantcoremodel.Dimension
	TrxDate           time.Time
	FinancialDate     *time.Time
	Status            ItemBalanceStatus
	Qty               float64
	TrxQty            float64
	TrxUnitID         string
	AmountPhysical    float64
	AmountFinancial   float64
	AmountAdjustment  float64
	SourceType        tenantcoremodel.TrxModule
	SourceJournalID   string
	SourceTrxType     string
	SourceLineNo      int
	References        tenantcoremodel.References
	Text              string
	VoucherNo         string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type InventTrxPerDimension struct {
	TrxDate          time.Time
	SourceType       string
	SourceTrxType    string
	SourceJournalID  string
	SourceLineNo     int
	Status           ItemBalanceStatus
	Qty              float64
	TrxQty           float64
	QtyConfirmed     float64 `grid:"hide"`
	QtyReserved      float64 `grid:"hide"`
	QtyPlanned       float64 `grid:"hide"`
	AmountPhysical   float64
	AmountFinancial  float64
	AmountAdjustment float64
	Amount           float64
}

type TrxBalance struct {
	Balance float64
	Date    *time.Time
}

func (o *InventTrx) TableName() string {
	return "InventTransactions"
}

func (o *InventTrx) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventTrx) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventTrx) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventTrx) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventTrx) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Status == ItemBalanceStatus("") {
		if o.FinancialDate != nil && o.FinancialDate.Year() > 1980 {
			o.Status = ItemConfirmed
		} else {
			if o.Qty < 0 {
				o.Status = ItemReserved
			} else {
				o.Status = ItemPlanned
			}
		}
	}

	o.InventDim = *o.InventDim.Calc()
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventTrx) PostSave(dbflex.IConnection) error {
	return nil
}
