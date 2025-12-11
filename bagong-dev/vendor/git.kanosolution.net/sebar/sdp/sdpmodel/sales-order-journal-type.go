package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesOrderJournalType struct {
	// orm.DataModelBase `bson:"-" json:"-"`
	// ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	// Name              string `form_required:"1" form_section:"General"`
	// PostingProfileID  string
	// Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	// LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`

	//
	orm.DataModelBase       `bson:"-" json:"-"`
	ID                      string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	Name                    string `form_required:"1" form_section:"General"`
	Customer                string `form_section:"General" form_lookup:"/tenant/customer/find|_id|Name"`
	NumberSequenceID        string
	TrxType                 InventTrxType             `form_read_only:"1" form_items:"Sales Order" form_section:"General"`
	DefaultOffset           SubledgerAccount          `grid:"hide"`
	UseTag1                 bool                      `form_section:"Tag Object" grid:"hide" form_section_auto_col:"1"`
	UseTag2                 bool                      `form_section:"Tag Object" grid:"hide"`
	Tag1Type                string                    `form_section:"Tag Object" grid:"hide" form_items:"ASSET|ITEM|EMPLOYEE"`
	Tag2Type                string                    `form_section:"Tag Object" grid:"hide" form_items:"ASSET|ITEM|EMPLOYEE"`
	PostingProfileID        string                    `form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ChecklistTemplateID     string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ChecklistTemplateLineID string                    `grid:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|Name"`
	ReferenceTemplateID     string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	ReferenceTemplateLineID string                    `grid:"hide" form_lookup:"/tenant/referencetemplate/find|_id|Name"`
	TaxCodes                []string                  `form_section:"Tax and Charges" form_lookup:"/fico/taxcode/find|_id|Name"`
	ChargeCodes             []string                  `form:"hide" form_section:"Tax and Charges" form_lookup:"/fico/chargecode/find|_id|Name"`
	Actions                 []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Previews                []JournalTypeContext      `grid:"hide" form_section:"Contexts"`
	Dimension               tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created                 time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate              time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SalesOrderJournalType) TableName() string {
	return "SalesOrderJournalTypes"
}

func (o *SalesOrderJournalType) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SalesOrderJournalType) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SalesOrderJournalType) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SalesOrderJournalType) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SalesOrderJournalType) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SalesOrderJournalType) PostSave(dbflex.IConnection) error {
	return nil
}
