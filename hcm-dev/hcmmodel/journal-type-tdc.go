package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TDCJournalType struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_section:"General" form_section_direction:"row" form_section_auto_col:"3"`
	Name              string                    `form_section:"General"`
	TransactionType   string                    `form_use_list:"1" form_section:"General" form_items:"Training - General & Participant|Training - Training Detail & Attachment"`
	PostingProfileID  string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name" form_section:"General2"`
	ReferenceTemplate string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name" form_section:"General2"`
	ChecklistTemplate string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name" form_section:"General2"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TDCJournalType) FormSections() []suim.FormSectionGroup {
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
func (o *TDCJournalType) TableName() string {
	return "TDCJournalTypes"
}

func (o *TDCJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TDCJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TDCJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TDCJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TDCJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TDCJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
