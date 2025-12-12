package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DetailItem struct {
	ID          string
	Number      string
	Description string
	Parent      string
}

// type Items []DetailItem

type ItemTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string       `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Module            string       `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MDL|_id|_id,Name"`
	Menu              string       `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MENU|_id|_id,Name" form_section_size:"3"`
	Name              string       `form_section:"General3"`
	Status            bool         `form_section:"General2"`
	Items             []DetailItem `grid:"hide" form:"hide"`
	Dimension         Dimension    `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	IsActive          bool         `form_section:"General3"`
	Created           time.Time    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemTemplate) FormSections() []suim.FormSectionGroup {
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

func (o *ItemTemplate) TableName() string {
	return "ItemTemplates"
}

func (o *ItemTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemTemplate) PostSave(dbflex.IConnection) error {
	return nil
}
