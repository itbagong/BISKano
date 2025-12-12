package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LedgerAccountType string
type LedgerDirection string

const (
	ProfitLossAccount   LedgerAccountType = "PL"
	BalanceSheetAccount LedgerAccountType = "BS"
	HeaderAccount       LedgerAccountType = "HEADER"
	TotalAccount        LedgerAccountType = "TOTAL"

	Credit LedgerDirection = "CREDIT"
	Debit  LedgerDirection = "DEBIT"
)

type LedgerTotalLine struct {
	FromLedgerAccount string
	ToLegderAccount   string
}

type LedgerAccount struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string            `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name              string            `form_section:"General2" form_required:"1"`
	AccountType       LedgerAccountType `form_section:"General3" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=ACT|_id|_id,Name"`
	Status            string            `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SCOA|_id|_id,Name"`
	AccountNature     LedgerDirection   `form_section:"General2" form_items:"Credit|Debit"`
	Lines             []LedgerTotalLine `grid:"hide" form:"hide"`
	BlockedFromGJ     bool              `form_section:"General3" form_label:"Blocked from General Journal" grid_label:"Blocked"`
	IsActive          bool              `form_section:"General"`
	Dimension         Dimension         `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LedgerAccount) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *LedgerAccount) TableName() string {
	return "LedgerAccounts"
}

func (o *LedgerAccount) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LedgerAccount) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LedgerAccount) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LedgerAccount) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LedgerAccount) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LedgerAccount) PostSave(dbflex.IConnection) error {
	return nil
}
