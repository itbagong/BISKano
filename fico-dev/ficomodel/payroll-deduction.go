package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayrollDeduction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string               `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string               `form_required:"1" form_section:"General"`
	CalcType          PayrollComponentType `form_items:"Manual|Percentage|Value"`
	Value             float64
	CustomEndPoint    string                    // diisi topic untuk custom calculation
	LedgerAccountID   string                    `form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	Dimension         tenantcoremodel.Dimension `grid:"hide"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PayrollDeduction) TableName() string {
	return "PayrollDeductions"
}

func (o *PayrollDeduction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PayrollDeduction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PayrollDeduction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PayrollDeduction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PayrollDeduction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PayrollDeduction) PostSave(dbflex.IConnection) error {
	return nil
}
