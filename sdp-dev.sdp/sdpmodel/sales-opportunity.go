package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesOpportunityGridView struct {
	OpportunityNo string     `label:"Opportunity No"`
	SalesType     string     `label:"Sales Opportunity Type" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOOT|_id|Name"`
	SalesStage    string     `label:"Sales Stage" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOSS|_id|_id,Name"`
	Customer      string     `form_lookup:"/tenant/customer/find|_id|Name"`
	OppStartDate  *time.Time `label:"Start Date Opportunity" form_kind:"date"`
	OppEndDate    *time.Time `label:"End Date Opportunity" form_kind:"date"`
	Name          string     `label:"Opportunity Name"`
	OppStatus     string     `label:"Opportunity Status" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOS|_id|Name"`
}
type SalesOpportunity struct {
	orm.DataModelBase `bson:"-" json:"-"`
	No                int64      `grid:"hide" form:"hide"`
	ID                string     `bson:"_id" json:"_id" key:"1" form_section:"Opportunity" form_section_size:"4" grid:"hide" form:"hide"`
	OpportunityNo     string     `form_required:"1"  form_read_only:"1" form_section:"Opportunity" label:"Opportunity No" form_section_show_title:"1"`
	Name              string     `form_required:"1" form_section:"Opportunity" label:"Opportunity Name" form_section_show_title:"1"`
	OppStartDate      *time.Time `form_required:"1" form_section:"Opportunity" label:"Start Date Opportunity" form_kind:"date"`
	OppEndDate        *time.Time `form_required:"1" form_section:"Opportunity" label:"End Date Opportunity" form_kind:"date"`
	CompanyID         string     `form_required:"1" form_section:"Opportunity" label:"Company ID" form_lookup:"/tenant/company/find|_id|Name"  grid:"hide" `
	TransactionType   string     `grid:"hide" form_required:"1" form_section:"Opportunity" label:"Transaction Type" form_items:"Asset|Item"`
	SalesType         string     `form_required:"1" form_section:"Sales" label:"Sales Opportunity Type" form_section_show_title:"1" form_section_size:"4" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOOT|_id|Name"`
	SalesStage        string     `form_required:"1" form_section:"Sales" label:"Sales Stage" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOSS|_id|_id,Name"`
	OppStatus         string     `form_required:"1" form_section:"Status" label:"Opportunity Status" form_section_show_title:"1" form_section_size:"4" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOS|_id|Name"`
	OppStatusReason   string     `form_required:"1" form_section:"Status" label:"Opportunity Status Reason" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOSR|_id|Name"`
	Customer          string     `form_required:"1" form_section:"Customer" form_section_show_title:"1" form_section_size:"4" form_lookup:"/tenant/customer/find|_id|Name"`
	GeneratedId       string     `grid:"hide" form:"hide"`
	Contact           string     `grid:"hide" form:"hide"`
	Phone             string     `grid:"hide" form:"hide"`
	Email             string     `grid:"hide" form:"hide"`
	Address           string     `grid:"hide" form:"hide"`
	// IsActive bool `form_section:"General4"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`

	Lines       []LinesOpportunity      `grid:"hide" form:"hide"`
	Competitors []CompetitorOpportunity `grid:"hide" form:"hide"`
	Events      []EventOpportunity      `grid:"hide" form:"hide"`
	Bonds       []BondOpportunity       `grid:"hide" form:"hide"`
	Created     time.Time               `grid:"hide" form:"hide"`
	LastUpdate  time.Time               `grid:"hide" form:"hide"`
}

func (o *SalesOpportunity) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Opportunity", ShowTitle: true, AutoCol: 1},
			{Title: "Sales", ShowTitle: true, AutoCol: 1},
			{Title: "Status", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Customer", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", AutoCol: 1},
		}},
	}
}

func (o *SalesOpportunity) TableName() string {
	return "SalesOpportunity"
}

func (o *SalesOpportunity) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SalesOpportunity) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SalesOpportunity) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SalesOpportunity) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SalesOpportunity) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SalesOpportunity) PostSave(dbflex.IConnection) error {
	return nil
}
