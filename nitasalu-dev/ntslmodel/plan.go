package ntslmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MealType string
type PlanStatus string

const (
	MealTypeNone   MealType = "None"
	MealTypeFull   MealType = "Full"
	MealTypeLight  MealType = "Light"
	MealTypeMinmal MealType = "Minimal"

	PlanDraft      PlanStatus = "Draft"
	PlanOpen       PlanStatus = "Open"
	PlanExpired    PlanStatus = "Expired"
	PlanInProgress PlanStatus = "InProgress"
	PlanExecuted   PlanStatus = "Executed"
	PlanCancelled  PlanStatus = "Cancelled"
	PlanDeferred   PlanStatus = "Deferred"
)

type PlanPreference struct {
	MakkahTour             bool
	MadinahTour            bool
	ExtraTours             []string
	Brakfast               MealType
	Lunch                  MealType
	Dinner                 MealType
	MakkahHotelDistanceKM  int
	MakkahHotelType        int
	MadinahHotelDistanceKM int
	MadinahHotelType       int
}

type PlanCity struct {
	CityID   string
	NumOfDay int
}

type Plan struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	UserID            string
	MemberCount       int
	ExpectedDateFrom  *time.Time
	ExpectedDateTo    *time.Time
	ExpectedCities    []string
	ExpectedBudget    float64
	Preferences       PlanPreference
	Status            PlanStatus
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Plan) TableName() string {
	return "Plans"
}

func (o *Plan) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Plan) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Plan) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Plan) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Plan) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Plan) PostSave(dbflex.IConnection) error {
	return nil
}
