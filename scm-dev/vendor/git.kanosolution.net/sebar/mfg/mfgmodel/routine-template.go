package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoutineTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"2" form_section:"General" form_pos:"1,1"`
	CategoryID        string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RTC|_id|_id,Name"`
	AssetType         string                    `form_section:"General" label:"Asset Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=AUT|_id|_id,Name"` // for lookup
	DriveType         string                    `form_section:"General" label:"Drive Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DTY|_id|_id,Name"` // for lookup
	CustomerID        string                    `form_section:"General" label:"Customer" form_lookup:"/fico/customerbalance/find|_id|_id,Name"`                     // for lookup (kalo kosong, bisa dipake global)
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_pos:"3"`                                                                 // for lookup: PC, Site (kalo kosong, bisa dipake global)
	Items             []RoutineTemplateItems    `form_section:"General" grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}
type RoutineTemplateItems struct {
	ItemNo        int `form_read_only:"1"`
	ItemName      string
	DangerousCode string `form_use_list:"1" form_items:"AA|A"`
}

func (o *RoutineTemplate) TableName() string {
	return "RoutineTemplates"
}

func (o *RoutineTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RoutineTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RoutineTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoutineTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoutineTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RoutineTemplate) PostSave(dbflex.IConnection) error {
	return nil
}
