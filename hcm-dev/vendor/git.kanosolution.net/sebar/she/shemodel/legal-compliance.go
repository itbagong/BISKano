package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LegalCompliance struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" grid:"hide" `
	SiteID            string `form_lookup:"/bagong/sitesetup/find|_id|Name" label:"Site"`
	PlantCompliance   int    `grid:"hide" form:"hide"`
	ActualCompliance  int    `grid:"hide" form:"hide"`
	Compliance        float64
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *LegalCompliance) TableName() string {
	return "SHELegalCompliances"
}

func (o *LegalCompliance) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LegalCompliance) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LegalCompliance) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LegalCompliance) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LegalCompliance) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LegalCompliance) PostSave(dbflex.IConnection) error {
	return nil
}
