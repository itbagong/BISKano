package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JournalType struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_section:"General" form_section_direction:"row" form_section_auto_col:"3"`
	Name              string                    `form_section:"General"`
	TransactionType   string                    `form_use_list:"1" form_section:"General" form_items:"Candidate Resume|Manpower Request|OL & Plotting|Contract|Talent Development - Promotion - General & Benefit|Talent Development - Promotion - Tracking Assessment|Talent Development - Promotion - Tracking SK Acting|Talent Development - Promotion - Tracking SK Tetap|Talent Development - Rotation|Talent Development - Demotion|Talent Development - Salary Change|Talent Development - POH Change|Overtime|Loan|Work Termination - Resign|Work Termination - PHK|Work Termination - Sakit Berkepanjangan|Work Termination - Meninggal|Work Termination - Pensiun|Leave Compensation|Business Trip|Coaching & Violation - SP1|Coaching & Violation - SP2|Coaching & Violation - SP3|Coaching & Violation - Surat Teguran Tertulis|Coaching & Violation - Surat Panggilan Masuk Kerja 1|Coaching & Violation - Surat Panggilan Masuk Kerja 2|Coaching & Violation - Form Coaching"`
	PostingProfileID  string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	ReferenceTemplate string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name" form_section:"General2"`
	ChecklistTemplate string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name" form_section:"General2"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *JournalType) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}
func (o *JournalType) TableName() string {
	return "HCMJournalTypes"
}

func (o *JournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *JournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *JournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *JournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *JournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *JournalType) PostSave(dbflex.IConnection) error {
	return nil
}
