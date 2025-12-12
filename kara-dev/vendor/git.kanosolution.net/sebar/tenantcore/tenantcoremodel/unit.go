package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UOMRatioType string
type UOMCategory string

const (
	//UOMRatioType
	BiggerThanCategory   UOMRatioType = "Bigger than category"
	SmallerThanCategory  UOMRatioType = "Smaller than category"
	ReferenceForCategory UOMRatioType = "Reference for category"

	//UOMCategory
	Unit     UOMCategory = "Unit"
	Weight   UOMCategory = "Weight"
	Volume   UOMCategory = "Volume"
	Distance UOMCategory = "Distance/Length"
	Time     UOMCategory = "Time"
)

type UoM struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string       `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	Name              string       `form_required:"1" form_section:"General"`
	UOMCategory       UOMCategory  `form_section:"General2" form_items:"Unit|Weight|Volume|Distance|Time"`
	UOMRatioType      UOMRatioType `form:"hide" grid:"hide" form_items:"BiggerThanCategory|SmallerThanCategory|ReferenceForCategory"`
	RoundingPrecision float64 `form_section:"General2"`
	Ratio             int `form:"hide" grid:"hide"`
	Status            bool `form_section:"General2"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	Dimension         Dimension `grid:"hide" form_section:"Dimension"`
}

func (o *UoM) TableName() string {
	return "UoMs"
}

func (o *UoM) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *UoM) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *UoM) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UoM) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UoM) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *UoM) PostSave(dbflex.IConnection) error {
	return nil
}
