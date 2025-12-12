package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LoanStatus string

const (
	LoanPaid   LoanStatus = "Paid"
	LoanUnpaid LoanStatus = "Unpaid"
)

type LoanSetup struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	EmployeeID         string    `form_lookup:"/tenant/employee/find|_id|Name" grid:"hide"`
	LoanDate           time.Time `form_kind:"date"`
	Period             int
	LoanAmount         float64
	AutodebetStartDate time.Time  `form_kind:"date"`
	Lines              []LoanLine `grid:"hide" form:"hide"`
	Created            time.Time  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type LoanLine struct {
	Date              time.Time `form_kind:"date"`
	Period            int
	InstallmentAmount float64
	RemainingLoan     float64
	Status            LoanStatus
}

func (o *LoanSetup) TableName() string {
	return "LoanSetups"
}

func (o *LoanSetup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LoanSetup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LoanSetup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LoanSetup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LoanSetup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LoanSetup) PostSave(dbflex.IConnection) error {
	return nil
}
