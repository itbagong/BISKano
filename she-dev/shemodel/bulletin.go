package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bulletin struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" label:"No." form_section_auto_col:"3" form_section_direction:"row"`
	Title             string `form_section:"General"`
	Category          string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=BCT|_id|Name" form_section:"General"`
	Tag               []string `grid:"hide" form_section:"General"`
	Theme             string                       `grid:"hide" form_section:"General1"`
	RefNews           string                       `grid:"hide" form_section:"General1"`
	IsPin             bool                         `form_section_show_title:"1" label:"Pin"`
	IsStatus          bool                         `label:"Status" form_section:"General1"`
	Banner            []tenantcoremodel.Attachment `grid:"hide" form_section:"General1"`
	Dimension         tenantcoremodel.Dimension    `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Content           string                       `form_multi_row:"5" form_section:"Content" grid:"hide" form_kind:"html"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
}

func (o *Bulletin) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Content", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *Bulletin) TableName() string {
	return "SHEBulletins"
}

func (o *Bulletin) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Bulletin) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Bulletin) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Bulletin) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Bulletin) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Bulletin) PostSave(dbflex.IConnection) error {
	return nil
}
