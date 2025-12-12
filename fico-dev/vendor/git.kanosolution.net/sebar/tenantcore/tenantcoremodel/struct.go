package tenantcoremodel

import "time"

type EmployeeTenantBagong struct {
	ID                     string
	Name                   string
	Email                  string
	EmployeeGroupID        string
	EmploymentType         EmploymentType
	CompanyID              string
	OtherCompanyIDs        []string
	Sites                  []string
	UserID                 string
	Dimension              Dimension
	IsLogin                bool
	IsActive               bool
	Created                time.Time
	LastUpdate             time.Time
	EmployeeID             string
	EmployeeNo             string
	SocialNo               string
	PlaceOfBirth           string
	DateOfBirth            time.Time
	Age                    string
	Religion               string
	MaritalStatus          string
	Gender                 string
	IdentityCardNo         string
	FamilyCardNo           string
	Address                string
	Village                string
	Subdistrict            string
	City                   string
	Province               string
	PostCode               string
	Phone                  string
	EmergencyPhone         string
	LastEducation          string
	Major                  string
	SchoolOrUniversityName string
	Position               string
	Grade                  string
	Department             string
	Rank                   string
	Level                  string
	Group                  string
	SubGroup               string
	UserCustomer           string
	BPJSTKProgram          string
	BPJSTKPercentage       float64
	BPJSKESPercentage      float64
	DirectSupervisor       string
	POH                    string
	WorkingPeriod          string
	EmployeeStatus         string
	PermanentEmployeeDate  *time.Time
	WorkerStatus           string
	ResignationDate        *time.Time
	BasicSalary            float64
	BankAccount            string
	BankAccountNo          string
	BankAccountName        string
	BPJSTK                 string
	BPJSKES                string
	TaxIdentityNo          string
	TaxIdentityName        string
	BiologicalMotherName   string
	SpouseName             string
	SpousePlaceOfBirth     string
	SpouseDateOfBirth      time.Time
}

type SiteTenantBagong struct {
	ID    string
	Name  string
	Alias string
}
