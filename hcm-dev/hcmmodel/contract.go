package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContractItemDetail struct {
	Aspect        string
	MaxScore      int
	AchievedScore int
}

type ContractAttendanceDetail struct {
	Name  string
	Score int
}

type ContractAttendance struct {
	Presence []ContractAttendanceDetail
	Absent   []ContractAttendanceDetail
	Sick     []ContractAttendanceDetail
	Leave    []ContractAttendanceDetail
	Late     []ContractAttendanceDetail
}

type Contract struct {
	orm.DataModelBase           `bson:"-" json:"-"`
	ID                          string                  `bson:"_id" json:"_id" form_section_direction:"row" form_read_only:"1" form_section:"Assesment" form_section_size:"3"`
	CompanyID                   string                  `grid:"hide" form:"hide"`
	JournalTypeID               string                  `grid:"hide" form_required:"1" form_section:"Assesment" form_lookup:"/hcm/journaltype/find?TransactionType=Contract|_id|Name"`
	PostingProfileID            string                  `grid:"hide" form:"hide"`
	EmployeeID                  string                  `form_lookup:"/tenant/employee/find|_id|Name" form_section:"Assesment"`
	JobID                       string                  `grid:"hide" form:"hide"`
	JobTitle                    string                  `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name" form_section:"Assesment"`
	JoinedDate                  time.Time               `form_section:"Assesment" form_kind:"date"`
	ExpiredContractDate         time.Time               `form_section:"Assesment" form_kind:"date"`
	Reviewer                    string                  `grid:"hide" form_section:"Assesment" form_lookup:"/tenant/employee/find|_id|Name"`
	Attendace                   ContractAttendance      `grid:"hide" form_section:"Attendance"`
	ItemTemplateID              string                  `grid:"hide" form_section:"DetailsTemplate" form_section_size:"3" form_lookup:"/she/mcuitemtemplate/find|_id|Name"`
	ItemDetails                 []ContractItemDetail    `grid:"hide" form_section:"Details"`
	MaxScoreTotal               int                     `grid:"hide" form_section:"Assesment" form:"hide"`
	AchievedScoreTotal          int                     `grid:"hide" form_section:"Assesment" form:"hide"`
	FinalScore                  float64                 `grid:"hide" form_section:"Assesment" form:"hide"`
	IsProbationEnd              bool                    `grid:"hide" form_section:"Result1" form_label:"Habis Probation/Kontrak"`
	IsBecomeEmployee            bool                    `grid:"hide" form:"hide" form_section:"Result1" form_label:"Diangkat menjadi karyawan tetap"`
	IsContractExtended          bool                    `grid:"hide" form:"hide" form_section:"Result1" form_label:"Perpanjang Masa Kontrak"`
	ExtendedExpiredContractDate *time.Time              `grid:"hide" form_section:"Result2"` // only appear in FE if IsContractExtended is true
	Status                      ficomodel.JournalStatus `form_section:"Assesment" form:"hide"`
	Created                     time.Time               `grid:"hide" form_section:"Assesment" form:"hide"`
	LastUpdate                  time.Time               `grid:"hide" form_section:"Assesment" form:"hide"`
}

func (o *Contract) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Assesment", ShowTitle: false, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Attendance", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "DetailsTemplate", ShowTitle: false},
		}},
		{Sections: []suim.FormSection{
			{Title: "Details", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Result", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Result1", ShowTitle: false, AutoCol: 1},
			{Title: "Result2", ShowTitle: false, AutoCol: 1},
		}},
	}
}
func (o *Contract) TableName() string {
	return "HCMContracts"
}

func (o *Contract) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Contract) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Contract) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Contract) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Contract) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Contract) PostSave(dbflex.IConnection) error {
	return nil
}
