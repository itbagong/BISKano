package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LegalRegister struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                       `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" label:"Ref No." form_section_auto_col:"4" form_section_direction:"row"`
	Date              time.Time                    `form_section:"General" form_kind:"date"`
	LegalNo           string                       `form_section:"General" form_required:"1"`
	JournalTypeID     string                       `grid:"hide" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                       `form:"hide" grid:"hide"`
	RelatedSite       SHERelatedSite               `grid:"hide" form_multiple:"1" form_items:"BTS|MINING|AKDP/ALBN|PARIWISATA|HO" form_section:"General" form_required:"1"`
	Type              string                       `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LTY|_id|Name" form_section:"General"`
	Category          string                       `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LCA|_id|Name" form_section:"General"`
	Fields            []string                     `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LFI|_id|Name" form_section:"General1"`
	Link              string                       `grid:"hide" form_multi_row:"3" form_section:"General1"`
	Reference         []tenantcoremodel.Attachment `grid:"hide" form_section:"General1"`
	StatusDoc         bool                         `form_section:"General1"`
	Status            string                       `grid:"hide" form:"hide"`
	PlantCompliance   int                          `grid:"hide" form:"hide"`
	ActualCompliance  int                          `grid:"hide" form:"hide"`
	Achievement       float64                      `grid:"hide" form:"hide"`
	LegalDetails      []LegalDetail                `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension    `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                       `grid:"hide" form_read_only:"1"  form_section:"Info"`
}

func (o *LegalRegister) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "LegalDetails", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type LegalDetail struct {
	Subject        string `form_multi_row:"3"`
	ActivityPoints []ActivityPoint
}

type ActivityPoint struct {
	ID       string `grid:"hide" bson:"_id" json:"_id" `
	Value    string
	IsComply bool
}

type SHESite struct {
	ID         string `bson:"_id" json:"_id" `
	Name       string
	Address    string
	Alias      string
	Dimension  tenantcoremodel.Dimension
	IsActive   bool
	Created    time.Time
	LastUpdate time.Time
}

func (o *LegalRegister) TableName() string {
	return "SHELegalRegisters"
}

func (o *LegalRegister) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LegalRegister) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LegalRegister) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LegalRegister) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LegalRegister) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Status == "" {
		o.Status = string(SHEStatusDraft)
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LegalRegister) PostSave(dbflex.IConnection) error {
	return nil
}
