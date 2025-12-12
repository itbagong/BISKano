package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JournalTypeContext struct {
	Label   string
	Address string
}

type VendorSubmissionType struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name                string `form_required:"1" form_section:"General"`
	VendorGroup         string `form_required:"1" form_section:"General" form_lookup:"/tenant/vendorgroup/find|_id|Name"`
	Vendor              string `form_required:"1" form_section:"General" form_lookup:"/tenant/vendor/find|_id|Name"` // TODO: UI: find Customer where GroupID == choosen VendorGroup
	TransactionType     string `form_items:"Vendor Purchase|Credit Note"`
	NumberSequenceID    string
	DefaultOffset       SubledgerAccount          `grid:"hide"`
	Tag1Type            string                    `form_section:"Tag Object" grid:"hide" form_items:"NONE|ASSET|ITEM|EMPLOYEE"`
	Tag2Type            string                    `form_section:"Tag Object" grid:"hide" form_items:"NONE|ASSET|ITEM|EMPLOYEE"`
	PostingProfileID    string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ChecklistTemplateID string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ReferenceTemplateID string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	TaxCodes            []string                  `form_section:"Tax and Charges" form_lookup:"/fico/taxcode/find|_id|Name"`
	ChargeCodes         []string                  `form_section:"Tax and Charges" form_lookup:"/fico/chargecode/find|_id|Name"`
	Actions             []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Previews            []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Dimension           tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created             time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate          time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *VendorSubmissionType) TableName() string {
	return "VendorJournalTypes"
}

func (o *VendorSubmissionType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *VendorSubmissionType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *VendorSubmissionType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *VendorSubmissionType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *VendorSubmissionType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.DefaultOffset.AccountType == "" {
		o.DefaultOffset.AccountType = string(ficomodel.SubledgerNone)
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *VendorSubmissionType) PostSave(dbflex.IConnection) error {
	return nil
}
