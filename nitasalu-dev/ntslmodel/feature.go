package ntslmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureType string

const (
	FeatureFlight    FeatureType = "Flight"
	FeatureTransport FeatureType = "Transport"
	FeatureLodging   FeatureType = "Lodging"
	FeatureActivity  FeatureType = "Activity"
	FeatureMeal      FeatureType = "Meal"
	FeatureTour      FeatureType = "Tour"
)

type Feature struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	CityID            string
	FeatureType       FeatureType
	Enable            bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Feature) TableName() string {
	return "Features"
}

func (o *Feature) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Feature) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Feature) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Feature) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Feature) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Feature) PostSave(dbflex.IConnection) error {
	return nil
}
