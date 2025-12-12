package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Meeting struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_new:"1" form_read_only_edit:"1" form_section:"General" form_label:"No."`
	MeetingType       string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MTY|_id|Name" form_required:"1"`
	Title             string
	MeetingDate       time.Time
	Location          string                       `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	Attachments       []tenantcoremodel.Attachment `grid:"hide"`
	Attendee          []string                     `grid:"hide" form_lookup:"/tenant/employee/find|_id|Name"`
	LocationDetail    string                       `grid:"hide" form_multi_row:"5" form_space_after:"1"`
	Description       string                       `form_multi_row:"5"`
	Dimension         tenantcoremodel.Dimension    `grid:"hide" form_section:"Dimension" form_section_size:"1"`
	Result            []MeetingResult              `form:"hide" grid:"hide"`
	Status            string                       `form:"hide" form_section:"Info"`
	Created           time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                    `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
}

func (o *Meeting) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type MeetingResult struct {
	ResultId string `grid:"hide"`
	Problem  string `form_multi_row:"5"`
	Solution string `form_multi_row:"5"`
	UsePica  bool   `grid:"hide"`
	Pica     *Pica  `form_section:"Pica"  form_section_size:"3"`
}

func (o *Meeting) TableName() string {
	return "SHEMeetings"
}

func (o *Meeting) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Meeting) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Meeting) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Meeting) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Meeting) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Meeting) PostSave(dbflex.IConnection) error {
	return nil
}
