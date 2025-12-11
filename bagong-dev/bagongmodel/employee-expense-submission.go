package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeExpenseSubmission struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	SubmissionNo      string
	SubmissionTitle   string
	EmployeeID        string `form_lookup:"/bagong/employeedetail/find|_id|Name"`
	EmployeeName      string
	JournalType       string
	JournalTypeID     string
	Dimension         tenantcoremodel.Dimension
	CompanyID         string `grid:"hide" form:"hide"`
	InclusiveTax      bool
	Detail            []EmployeeExpenseSubmissionDetail
	SubmissionDate    time.Time
	Status            string
	Attachment        []SiteAttachment `form:"hide"`
	Created           time.Time        `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time        `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}
type EmployeeExpenseSubmissionDetail struct {
	Asset         string `grid:"hide" form:"hide"`
	Description   string
	Quantity      float64
	UoM           string
	UnitPrice     float64
	Total         float64
	OffsetAccount string
	CashBank      string
	Critical      string
}

func (o *EmployeeExpenseSubmission) TableName() string {
	return "BGEmployeeExpenseSubmission"
}

func (o *EmployeeExpenseSubmission) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *EmployeeExpenseSubmission) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *EmployeeExpenseSubmission) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *EmployeeExpenseSubmission) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *EmployeeExpenseSubmission) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *EmployeeExpenseSubmission) PostSave(dbflex.IConnection) error {
	return nil
}
