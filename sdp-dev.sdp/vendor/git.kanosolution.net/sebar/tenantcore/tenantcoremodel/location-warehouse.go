package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LocationWarehouse struct {
	orm.DataModelBase        `bson:"-" json:"-"`
	ID                       string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name                     string    `form_required:"1" form_section:"General" `
	Address                  string    `form_multi_row:"2" form_section:"General"`
	PIC                      string    `form_section:"General" label:"PIC" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	WhsGroupID               string    `form_section:"General2" form_section_size:"3" form_lookup:"/tenant/warehouse/group/find|_id|_id,Name"`
	PostingProfileIDReceive  string    `form_section:"General2" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	PostingProfileIDIssuance string    `form_section:"General2" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	IsActive                 bool      `form_section:"General2"`
	Created                  time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2" `
	LastUpdate               time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	Dimension                Dimension `grid:"hide" form_section:"Dimension"`
}

func (o *LocationWarehouse) TableName() string {
	return "LocationWarehouses"
}

func (o *LocationWarehouse) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LocationWarehouse) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LocationWarehouse) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LocationWarehouse) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LocationWarehouse) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LocationWarehouse) PostSave(dbflex.IConnection) error {
	return nil
}
