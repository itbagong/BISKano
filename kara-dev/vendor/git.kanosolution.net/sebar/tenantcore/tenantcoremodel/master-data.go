package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MasterData struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"VendorForm1"`
	MasterDataTypeID  string    `form_section:"VendorForm1" form_lookup:"/tenant/masterdatatype/find|_id|Name" form:"hide"`
	ParentID          string    `form_section:"VendorForm1" form_lookup:"/tenant/masterdata/find|_id|Name"`
	Name              string    `form_section:"VendorForm1"`
	IsActive          bool      `form_section:"VendorForm1"`
	Created           time.Time `form_read_only:"1" grid:"hide" form_section:"VendorForm2"`
	LastUpdate        time.Time `form_read_only:"1" grid:"hide"  form_section:"VendorForm2"`
}

func (o *MasterData) TableName() string {
	return "MasterData"
}

func (o *MasterData) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MasterData) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MasterData) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MasterData) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MasterData) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MasterData) PostSave(dbflex.IConnection) error {
	return nil
}
