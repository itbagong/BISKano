package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RSCA struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1"`
	Name              string                    `form_section:"General"`
	Location          string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail    string                    `form_section:"General" form_multi_row:"3"`
	RSCATeam          []string                  `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name" label:"RSCA Team"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Lines             []RSCALine                `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
}

type RSCALine struct {
	ID               string `grid:"hide"`
	LineNo           string `label:"No" form_read_only:"1"`
	Category         string `label:"Category"`
	RiskNo           string
	Division         string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RSCADivision|_id|Name"`
	Department       string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RSCADept|_id|Name"`
	CriticalActivity string    `label:"Critical Activity / Process" form_lookup:"/she/legalregister/find|_id|LegalNo"`
	ParentId         string    `grid:"hide"`
	CreatedBy        string    `form_read_only:"1"`
	CreatedTime      time.Time `grid:"hide"`
	UpdatedBy        string    `form_read_only:"1"`
	UpdatedTime      time.Time `grid:"hide"`
	IsUpdated        bool      `grid:"hide"`
}

func (o *RSCA) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *RSCA) TableName() string {
	return "SHERSCAs"
}

func (o *RSCA) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RSCA) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RSCA) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RSCA) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RSCA) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RSCA) PostSave(dbflex.IConnection) error {
	return nil
}
