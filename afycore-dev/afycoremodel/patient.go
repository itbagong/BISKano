package afycoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Patient struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	FirstName         string
	LastName          string
	MiddleName        string
	Gender            string
	BirthDate         time.Time
	Citizenship       string
	IDCardNo          string
	IDCardType        string
	Address           string
	City              string
	State             string
	PostalCode        string
	CountryID         string
	PhoneNo           string
	Email             string
	BloodType         string
	Notes             string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Patient) TableName() string {
	return "Patients"
}

func (o *Patient) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Patient) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Patient) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Patient) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Patient) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Patient) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Patient) DbIndexes() []dbflex.DbIndex {
	res := []dbflex.DbIndex{
		{Name: "ByIDCard", IsUnique: true, Fields: []string{"IDCardType", "IDCardNo"}},
	}
	return res
}
