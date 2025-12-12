package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerConfiguration struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `form:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1"`
	CustomerID        string `form:"hide"`
	Round             string `form_required:"1" form_section:"Physical Availability" label:"Round" form_pos:"1,1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RND|_id|Name"`
	Decimals          int    `form_required:"1" form_section:"Physical Availability" label:"Decimals" form_pos:"1,2"`

	DividerType string  `form_items:"Manual|Auto" form_section:"Unit Rate" form_pos:"1,1" form:"hide"`
	Divider     float64 `form_required:"1" form_section:"Unit Rate" form_pos:"1,2" form:"hide"`
	StandbyRate float64 `form_required:"1" form_section:"Unit Rate" form_pos:"1,3" form:"hide"`
	WorkingHour float64 `form_required:"1" form_section:"Unit Rate" label:"Working Hours" form_pos:"1,4" form:"hide"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CustomerConfiguration) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Physical Availability", ShowTitle: true, AutoCol: 1},
			// {Title: "Unit Rate", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *CustomerConfiguration) TableName() string {
	return "BGCustomerConfigurations"
}

func (o *CustomerConfiguration) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomerConfiguration) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomerConfiguration) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomerConfiguration) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomerConfiguration) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomerConfiguration) PostSave(dbflex.IConnection) error {
	return nil
}
