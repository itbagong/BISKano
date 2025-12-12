package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RentalContracts struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	AssetID           string                     `form_lookup:"/tenant/asset/find|_id|Name"`
	FromDate          time.Time                  `form_required:"1" form_section:"Period" form_section_auto_col:"2" form_kind:"date" form_date_format:"DD MMM YYYY"`
	ToDate            time.Time                  `form_required:"1" form_section:"Period" form_section_auto_col:"2" form_kind:"date" form_date_format:"DD MMM YYYY"`
	CustomerID        string                     `form_lookup:"/tenant/customer/find|_id|Name"`
	RentalRate        float64                    `form_kind:"number"`
	Status            string                     `form_items:"Active|Hold|Stopped"`
	References        tenantcoremodel.References `grid:"hide" form:"hide"`
	CompanyID         string                     `grid:"hide" form:"hide"`
	Created           time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *RentalContracts) TableName() string {
	return "RentalContracts"
}

func (o *RentalContracts) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RentalContracts) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RentalContracts) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RentalContracts) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RentalContracts) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RentalContracts) PostSave(dbflex.IConnection) error {
	return nil
}
