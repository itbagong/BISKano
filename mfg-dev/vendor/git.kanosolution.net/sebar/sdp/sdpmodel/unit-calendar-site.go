package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UnitCalendarSite model for Site Unit Calendar
type UnitCalendarSite struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1"`
	StartDate         time.Time
	EndDate           time.Time

	Dimension tenantcoremodel.Dimension

	Created    time.Time
	LastUpdate time.Time
}

func (o *UnitCalendarSite) TableName() string {
	return "UnitCalendarSite"
}

func (o *UnitCalendarSite) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *UnitCalendarSite) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *UnitCalendarSite) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UnitCalendarSite) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UnitCalendarSite) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *UnitCalendarSite) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *UnitCalendarSite) PreDelete(dbflex.IConnection) error {
	// if o.Status != UnitCalendarDraft {
	// 	return errors.New("protected record, could not be deleted")
	// }

	return nil
}
