package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashReconStatus string

const (
	CashReconStatusDraft     CashReconStatus = "DRAFT"
	CashReconStatusCompleted CashReconStatus = "COMPLETED"
)

type CashReconGrid struct {
	ID           string
	SourceType   string
	Amount       float64
	TrxType      string
	VoucherNo    string
	ChequeGiroID string
	TrxDate      time.Time
}
type CashRecon struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name               string `form_required:"1" form_section:"General"`
	CashBankID         string
	ReconDate          time.Time
	PreviousReconDate  *time.Time
	PreviousBalance    float64
	ReconBalance       float64
	ReconJournalTypeID string
	Lines              JournalLines
	Diff               float64
	CompanyID          string
	Status             CashReconStatus
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashRecon) TableName() string {
	return "CashRecons"
}

func (o *CashRecon) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashRecon) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashRecon) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashRecon) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashRecon) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashRecon) PostSave(dbflex.IConnection) error {
	return nil
}
