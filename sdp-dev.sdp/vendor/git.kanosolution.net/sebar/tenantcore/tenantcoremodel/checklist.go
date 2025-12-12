package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChecklistItem struct {
	Done        bool
	Key         string
	Description string
	PIC         string     `label:"PIC"`
	Expected    *time.Time `form_kind:"date"`
	Actual      *time.Time `form_kind:"date"`
	Notes       string
}

type Checklists []ChecklistItem

type ChecklistTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string     `form_required:"1" form_section:"General"`
	Checklists        Checklists `grid:"hide"`
	Dimension         Dimension  `grid:"hide"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ChecklistTemplate) TableName() string {
	return "ChecklistTemplates"
}

func (o *ChecklistTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ChecklistTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ChecklistTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ChecklistTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ChecklistTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ChecklistTemplate) PostSave(dbflex.IConnection) error {
	return nil
}
