package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerJournalTypeConfiguration struct {
	orm.DataModelBase     `bson:"-" json:"-"`
	ID                    string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	CustomerJournalTypeID string    `form:"hide"`
	DividerType           string    `form_items:"Manual|Auto" form_section:"Unit Rate" form_pos:"1,1"`
	Divider               float64   `form_required:"1" form_section:"Unit Rate" form_pos:"1,2"`
	StandbyRate           float64   `form_required:"1" form_section:"Unit Rate" form_pos:"1,3"`
	WorkingHour           float64   `form_required:"1" form_section:"Unit Rate" label:"Working Hours" form_pos:"1,4"`
	Created               time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CustomerJournalTypeConfiguration) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Unit Rate", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *CustomerJournalTypeConfiguration) TableName() string {
	return "CustomerJournalTypeConfigurations"
}

func (o *CustomerJournalTypeConfiguration) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomerJournalTypeConfiguration) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomerJournalTypeConfiguration) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomerJournalTypeConfiguration) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomerJournalTypeConfiguration) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomerJournalTypeConfiguration) PostSave(dbflex.IConnection) error {
	return nil
}
