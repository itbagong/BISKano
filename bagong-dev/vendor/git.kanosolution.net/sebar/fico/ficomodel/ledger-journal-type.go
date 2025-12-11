package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/kiva"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LedgerJournalType struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name                string                    `form_required:"1" form_section:"General"`
	TrxType             string                    `form_items:"General Journal|Asset Acquisition|Asset Disposal|Asset Depreciation"`
	NumberSequenceID    string                    `form_lookup:"/tenant/numseq/find|_id|_id,Name"`
	DefaultOffset       SubledgerAccount          `grid:"hide"`
	Tag1Type            string                    `form_section:"Tag Object" grid:"hide" form_items:"NONE|ASSET|ITEM|EMPLOYEE"`
	Tag2Type            string                    `form_section:"Tag Object" grid:"hide" form_items:"NONE|ASSET|ITEM|EMPLOYEE"`
	PostingProfileID    string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ChecklistTemplateID string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ReferenceTemplateID string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	Dimension           tenantcoremodel.Dimension `form_section:"Tag Object"`
	Actions             []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Previews            []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Attachment          bool                      `form_section:"Contexts"`
	Created             time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate          time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *LedgerJournalType) TableName() string {
	return "LedgerJournalTypes"
}

func (o *LedgerJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *LedgerJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *LedgerJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *LedgerJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *LedgerJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.Tag1Type == "" {
		o.Tag1Type = "NONE"
	}
	if o.Tag2Type == "" {
		o.Tag2Type = "NONE"
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *LedgerJournalType) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *LedgerJournalType) CacheSetup() *kiva.CacheOptions {
	return &kiva.CacheOptions{}
}
