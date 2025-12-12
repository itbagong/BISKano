package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemSerial struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	ItemID            string    `form_required:"1" form_section:"General" form_lookup:"/tenant/item/find|_id|Name"`
	ReleaseDate       time.Time `form_required:"1" form_section:"General" form_kind:"date"`
	ExpiryDate        time.Time `form_required:"1" form_section:"General" form_kind:"date"`
	LifeTimeDate      time.Time `form_section:"General" form_kind:"date"`
	LifeTimeValue     float64   `form_section:"General"`
	LifeTimeValueUnit string    `form_section:"General"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemSerial) TableName() string {
	return "ItemSerials"
}

func (o *ItemSerial) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemSerial) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemSerial) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemSerial) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemSerial) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemSerial) PostSave(dbflex.IConnection) error {
	return nil
}
