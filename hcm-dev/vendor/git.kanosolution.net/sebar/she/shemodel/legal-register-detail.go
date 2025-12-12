package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LegalRegisterDetail struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                       `bson:"_id" json:"_id" key:"1" form:"hide" label:"Ref No." form_section_auto_col:"3" form_section_direction:"row"`
	Date              time.Time                    `form_section:"General" grid:"hide"`
	LegalNo           string                       `form_section:"Legal Info" form_required:"1" form_read_only:"1"`
	Type              string                       `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LTY|_id|Name" form_section:"General"`
	Category          string                       `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LCA|_id|Name" form_section:"General1"`
	SiteID            string                       `grid:"hide" form:"hide"`
	RelatedSite       SHERelatedSite               `grid:"hide" form_lookup:"/bagong/sitesetup/find|_id|Name" form_section:"General1" form_required:"1"`
	Fields            []string                     `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LFI|_id|Name" form_section:"General1"`
	Link              string                       `grid:"hide" form_multi_row:"3" form_section:"Legal Info" form_read_only:"1"`
	Reference         []tenantcoremodel.Attachment `grid:"hide" form_section:"Legal Info" form_read_only:"1"`
	Status            bool                         `form_section:"General2"`
	PlantCompliance   int                          `form_section:"Summary" form_read_only:"1"`
	ActualCompliance  int                          `form_section:"Summary" form_read_only:"1"`
	Achievement       float64                      `form_section:"Summary" form_read_only:"1" label:"Achievement (%)"`
	LegalDetails      []LegalDetail                `form_section:"Compliance Assessment" grid:"hide"`
	Dimension         tenantcoremodel.Dimension    `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *LegalRegisterDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Summary", ShowTitle: true, AutoCol: 3},
		}},
		{Sections: []suim.FormSection{
			{Title: "Legal Info", ShowTitle: true, AutoCol: 3},
		}},
		{Sections: []suim.FormSection{
			{Title: "Compliance Assessment", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *LegalRegisterDetail) TableName() string {
	return "SHELegalRegisterDetails"
}

func (o *LegalRegisterDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LegalRegisterDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LegalRegisterDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LegalRegisterDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LegalRegisterDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LegalRegisterDetail) PostSave(dbflex.IConnection) error {
	return nil
}
