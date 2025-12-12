package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CashJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                     `bson:"_id" json:"_id" key:"1"  form_read_only:"1"  form_section:"General"   form_section_direction:"row" form_section_size:"3"  grid_sortable:"1"`
	JournalTypeID     string                     `grid:"hide" form_lookup:"/fico/cashjournaltype/find|_id|Name,PostingProfileID,ReferenceTemplateID,TransactionType" form_section:"General"  `
	PostingProfileID  string                     `grid:"hide" form_section:"General" form_read_only:"1"  form_lookup:"/fico/postingprofile/find|_id|Name"`
	TrxDate           time.Time                  `form_kind:"date" form_section:"General" grid_sortable:"1"`
	CashJournalType   string                     `grid:"hide" form_section:"General" form:"hide"`
	CashBookID        string                     `form_section:"General2" form_section_direction:"row" form_lookup:"/tenant/cashbank/find|_id|Name" form_section_size:"3"   form_label:"Cash & Bank"  grid:"hide" `
	Text              string                     `form_section:"General2" form_multi_row:"5" grid_keyword:"1"`
	References        tenantcoremodel.References `form:"hide" grid:"hide" `
	Lines             JournalLines               `form:"hide" grid:"hide" `
	CurrencyID        string                     `form_section:"General3" form:"hide" form_section_direction:"row" form_section_size:"3"  form_read_only:"1"`
	Amount            float64                    `form_section:"General3" form_section_direction:"row" form_section_size:"3"  form_read_only:"1"`
	ReportingAmount   float64                    `form_section:"General3" grid:"hide" form_section_direction:"row" form_section_size:"3"  form_read_only:"1"`
	Status            JournalStatus              `form_section:"General" form_read_only:"1"`
	CompanyID         string                     `form:"hide"  grid:"hide" `
	Dimension         tenantcoremodel.Dimension  `form_section:"Dimension" form_section_direction:"row" form_section_size:"3"  grid:"hide" `
	Created           time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CashJournal) FormSections() []suim.FormSectionGroup {
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
func (o *CashJournal) TableName() string {
	return "CashJournals"
}

func (o *CashJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CashJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CashJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CashJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CashJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CashJournal) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *CashJournal) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "JournalType", Fields: []string{"CompanyID", "JournalTypeID"}},
		{Name: "Dimension", Fields: []string{"CompanyID", "Dimension.Key", "Dimension.Value"}},
	}
}
