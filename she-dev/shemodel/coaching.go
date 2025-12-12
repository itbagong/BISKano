package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Coaching struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_label:"No." form_section_auto_col:"2"`
	Coach                  string `form_lookup:"/tenant/employee/find|_id|Name"`
	CoachTitle             string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	JournalTypeID          string `grid:"hide" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	Coachee                string `form_lookup:"/tenant/employee/find|_id|Name"`
	CoacheeTitle           string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	PostingProfileID       string `form:"hide" grid:"hide"`
	Date                   time.Time
	Topic                  string                    `form_multi_row:"3" form_section:"Coach Suggestion" form_section_auto_col:"2"`
	Goal                   string                    `grid:"hide" form_multi_row:"3" form_section:"Coach Suggestion"`
	Benefit                string                    `grid:"hide" form_multi_row:"3" form_section:"Coach Suggestion"`
	Target                 string                    `form_multi_row:"3" form_section:"Coach Suggestion"`
	ProblemClarification   string                    `grid:"hide" form_multi_row:"3" form_section:"Coach Suggestion"`
	Improvement            string                    `grid:"hide" form_multi_row:"3" form_section:"Coach Suggestion"`
	ObstacleIdentification string                    `grid:"hide" form_multi_row:"3" form_section:"Coach Suggestion"`
	Feedback               string                    `form_multi_row:"3" form_section:"Coach Suggestion" width:"200px"`
	Dimension              tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"3"`
	Created                time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate             time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy              string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy           string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status                 string                    `form:"hide" form_section:"Info"`
}

func (o *Coaching) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "Coach Suggestion", ShowTitle: true, AutoCol: 3},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *Coaching) TableName() string {
	return "SHECoachings"
}

func (o *Coaching) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Coaching) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Coaching) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Coaching) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Coaching) PreSave(dbflex.IConnection) error {
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

func (o *Coaching) PostSave(dbflex.IConnection) error {
	return nil
}
