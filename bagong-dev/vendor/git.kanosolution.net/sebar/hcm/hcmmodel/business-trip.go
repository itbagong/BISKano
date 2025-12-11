package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BusinessTripLineDetail struct {
	Description string
	Cost        float64
}
type BusinessTripLine struct {
	EmployeeID string                   `label:"Employee Name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	Location   string                   `form:"hide"`
	Task       string                   `grid:"hide" form:"hide"`
	Details    []BusinessTripLineDetail `grid:"hide"`
	TotalCost  float64                  `form_read_only_edit:"1" form_read_only_new:"1"`
}

type BusinessTrip struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	CompanyID         string                    `grid:"hide" form:"hide"`
	JournalTypeID     string                    `grid:"hide" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Business Trip|_id|Name"`
	PostingProfileID  string                    `grid:"hide" form:"hide"`
	RequestorID       string                    `form_label:"Requestor Name" form_section:"General" form_lookup:"/tenant/employee/find|_id|Name"`
	RequestDate       time.Time                 `form_section:"General2" form_kind:"date"`
	DateFrom          time.Time                 `form_section:"General3" form_kind:"date"`
	DateTo            time.Time                 `form_section:"General" form_kind:"date"`
	Status            ficomodel.JournalStatus   `form:"hide"`
	Lines             []BusinessTripLine        `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type BusinessTripForm struct {
	ID                string                    `bson:"_id" json:"_id" grid:"hide" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	RequestorID       string                    `form_label:"Requestor Name" form_section:"General" form_lookup:"/tenant/employee/find|_id|Name"`
	RequestorNIK      string                    `form_section:"General" form_read_only:"1" label:"Requestor NIK"`
	RequestorPosition string                    `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name" form_read_only:"1"`
	RequestorEmail    string                    `form_section:"General" form_read_only:"1"`
	RequestorSite     string                    `form_section:"General" form_read_only:"1" form_lookup:"/tenant/dimension/find?DimensionType=Site|_id|Name"`
	RequestDate       time.Time                 `form_section:"General2" form_kind:"date"`
	JournalTypeID     string                    `grid:"hide" form_section:"General2" form_required:"1" form_lookup:"/hcm/journaltype/find?TransactionType=Business Trip|_id|Name"`
	PostingProfileID  string                    `grid:"hide" form_section:"General2" form:"hide"`
	DateFrom          time.Time                 `form_section:"General3" form_kind:"date"`
	DateTo            time.Time                 `form_section:"General3" form_kind:"date"`
	Status            string                    `form:"hide"`
	Lines             []BusinessTripLine        `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *BusinessTripForm) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *BusinessTrip) TableName() string {
	return "HCMBusinessTrips"
}

func (o *BusinessTrip) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *BusinessTrip) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *BusinessTrip) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *BusinessTrip) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *BusinessTrip) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *BusinessTrip) PostSave(dbflex.IConnection) error {
	return nil
}
