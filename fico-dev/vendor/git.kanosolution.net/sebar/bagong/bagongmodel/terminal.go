package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Corridor struct {
	Name     string
	TrayekID string `form_lookup:"/bagong/trayek/find|_id|Name" label:"Trayek"`
}

type Terminal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                        `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string                        `form_required:"1" form_section:"General"`
	Corridor          []Corridor                    `grid:"hide" form_section:"General1" form_section_auto_col:"1"`
	Expenses          []tenantcoremodel.ExpenseType `grid:"hide" form_section:"General1"`
	IsActive          bool                          `form_section:"General2" form_section_auto_col:"2"`
	Dimension         tenantcoremodel.Dimension     `grid:"hide" form_section:"General2"`
	Created           time.Time                     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Terminal) TableName() string {
	return "BGTerminals"
}

func (o *Terminal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Terminal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Terminal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Terminal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Terminal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Terminal) PostSave(dbflex.IConnection) error {
	return nil
}
