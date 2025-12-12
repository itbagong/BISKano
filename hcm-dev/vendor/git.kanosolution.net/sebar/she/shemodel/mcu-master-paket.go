package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MCUMasterPackage struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                 `bson:"_id" json:"_id" key:"1" form_read_only:"1"`
	TrxDate           time.Time              `form_section:"General" form_kind:"date"`
	PackageName       string                 `form_section:"General"`
	Provider          string                 `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MPR|_id|Name" form_section:"General"`
	Lines             []MCUMasterPackageLine `form_section:"Line" grid:"hide" form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *MCUMasterPackage) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Info", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type MCUMasterPackageLine struct {
	JenisPemeriksaan string                `form_lookup:"/she/mcuitemtemplate/find|_id|Name"`
	Lines            []MCUItemTemplateLine `form_section:"Detail" grid:"hide"`
}

func (o *MCUMasterPackage) TableName() string {
	return "SHEMCUMasterPackages"
}

func (o *MCUMasterPackage) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MCUMasterPackage) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MCUMasterPackage) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MCUMasterPackage) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MCUMasterPackage) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MCUMasterPackage) PostSave(dbflex.IConnection) error {
	return nil
}
