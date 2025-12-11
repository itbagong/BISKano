package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetTransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	Asset             tenantcoremodel.Asset
	TrxDate           time.Time
	SourceType        tenantcoremodel.TrxModule
	SourceJournalType string
	SourceJournalID   string
	SourceLineNo      int
	TrxType           string
	Amount            float64
	Text              string
	CompanyID         string
	Status            AmountStatus
	Dimension         tenantcoremodel.Dimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AssetTransaction) TableName() string {
	return "AssetTransactions"
}

func (o *AssetTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
