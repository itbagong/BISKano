package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LedgerJournal struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	CompanyID           string                     `form_lookup:"/tenant/company/find|_id|Name" grid:"hide" form:"hide" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3"`
	ID                  string                     `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" grid_sortable:"1"`
	JournalTypeID       string                     `form_lookup:"/fico/ledgerjournaltype/find|_id|_id,Name" form_section:"General" form_required:"1"`
	TrxDate             time.Time                  `form_kind:"date" form_section:"General" grid_sortable:"1"`
	Text                string                     `form_required:"1" form_section:"General2" form_multi_row:"5"  form_section_direction:"row" form_section_size:"3"`
	EmployeeID          string                     `label:"Employee" form_lookup:"/tenant/employee/find|_id|Name" form:"hide" grid:"hide"`
	DefaultOffset       SubledgerAccount           `grid:"hide" form:"hide" form_section:"General2"`
	ReferenceTemplateID string                     `grid:"hide" form:"hide" form_lookup:"/tenant/referencetemplate/find|_id|_id,Name" form_section:"General2"`
	ChecklistTemplateID string                     `grid:"hide" form:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|_id,Name" form_section:"General2"`
	References          tenantcoremodel.References `grid:"hide" form:"hide" form_section:"General3" form_section_direction:"row" form_section_size:"3"`
	Checklists          tenantcoremodel.Checklists `grid:"hide" form:"hide" form_section:"General3"`
	Lines               []JournalLine              `form:"hide" grid:"hide" form_section:"General3"`
	PostingProfileID    string                     `form:"hide" grid:"hide"`
	Status              JournalStatus              `form_read_only:"1" form_section:"General"`
	Dimension           tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created             time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate          time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LedgerJournal) FormSections() []suim.FormSectionGroup {
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

func (o *LedgerJournal) TableName() string {
	return "LedgerJournalHeaders"
}

func (o *LedgerJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LedgerJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LedgerJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LedgerJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LedgerJournal) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LedgerJournal) PostSave(dbflex.IConnection) error {
	return nil
}
