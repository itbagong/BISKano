package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerGroup struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string          `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" form_pos:"1,3"`
	Name               string          `form_required:"1" form_section:"General" form_pos:"1,3"`
	CustomerGroupAlias string          `form_pos:"1,3"`
	Setting            CustomerSetting `form_pos:"2,2"`
	Dimension          Dimension       `grid:"hide" form_pos:"2,2"`
	IsActive           bool            `form_pos:"3"`
	Created            time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time       `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *CustomerGroup) TableName() string {
	return "CustomerGroups"
}

func (o *CustomerGroup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *CustomerGroup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *CustomerGroup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *CustomerGroup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *CustomerGroup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *CustomerGroup) PostSave(dbflex.IConnection) error {
	return nil
}
