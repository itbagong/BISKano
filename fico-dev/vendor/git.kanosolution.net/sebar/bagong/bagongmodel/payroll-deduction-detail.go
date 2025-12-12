package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BGPayrollDeductionDetail struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string    `form:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"Detail" form_pos:"1,2"`
	PayrollDeductionID string    `form:"hide" form_pos:"1,2"`
	Round              string    `form_items:"Round Auto|Up|Down" form_pos:"2,2"`
	Decimal            float64   `form_pos:"2,2"`
	IsActive           bool      `form_pos:"3"`
	Created            time.Time `form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_read_only:"1" grid:"hide"  form_section:"Time Info"`
}

func (o *BGPayrollDeductionDetail) TableName() string {
	return "BGPayrollDeductionDetail"
}

func (o *BGPayrollDeductionDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BGPayrollDeductionDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BGPayrollDeductionDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BGPayrollDeductionDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BGPayrollDeductionDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BGPayrollDeductionDetail) PostSave(dbflex.IConnection) error {
	return nil
}
