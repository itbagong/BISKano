package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Claim struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid:"hide"`
	EmployeeID        string                    `form_label:"Employee" grid_label:"Name" form_lookup:"/tenant/employee/find|_id|Name"`
	Position          string                    `form_read_only:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	JournalTypeID     string                    `form_lookup:"/fico/vendorjournaltype/find|_id|_id,Name" grid:"hide"`
	Summary           []ClaimSummaryAmount      `form_section:"Summary" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_section_direction:"row" grid:"hide" form_section_size:"3"`
	IsActive          bool                      `grid:"hide"`
	Lines             []ClaimLine               `form_section:"Lines" grid:"hide" form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type ClaimSummaryAmount struct {
	ClaimTypeID        string `grid_label:"Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CLT|_id|Name"`
	ClaimSummaryAmount float64
	Balance            float64
	OffsetAccount      string `form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
}

type ClaimLine struct {
	Date        time.Time `form_kind:"date"`
	Mutation    float64
	ClaimTypeID string `grid_label:"Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CLT|_id|Name"`
	isPlus      bool   `grid:"hide"`
}

func (o *Claim) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Summary", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Lines", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *Claim) TableName() string {
	return "BGClaims"
}

func (o *Claim) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Claim) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Claim) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Claim) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Claim) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Claim) PostSave(dbflex.IConnection) error {
	return nil
}
