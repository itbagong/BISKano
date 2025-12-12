package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"2"`
	Name              string          `form_required:"1" form_section:"General2" form_section_direction:"row" form_section_size:"2"`
	CustomerAlias     string          `form_section:"General2" form_section_direction:"row" form_section_size:"2"`
	GroupID           string          `form_lookup:"/tenant/customergroup/find|_id|Name" form_section:"General"`
	Setting           CustomerSetting `grid:"hide" form_section:"General"` // sementara di hide krn bingung UI nya
	Dimension         Dimension       `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"2"`
	IsActive          bool            `form_section:"General2"`
	Created           time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Customer) FormSections() []suim.FormSectionGroup {
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

func (o *Customer) TableName() string {
	return "Customers"
}

func (o *Customer) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Customer) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Customer) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Customer) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Customer) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Customer) PostSave(dbflex.IConnection) error {
	return nil
}
