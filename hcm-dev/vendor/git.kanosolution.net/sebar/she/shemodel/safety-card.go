package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SafetyCard struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_new:"1" form_read_only_edit:"1" form_section:"General" form_label:"No." form_space_after:"2"`
	CategoryID        string                    `form_lookup:"/tenant/masterdatatype/find?ParentID=SCC|_id|Name" grid_label:"Category" form_required:"1"`
	ActivityID        string                    `grid:"hide" form_lookup:"/tenant/masterdata/find|_id|Name" form_required:"1"`
	LocationID        string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name" grid_label:"Location" form_required:"1"`
	PositionID         string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name" grid_label:"Position" form_required:"1"`
	LocationDetail    string                    `grid:"hide" form_multi_row:"5"`
	DetailsFinding    DetailsFinding            `grid_label:"Finding" form_section:"Detail Finding" width:"300px"`
	FollowUp          DetailsFollowUp           `grid_label:"Response" form_section:"Follow Up" width:"300px"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Pica              *Pica                     `form_section:"PICA" label:"PICA"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *SafetyCard) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "Detail Finding", ShowTitle: true, AutoCol: 1},
			{Title: "Follow Up", ShowTitle: true, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type DetailsFinding struct {
	DetailsFinding string    `form_multi_row:"5" `
	Status         SHEStatus `form_required:"1" form_items:"Open|Completed"`
	Attachments    []tenantcoremodel.Attachment
}

type DetailsFollowUp struct {
	ResponseDescription string `form_multi_row:"5" `
	Attachments         []tenantcoremodel.Attachment
}

func (o *SafetyCard) TableName() string {
	return "SHESafetyCards"
}

func (o *SafetyCard) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SafetyCard) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SafetyCard) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SafetyCard) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SafetyCard) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SafetyCard) PostSave(dbflex.IConnection) error {
	return nil
}
