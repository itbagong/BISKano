package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrxMetadata struct {
	Module      TrxModule `json:",omitempty" bson:",omitempty"`
	JournalID   string    `json:",omitempty" bson:",omitempty"`
	LineNo      int       `json:",omitempty" bson:",omitempty"`
	JournalType string    `json:",omitempty" bson:",omitempty"`
	TrxType     TrxType   `json:",omitempty" bson:",omitempty"`
	VoucherNum  string    `json:",omitempty" bson:",omitempty"`
	RecordID    string    `json:",omitempty" bson:",omitempty"`
}

type TrxApply struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	CompanyID         string
	Source            TrxMetadata
	ApplyTo           TrxMetadata
	Amount            float64
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TrxApply) TableName() string {
	return "TrxApplies"
}

func (o *TrxApply) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrxApply) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrxApply) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrxApply) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrxApply) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrxApply) PostSave(dbflex.IConnection) error {
	return nil
}
