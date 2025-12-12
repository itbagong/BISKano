package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashBalance struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CashBookID        string     `form_read_only:"1" form_lookup:"/tenant/cashbank/find|_id|_id,Name"`
	BalanceDate       *time.Time `form_kind:"date"`
	Balance           float64
	Reserved          float64
	Planned           float64
	Available         float64
	Closed            bool
	CompanyID         string
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashBalance) TableName() string {
	return "CashBalances"
}

func (o *CashBalance) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashBalance) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashBalance) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashBalance) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashBalance) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	o.Calc()
	return nil
}

func (o *CashBalance) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *CashBalance) Calc() {
	o.Available = o.Balance + o.Planned + o.Reserved
}
