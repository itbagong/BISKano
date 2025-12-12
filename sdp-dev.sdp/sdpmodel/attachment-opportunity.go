package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AttachmentOpportunity struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	OpportunityId     string `form_required:"1" form_section:"General"` // Get id from Opportunity
	Name              string `form_required:"1" form_section:"General"`
	FileName          string `form_required:"1" form_section:"General"`
	OriginalName      string `form_required:"1" form_section:"General"`
	Type              string `form_required:"1" form_section:"General"`
	Path              string `form_required:"1" form_section:"General"`
	IsActive          bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AttachmentOpportunity) TableName() string {
	return "AttachmentOpportunity"
}

func (o *AttachmentOpportunity) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AttachmentOpportunity) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AttachmentOpportunity) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AttachmentOpportunity) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AttachmentOpportunity) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AttachmentOpportunity) PostSave(dbflex.IConnection) error {
	return nil
}
