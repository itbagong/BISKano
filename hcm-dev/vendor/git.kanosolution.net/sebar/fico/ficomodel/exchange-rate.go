package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExchangeRate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	GroupID           string
	FromCurrencyID    string
	ToCurrencyID      string
	ExchDate          time.Time
	Rate              float64
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ExchangeRate) TableName() string {
	return "ExchangeRates"
}

func (o *ExchangeRate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ExchangeRate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ExchangeRate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ExchangeRate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ExchangeRate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ExchangeRate) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *ExchangeRate) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "GroupFromToIndex", Fields: []string{"GroupID", "FromCurrencyID", "ToCurrencyID"}},
		{Name: "GroupDateFromToIndex", Fields: []string{"GroupID", "-ExchDate", "FromCurrencyID", "ToCurrencyID"}},
	}
}
