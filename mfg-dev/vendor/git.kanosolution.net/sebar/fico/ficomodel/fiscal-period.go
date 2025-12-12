package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PeriodStatus string

const (
	PeriodBlocked PeriodStatus = "BLOCKED"
	PeriodOpen    PeriodStatus = "OPEN"
	PeriodClose   PeriodStatus = "CLOSED"
)

type FiscalPeriod struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                  `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" grid:"hide" form:"hide" form_section:"General" form_section_auto_col:"2"`
	Name              string                  `form_required:"1" form_section:"General"`
	FiscalYearID      string                  `form_required:"1" form_section:"General" grid:"hide" form:"hide"`
	FromDate          time.Time               `form_required:"1" form_section:"Period" form_section_auto_col:"2" form_kind:"date" form_date_format:"DD MMM YYYY"`
	ToDate            time.Time               `form_required:"1" form_section:"Period" form_section_auto_col:"2" form_kind:"date" form_date_format:"DD MMM YYYY"`
	Modules           map[string]PeriodStatus `form_required:"1" form_section:"Status" form_items:"BLOCKED|OPEN|CLOSED"`
	CompanyID         string                  `grid:"hide" form:"hide"`
	Created           time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time               `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *FiscalPeriod) TableName() string {
	return "FiscalPeriods"
}

func (o *FiscalPeriod) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *FiscalPeriod) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *FiscalPeriod) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FiscalPeriod) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FiscalPeriod) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *FiscalPeriod) PostSave(dbflex.IConnection) error {
	return nil
}
