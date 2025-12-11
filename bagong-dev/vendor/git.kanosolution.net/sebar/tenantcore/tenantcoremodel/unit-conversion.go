package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UnitConversion struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	FromUnit          string
	ToQty             float64   // FromQty always 1, ex: TON = 1000 KG | 1 KG = 1000 GRAM | GRAM = 0,001 KG | GRAM = 0,000001 TON (changed per 27 Jan 2024)
	ToUnit            string    `form_lookup:"/tenant/unit/find|_id|Name"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *UnitConversion) TableName() string {
	return "UnitConversions"
}

func (o *UnitConversion) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *UnitConversion) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *UnitConversion) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UnitConversion) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UnitConversion) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *UnitConversion) PostSave(dbflex.IConnection) error {
	return nil
}
