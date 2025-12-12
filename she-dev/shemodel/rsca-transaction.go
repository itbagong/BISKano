package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RSCATransaction struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" label:"No" form_read_only:"1" form_section_auto_col:"4" form_section_direction:"row"`
	Name              string                    `form_section:"General"`
	TemplateID        string                    `form_section:"General" form_lookup:"/she/masterrsca/find|_id|Name" label:"Template"`
	Location          string                    `form_section:"General1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail    string                    `form_section:"General1" form_multi_row:"3"`
	IBPRTeam          []string                  `form_section:"General1" form_lookup:"/tenant/employee/find|_id|Name" label:"IBPR Team"`
	Lines             []RSCATrxLine             `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *RSCATransaction) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		// {Sections: []suim.FormSection{
		// 	{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		// }},
	}
}

type RSCATrxLine struct {
	ID                               string    `grid:"hide"`
	Category                         string    `label:"Category" form_read_only:"1"`
	LineNo                           string    `label:"Risk No" form_read_only:"1"`
	Division                         string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RSCARoutine|_id|Name"`
	Department                       string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=RSCAStatus|_id|Name"`
	CriticalActivity                 string    `label:"Critical Activity / Process" form_lookup:"/she/legalregister/find|_id|LegalNo"`
	ParentId                         string    `grid:"hide"`
	CreatedBy                        string    `form_read_only:"1" grid:"hide"`
	CreatedTime                      time.Time `grid:"hide"`
	UpdatedBy                        string    `form_read_only:"1" grid:"hide"`
	UpdatedTime                      time.Time `grid:"hide"`
	IsUpdated                        bool      `grid:"hide"`
	DescriptionRiskType              string    `label:"Description Risk Type"`
	DescriptionRiskAddessed          string    `label:"Description Risk that need to be addressed" form_multi_row:"3"`
	DescriptionOpportunitiesAddessed string    `label:"Description Opportunities that need to be addressed" form_multi_row:"3"`
	DescriptionCause                 string    `label:"Description Risk Cause" form_multi_row:"3"`
	DescriptionRiskLevel             string    `label:"Description Risk Level"`
	InherentImpact                   string    `label:"Inherent Impact" form_lookup:"/bagong/severity/find?Type=RSCA|_id|ParameterName"`
	InherentLikelihood               string    `label:"Inherent Likelihood" form_lookup:"/bagong/likelihood/find?Type=RSCA|_id|ParameterName"`
	InherentRiskLevel                string    `label:"Inherent Risk Level" form_lookup:"/bagong/riskmatrix/find?Type=RSCA|_id|RiskID"`
	ExistingControl                  string    `label:"Existing Control" form_multi_row:"3"`
	ResidualImpact                   string    `label:"Residual Impact" form_lookup:"/bagong/severity/find?Type=RSCA|_id|ParameterName"`
	ResidualLikelihood               string    `label:"Residual Likelihood" form_lookup:"/bagong/likelihood/find?Type=RSCA|_id|ParameterName"`
	ResidualRiskLevel                string    `label:"Residual Risk Level" form_lookup:"/bagong/riskmatrix/find?Type=RSCA|_id|RiskID"`
	AcceptRisk                       bool      `label:"Accept Risk"`
	TreatmenPlan                     string    `label:"Treatmen Plan" form_multi_row:"3"`
	ExpectedImpact                   string    `label:"Expected Impact" form_lookup:"/bagong/severity/find?Type=RSCA|_id|ParameterName"`
	ExpectedLikelihood               string    `label:"Expected Likelihood" form_lookup:"/bagong/likelihood/find?Type=RSCA|_id|ParameterName"`
	ExpectedRiskLevel                string    `label:"Expected Risk Level" form_lookup:"/bagong/riskmatrix/find?Type=RSCA|_id|RiskID"`
	PIC                              string    `label:"PIC" form_lookup:"/tenant/employee/find|_id|Name"`
	DueDate                          time.Time `label:"Due Date" form_kind:"date"`
	Status                           string    `label:"Progress" form_items:"Done|In Progress|Cancel|Overdue"`
}

func (o *RSCATransaction) TableName() string {
	return "SHERSCATransactions"
}

func (o *RSCATransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *RSCATransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *RSCATransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RSCATransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RSCATransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *RSCATransaction) PostSave(dbflex.IConnection) error {
	return nil
}
