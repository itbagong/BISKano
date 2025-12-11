package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetJournalType struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name                string                    `form_required:"1" form_section:"General"`
	NumberSequenceID    string                    `form_section:"General" form_section_direction:"row" form_section_size:"3"`
	TransactionType     string                    `form_section:"General" form_items:"LEDGER ACCOUNT|CUSTOMER GROUP|CUSTOMER|CASH IN|CASH OUT"`
	DefaultOffset       SubledgerAccount          `form_section:"General" grid:"hide" form_section_direction:"row" form_section_auto_col:"1"`
	UseTag1             bool                      `form_section:"General2" grid:"hide" form_section_direction:"row" form_section_auto_col:"1"`
	UseTag2             bool                      `form_section:"General2" grid:"hide" form_section_direction:"row"`
	Tag1Type            string                    `form_section:"General2" grid:"hide" form_items:"ASSET|ITEM|EMPLOYEE"`
	Tag2Type            string                    `form_section:"General2" grid:"hide" form_items:"ASSET|ITEM|EMPLOYEE"`
	PostingProfileID    string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ChecklistTemplateID string                    `form_section:"General2" grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ReferenceTemplateID string                    `form_section:"General2" grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	TaxCodes            []string                  `form_section:"General2" form_lookup:"/fico/taxcode/find|_id|Name"`
	ChargeCodes         []string                  `form:"hide" form_section:"General3" form_lookup:"/fico/chargecode/find|_id|Name"`
	Actions             []JournalTypeContext      `grid:"hide" form_section:"General3"`
	Previews            []JournalTypeContext      `grid:"hide" form_section:"General3"`
	Attachment          bool                      `form_section:"General3"`
	Dimension           tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created             time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate          time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AssetJournalType) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *AssetJournalType) TableName() string {
	return "AssetJournalType"
}

func (o *AssetJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
