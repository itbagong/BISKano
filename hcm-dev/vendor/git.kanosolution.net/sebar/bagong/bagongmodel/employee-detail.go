package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmployeeDetail struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string     `bson:"_id" json:"_id" key:"1" form:"hide" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"3"`
	EmployeeID             string     `form_required:"1" form_section:"General" label:"Employee ID"`
	EmployeeNo             string     `form_required:"1" form_section:"General" label:"Employee Identity Number"`
	JobVacancyTitle        string     `form_required:"1" form_section:"General" label:"Job Vacancy Title"`
	SocialNo               string     `form_required:"1" form_section:"General" label:"Social Number"`
	PlaceOfBirth           string     `form_required:"1" form_section:"General" label:"Place of Birth"`
	DateOfBirth            time.Time  `form_required:"1" form_section:"General" label:"Date of Birth" form_kind:"date"`
	Age                    string     `form_required:"1" form_section:"General" label:"Age"`
	Religion               string     `form_required:"1" form_section:"General" label:"Religion"`
	MaritalStatus          string     `form_required:"1" form_section:"General" label:"Marital Status"`
	Gender                 string     `form_required:"1" form_section:"General" label:"Gender"`
	IdentityCardNo         string     `form_required:"1" form_section:"General" label:"Identity Card No"`
	FamilyCardNo           string     `form_required:"1" form_section:"General" label:"Family Card No"`
	Address                string     `form_required:"1" form_section:"General" label:"Address"`
	Village                string     `form_required:"1" form_section:"General" label:"Village"`
	Subdistrict            string     `form_required:"1" form_section:"General" label:"Subdistrict"`
	City                   string     `form_required:"1" form_section:"General" label:"City"`
	Province               string     `form_required:"1" form_section:"General" label:"Province"`
	Domicile               string     `form_required:"1" form_section:"General" label:"Domicile"`
	PostCode               string     `form_required:"1" form_section:"General" label:"PostCode"`
	Phone                  string     `form_required:"1" form_section:"General" label:"Phone Number"`
	EmergencyPhone         string     `form_required:"1" form_section:"General" label:"Emergency Phone Number"`
	LastEducation          string     `form_required:"1" form_section:"General" label:"Last Education"`
	Major                  string     `form_required:"1" form_section:"General" label:"Major"`
	SchoolOrUniversityName string     `form_required:"1" form_section:"General" label:"Name of School/University"`
	WorkingExperience      int        `form_required:"1" form_section:"General" label:"Working Experience (Years)"`
	Position               string     `form_required:"1" form_section:"General" label:"Position"`
	Grade                  string     `form_required:"1" form_section:"General" label:"Grade"`
	Department             string     `form_required:"1" form_section:"General" label:"Department"`
	Rank                   string     `form_required:"1" form_section:"General" label:"Rank"`
	Level                  string     `form_required:"1" form_section:"General" label:"Level"`
	Group                  string     `form_required:"1" form_section:"General" label:"Group"`
	SubGroup               string     `form_required:"1" form_section:"General" label:"Sub Group"`
	UserCustomer           string     `form_required:"1" form_section:"General" label:"User Customer"`
	BPJSTKProgram          string     `form_required:"1" form_section:"General" label:"BPJS TK Program"`
	BPJSTKPercentage       float64    `form:"hide" form_section:"General" label:"BPJS TK Percentage"`
	BPJSKESPercentage      float64    `form:"hide" form_section:"General" label:"BPJS KES Percentage"`
	SIMType                string     `form_required:"1" form_section:"General" label:"SIM Type"`
	SIMNo                  string     `form_required:"1" form_section:"General" label:"SIM No"`
	SIMIssueDate           *time.Time `form_required:"1" form_section:"General" label:"SIM Issue Date"`
	SIMExpirationDate      *time.Time `form_required:"1" form_section:"General" label:"SIM Expiration Date"`
	DirectSupervisor       string     `form_required:"1" form_section:"General" label:"Direct Supervisor"`
	POH                    string     `form_required:"1" form_section:"General" label:"POH"`
	WorkingPeriod          string     `form_required:"1" form_section:"General" label:"Working Period"`
	EmployeeStatus         string     `form_required:"1" form_section:"General" label:"Employee Status"`
	PermanentEmployeeDate  *time.Time `form_required:"1" form_section:"General" label:"Permanent Employee Date" form_kind:"date"`
	WorkerStatus           string     `form_required:"1" form_section:"General" label:"Worker Status"`
	ResignationDate        *time.Time `form_required:"1" form_section:"General" label:"Resignation Date" form_kind:"date"`
	BasicSalary            float64    `form_required:"1" form_section:"General" label:"Basic Salary" form_kind:"number"`
	BankAccount            string     `form_required:"1" form_section:"General" label:"Bank Account"`
	BankAccountNo          string     `form_required:"1" form_section:"General" label:"Bank Account Number"`
	BankAccountName        string     `form_required:"1" form_section:"General" label:"Bank Account Name"`
	BPJSTK                 string     `form_required:"1" form_section:"General" label:"BPJS TK"`
	BPJSKES                string     `form_required:"1" form_section:"General" label:"BPJS KES"`
	TaxIdentityNo          string     `form_required:"1" form_section:"General" label:"NPWP Number"`
	TaxIdentityName        string     `form_required:"1" form_section:"General" label:"NPWP Name"`
	BiologicalMotherName   string     `form_required:"1" form_section:"General" label:"Biological Mother's Name"`
	SpouseName             string     `form_required:"1" form_section:"General" label:"Spouse Name"`
	SpousePlaceOfBirth     string     `form_required:"1" form_section:"General" label:"Spouse Place of Birth"`
	SpouseDateOfBirth      time.Time  `form_required:"1" form_section:"General" label:"Spouse Date of Birth" form_kind:"date"`
	FamilyMembers          []FamilyMembers
	Checklists             tenantcoremodel.Checklists `form:"hide" grid:"hide" `
	Created                time.Time                  `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate             time.Time                  `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info"`
}

type FamilyMembers struct {
	Name         string
	Relation     string    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=Relation|_id|Name"`
	PlaceOfBirth string    `label:"Place of Birth"`
	DateofBirth  time.Time `label:"Date of Birth" form_kind:"date"`
	Gender       string    `label:"Gender" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
}

func (o *EmployeeDetail) TableName() string {
	return "EmployeeDetails"
}

func (o *EmployeeDetail) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *EmployeeDetail) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *EmployeeDetail) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *EmployeeDetail) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *EmployeeDetail) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *EmployeeDetail) PostSave(dbflex.IConnection) error {
	return nil
}
