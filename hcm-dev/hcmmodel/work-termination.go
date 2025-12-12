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

type WorkTermination struct {
	orm.DataModelBase             `bson:"-" json:"-"`
	ID                            string                     `bson:"_id" json:"_id" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3"`
	RequestDate                   time.Time                  `form_kind:"date" form_section:"General"`
	Requestor                     string                     ` form_section:"General" label:"Requestor Name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	CompanyID                     string                     `grid:"hide" form:"hide"`
	JournalTypeID                 string                     `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID              string                     `grid:"hide" form:"hide"`
	ResignDate                    time.Time                  `form_required:"1" form_kind:"date" form_section:"General"`
	EmployeeID                    string                     `form_required:"1" form_section:"General" label:"Employee ID" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	Reason                        string                     `form_section:"General2"`
	Type                          string                     `form_section:"General2"`
	InterviewAnswers              tenantcoremodel.References `form_section:"General2" grid:"hide"`
	Administrative                tenantcoremodel.Checklists `grid:"hide"`
	NonTaxableIncomeTemplateID    string                     `grid:"hide"`
	NonTaxableIncome              []NonTaxableIncomeDetail   `grid:"hide"`
	TaxableIncomeTemplateID       string                     `grid:"hide"`
	TaxableIncome                 []TaxableIncomeDetail      `grid:"hide"`
	MandatoryWorkTemplateID       string                     `grid:"hide"`
	MandatoryWork                 []MandatoryWorkDetail      `grid:"hide"`
	Status                        ficomodel.JournalStatus    `form_section:"General2"`
	ResignOnCompanyInitiative     string                     `grid:"hide" form_kind:"radio" form_items:"Pelanggaran|Rasionalisasi|Medical Unfit|Meninggal Dunia|Pensiun|Tidak Memenuhi Standar Prestasi|Lain-lain (sebutkan)"`
	ResignOnEmployeeInitiative    string                     `grid:"hide" form_kind:"radio" form_items:"Salary|Keluarga|Melanjutkan Studi|Jenjang Karir|Kondisi Sosial|Mengundurkan Diri|Lain-lain (sebutkan)"`
	ResignOnCompanyInitiativeEtc  string                     `grid:"hide"`
	ResignOnEmployeeInitiativeEtc string                     `grid:"hide"`
	NettoAmount                   float64                    `grid:"hide"`
	TaxableAmountWorkerAccept     float64                    `grid:"hide"`
	NettoAmountMandatoryWork      float64                    `grid:"hide"`
	AmountWorkerAccept            float64                    `grid:"hide"`
	AmountWorkerAcceptWord        string                     `grid:"hide"`
	Dimension                     tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created                       time.Time                  `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate                    time.Time                  `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info"`
}

type NonTaxableIncomeDetail struct {
	Number      string
	Name        string
	Calculation float64
	PPNo        string
	Amount      float64
}

type TaxableIncomeDetail struct {
	Number      string
	Name        string
	Calculation interface{}
	Amount      float64
}

type MandatoryWorkDetail struct {
	Number      string
	Name        string
	Calculation float64
	RVNo        string
	Amount      float64
}

func (o *WorkTermination) TableName() string {
	return "HCMWorkTerminations"
}

func (o *WorkTermination) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkTermination) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkTermination) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkTermination) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkTermination) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkTermination) PostSave(dbflex.IConnection) error {
	return nil
}

type WorkTerminationForm struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                     `bson:"_id" json:"_id" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"4"`
	RequestDate       time.Time                  `form_kind:"date" form_section:"General"`
	Requestor         string                     ` form_section:"General" label:"Requestor Name" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	ResignDate        time.Time                  `form_kind:"date" form_section:"General2" form_required:"1" `
	Reason            string                     `form_section:"General2"`
	JournalTypeID     string                     `grid:"hide" form_required:"1" form_section:"General2"`
	PostingProfileID  string                     `grid:"hide" form:"hide"`
	Type              string                     `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=WorkTerminationType|_id|Name"`
	InterviewAnswers  tenantcoremodel.References `form_section:"General2" form:"hide"`
	Administrative    tenantcoremodel.Checklists `form:"hide"`
	Status            string                     `form_section:"General2" form_read_only:"1"`
	EmployeeID        string                     `form_required:"1" form_section:"General3" form_lookup:"/tenant/employee/find|_id|_id,Name"`
	NIK               string                     `form_section:"General3" label:"NIK" form_read_only:"1"`
	EmployeeName      string                     `form_required:"1" form_section:"General3" label:"Name" form_read_only:"1"`
	Site              string                     `form_required:"1" form_section:"General3" label:"Site" form_read_only:"1"`
	PointOfHire       string                     `form_required:"1" form_section:"General3" label:"Point of Hire" form_read_only:"1"`
	JoinedDate        string                     `form_required:"1" form_section:"General3" label:"Joined Date" form_read_only:"1"`
	EmployeeDetail    string                     `form_section:"General3"`
	Dimension         tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"4"`
	Created           time.Time                  `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                  `form_kind:"datetime" grid:"hide" form_read_only:"1" form_section:"Time Info"`
}

func (o *WorkTerminationForm) FormSections() []suim.FormSectionGroup {
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
