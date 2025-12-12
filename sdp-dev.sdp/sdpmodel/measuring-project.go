package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MeasuringProject struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_read_only_edit:"1" form_section:"General1" form_read_only:"1" form_section_size:"4"`
	SoRefNo           string `form_required:"1" form_section:"General1" label:"SO Ref No." grid:"hide" form:"hide"`
	ProjectID         string `form_required:"1" form_section:"General1" label:"Project ID"`
	ProjectName       string `form_required:"1" form_section:"General1" label:"Project Name"`
	ProjectAlias      string `form_section:"General1" label:"Project Alias"`
	ProjectType       string `form_section:"General1" label:"Project Type" grid:"hide" form:"hide"`
	SalesPriceBook    string `form_section:"General1" label:"Sales Price Book" form_lookup:"/sdp//salespricebook/find|ID|Name" grid:"hide" form:"hide"`

	CustomerID       string    `form_required:"1" form_section:"General2" label:"Customer Name" grid:"hide" form:"hide"`
	ProjectPeriod    int       `form_required:"1" form_section:"General2" label:"Project Period" form_kind:"number"`
	StartPeriodMonth time.Time `form_required:"1" form_section:"General2" label:"Start Period Month" form_kind:"date" grid:"hide"`
	EndPeriodMonth   time.Time `form_required:"1" form_section:"General2" label:"End Period Month" form_kind:"date" grid:"hide"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"General4"`

	RevenueEstimation float64 `form:"hide"`
	ExpenseEstimation float64 `form:"hide"`

	Lines []LinesMeasuringProject `grid:"hide" form:"hide"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide"`
}

func (o *MeasuringProject) TableName() string {
	return "MeasuringProject"
}

func (o *MeasuringProject) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MeasuringProject) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MeasuringProject) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MeasuringProject) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MeasuringProject) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()

		// StartPeriodMonth := o.StartPeriodMonth
		// EndPeriodMonth := o.EndPeriodMonth

		// // Perlu di tambah loop ke ledger account
		// for i := o.StartPeriodMonth.Year(); i <= o.EndPeriodMonth.Year(); i++ {
		// 	Line := LinesMeasuringProject{}
		// 	Line.Budget = "Revenue"
		// 	Line.Year = i
		// 	Line.LedgerAccount = "TEST"

		// 	var arrMonths map[string]float64
		// 	arrMonths = map[string]float64{}
		// 	for EndPeriodMonth.After(StartPeriodMonth) {
		// 		sMonth := time.Month(StartPeriodMonth.Month()).String()
		// 		arrMonths[sMonth] = 0.0

		// 		StartPeriodMonth = StartPeriodMonth.AddDate(0, 1, 0)
		// 		if int(StartPeriodMonth.Month()) == 1 {
		// 			break
		// 		}
		// 	}

		// 	Line.Month = arrMonths
		// 	o.Lines = append(o.Lines, Line)
		// }
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()

	return nil
}

func (o *MeasuringProject) PostSave(dbflex.IConnection) error {
	return nil
}
