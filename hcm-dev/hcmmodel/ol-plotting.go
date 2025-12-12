package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OLPlotting struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" form_section:"General"  form_section_auto_col:"2" form_read_only:"1"`
	CompanyID         string `grid:"hide" form:"hide"`
	JournalTypeID     string `grid:"hide" form_required:"1" form_section:"General" form_lookup:"/hcm/journaltype/find?TransactionType=OL%20%26%20Plotting|_id|Name"`
	PostingProfileID  string `grid:"hide" form:"hide"`
	CandidateID       string `form_lookup:"/tenant/employee/find|_id|Name" form:"hide" form_section:"General" grid_read_only:"1"`
	JobVacancyID      string `form_lookup:"/hcm/manpowerrequest/find|_id|Name"    form:"hide" form_section:"General"`
	Plotting          string `form_lookup:"/tenant/dimension/find?DimensionType=Site|_id|Label" form_pos:"1,2"  form_section:"General"  form_section_auto_col:"2"`

	// Offering Letter
	OfferingLetter bool   `grid:"hide" form_pos:"1,1"  form_section:"General"`
	Position       string `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"  form_section:"Offering Letter" form_section_show_title:"1" form_section_auto_col:"2"`
	Level          string `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LME|_id|Name"  form_section:"Offering Letter"`
	SubGroup       string `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SGE|_id|Name"  form_section:"Offering Letter"`
	Department     string `grid:"hide" form_lookup:"/tenant/dimension/find?DimensionType=CC|_id|Label" form_section:"Offering Letter" `
	POH            string `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PME|_id|Name" form_section:"Offering Letter" form_label:"POH"`
	WorkLocation   string `grid:"hide" form_lookup:"/bagong/sitesetup/find|_id|Name" form_section:"Offering Letter"`
	EmployeeStatus string `grid:"hide" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=ESM|_id|Name"  form_section:"Offering Letter"`
	ContractPeriod int    `grid:"hide" form_section:"Offering Letter" form_label:"Contract Period (month)"`
	Salary         int    `grid:"hide" form_section:"Offering Letter2" form_section_auto_col:"2" `
	Benefit        string `grid:"hide" form_section:"Offering Letter2"`
	WorkingHour    string `grid:"hide" form_section:"Offering Letter3" form_section_auto_col:"2"`
	THR            string `grid:"hide" form_section:"Offering Letter3" form_label:"THR"`
	BPJSTK         string `grid:"hide" form_section:"Offering Letter3" form_label:"BPJSTK"`
	BPJSHealth     string `grid:"hide" form_section:"Offering Letter3" form_label:"BPJS Kesehatan"`
	Facility       string `grid:"hide" form_section:"Offering Letter4" form_multi_row:"5"`

	Status      ficomodel.JournalStatus `form_read_only:"1" form_section:"General"`
	StageStatus string                  `form_read_only:"1" form_section:"General"`
	Created     time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
	LastUpdate  time.Time               `grid:"hide" form:"hide" form_section:"Time Info"`
}

func (o *OLPlotting) TableName() string {
	return "HCMOLPlottings"
}

func (o *OLPlotting) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *OLPlotting) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *OLPlotting) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *OLPlotting) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *OLPlotting) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *OLPlotting) PostSave(dbflex.IConnection) error {
	return nil
}
