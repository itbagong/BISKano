package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MasterSOPSummary struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" label:"Ref No."`
	DocumentType      SOPDocumentType
	TitleOfDocument   string
	EffectiveDate     time.Time `form_kind:"date" `
	Attachment        tenantcoremodel.Attachment
	IsActive          bool      `grid:"hide" form:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *MasterSOPSummary) TableName() string {
	return "SHEMasterSOPSummarys"
}

func (o *MasterSOPSummary) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MasterSOPSummary) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MasterSOPSummary) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MasterSOPSummary) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MasterSOPSummary) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MasterSOPSummary) PostSave(dbflex.IConnection) error {
	return nil
}
