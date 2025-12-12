package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashBank struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid_keyword:"1"`
	Name               string    `form_required:"1" form_section:"General" grid_keyword:"1"`
	BankName           string    `form_required:"1" form_section:"General"`
	BankAccountNo      string    `form_section:"General"`
	BankAccountName    string    `form_section:"General"`
	CashBankGroupID    string    `form_lookup:"/tenant/cashbankgroup/find|_id|Name"`
	CurrencyID         string    `form_lookup:"/tenant/currency/find|_id|Name"`
	MinimumBalance     float64   `grid:"hide"`
	MaximumBalance     float64   `grid:"hide"`
	MainBalanceAccount string    `form_lookup:"/tenant/ledgeraccount/coa/find|_id|_id,Name"`
	Dimension          Dimension `grid:"hide"`
	AllowNegative      bool
	IsActive           bool
	Available          float64   `form:"hide"`
	Planned            float64   `form:"hide"`
	Reserved           float64   `form:"hide"`
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashBank) TableName() string {
	return "CashBanks"
}

func (o *CashBank) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashBank) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashBank) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashBank) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashBank) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashBank) PostSave(dbflex.IConnection) error {
	return nil
}
