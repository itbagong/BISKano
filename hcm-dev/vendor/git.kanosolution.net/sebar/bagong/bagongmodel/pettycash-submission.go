package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PettyCashSubmission struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	SubmissionNo      string
	SubmissionTitle   string
	JurnalType        string
	SubmissionDate    time.Time
	FromCashBankID    string `grid_label:"From Cash Bank" form_lookup:"/tenant/cashbank/find|_id|Name" grid:"hide"`
	ToCashBankID      string `grid_label:"To Cash Bank" form_lookup:"/tenant/cashbank/find|_id|Name" grid:"hide"`
	Dimension         tenantcoremodel.Dimension
	CompanyID         string `grid:"hide" form:"hide"`
	InclusiveTax      bool
	Detail            []PettyCashSubmissionDetail
	TotalRequested    float64
	ApprovedAmount    float64
	Attachment        []SiteAttachment `form:"hide"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}
type PettyCashSubmissionDetail struct {
	Asset          string
	Description    string
	Quantity       float64
	UoM            string
	UnitPrice      float64
	Total          float64
	ApprovedAmount float64
	Urgent         string
	Ledger         string
	CashBank       string `grid:"hide" form:"hide"`
	Critical       string
}

func (o *PettyCashSubmission) TableName() string {
	return "BGPettyCashSubmission"
}

func (o *PettyCashSubmission) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PettyCashSubmission) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PettyCashSubmission) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PettyCashSubmission) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PettyCashSubmission) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PettyCashSubmission) PostSave(dbflex.IConnection) error {
	return nil
}
