package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CGType string
type CGStatus string

const (
	Cheque CGType = "Cheque"
	Giro   CGType = "Giro"

	CGOpen     CGStatus = "Open"
	CGReserved CGStatus = "Reserved"
	CGCleared  CGStatus = "Cleared"
)

type ChequeGiro struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string     `bson:"_id" json:"_id" key:"1" form_section:"General" form_section_auto_col:"1" form_label:"Cheque/Giro No." form_pos:"3" grid_label:"Cheque/Giro No." grid_pos:"3" `
	CashBookID        string     `form_pos:"1" grid_label:"Bank Account No"  grid_pos:"1"`
	CheckBookID       string     `form:"hide" grid:"hide"`
	Kind              CGType     `form_pos:"2" grid_label:"Type"  grid_pos:"2"`
	Amount            float64    `form_pos:"6" grid_pos:"4"`
	IssueDate         *time.Time `form_kind:"date" form_pos:"4" grid_pos:"5"`
	ReleaseDate       *time.Time `form:"hide" grid:"hide"`
	ClearDate         *time.Time `form_kind:"date" form_pos:"5" grid_pos:"6"`
	Status            CGStatus   `form:"hide" grid_pos:"7"`
	Memo              string     `form_multi_row:"3" grid:"hide"`
	BfcName           string     `grid:"hide"`
	BfcAccount        string     `form:"hide" grid:"hide"`
	BfcBank           string     `form:"hide" grid:"hide"`
	BfcSwift          string     `form:"hide" grid:"hide"`
	CashJournalID     string     `form:"hide" grid:"hide"`
	LineNo            int        `form:"hide" grid:"hide"`
	CompanyID         string     `form:"hide" grid:"hide"`
	Created           time.Time  `form:"hide" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time  `form:"hide" form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ChequeGiro) TableName() string {
	return "ChequeGiros"
}

func (o *ChequeGiro) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ChequeGiro) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ChequeGiro) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ChequeGiro) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ChequeGiro) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ChequeGiro) PostSave(dbflex.IConnection) error {
	return nil
}
