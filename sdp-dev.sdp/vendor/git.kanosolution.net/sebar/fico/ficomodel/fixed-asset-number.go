package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FixedAssetNumber struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form:"hide" grid:"hide"`
	Name              string                    `form_section:"General" form_required:"1"`
	TotalAssetNumber  int                       `form:"hide"`
	Details           []FixedAssetNumberDetail  `grid:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type FixedAssetNumberDetail struct {
	AssetName          string `form_required:"1"`
	FixedAssetGrup     string `form_lookup:"/tenant/assetgroup/find|_id|_id,Name"`
	NumberAsset        int    `form_required:"1"`
	InitialAssetNumber int    `form_read_only:"1"`
	LastAssetNumber    int    `form_read_only:"1"`
}

func (o *FixedAssetNumber) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *FixedAssetNumber) TableName() string {
	return "FixedAssetNumbers"
}

func (o *FixedAssetNumber) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *FixedAssetNumber) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *FixedAssetNumber) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FixedAssetNumber) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FixedAssetNumber) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *FixedAssetNumber) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *FixedAssetNumber) PreDelete(dbflex.IConnection) error {

	return nil
}
