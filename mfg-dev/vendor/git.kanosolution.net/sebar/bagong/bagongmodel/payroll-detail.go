package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BGPayrollDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Period            string
	SiteID            string
	EmployeeID        string
	Name              string
	BaseSalary        float64
	Benefits          []DetailBenDeduct
	Deductions        []DetailBenDeduct
	Dimension         tenantcoremodel.Dimension
	AttendanceNum     int
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}
type DetailBenDeduct struct {
	Name     string
	ID       string
	Amount   float64
	IsManual bool
}

func (o *BGPayrollDetail) GetTakeHome() float64 {
	total := o.BaseSalary
	for _, hh := range o.Benefits {
		total += hh.Amount
	}
	for _, hh := range o.Deductions {
		total -= hh.Amount
	}
	return total
}
func (o *BGPayrollDetail) CheckBenefitExists(benefitId string) bool {
	for _, val := range o.Benefits {
		if val.ID == benefitId {
			return true
		}
	}
	return false
}
func (o *BGPayrollDetail) CheckDeductionExists(deductionId string) bool {
	for _, val := range o.Deductions {
		if val.ID == deductionId {
			return true
		}
	}
	return false
}
func (o *BGPayrollDetail) SetBenefitValue(benefitId, name string, value float64, calcType ficomodel.PayrollComponentType) {
	for idx, val := range o.Benefits {
		if val.ID == benefitId {
			o.Benefits[idx].Amount = value
			return
		}
	}

	isManual := false
	if calcType == ficomodel.PayrollComponentFixed {
		isManual = true
	}

	jj := DetailBenDeduct{ID: benefitId, Name: name, Amount: value, IsManual: isManual}
	o.Benefits = append(o.Benefits, jj)
}
func (o *BGPayrollDetail) SetDeductionValue(deductionId, name string, value float64, calcType ficomodel.PayrollComponentType) {
	for idx, val := range o.Deductions {
		if val.ID == deductionId {
			o.Deductions[idx].Amount = value
			return
		}
	}

	isManual := false
	if calcType == ficomodel.PayrollComponentFixed {
		isManual = true
	}

	jj := DetailBenDeduct{ID: deductionId, Name: name, Amount: value, IsManual: isManual}
	o.Deductions = append(o.Deductions, jj)
}
func (o *BGPayrollDetail) TableName() string {
	return "BGPayrollDetail"
}

func (o *BGPayrollDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BGPayrollDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BGPayrollDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BGPayrollDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BGPayrollDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BGPayrollDetail) PostSave(dbflex.IConnection) error {
	return nil
}
