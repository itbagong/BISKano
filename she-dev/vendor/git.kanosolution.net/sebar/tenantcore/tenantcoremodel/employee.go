package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmploymentType string

const (
	PermanentEmployee EmploymentType = "PERMANENT"
	ContractEmployee  EmploymentType = "CONTRACT"
	OutsourceEmployee EmploymentType = "OUTSOURCE"
	ExternalEmployee  EmploymentType = "EXTERNAL"
)

type Employee struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string         `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string         `form_required:"1" form_section:"General"`
	Email             string         `form_required:"1" form_section:"General"`
	EmployeeGroupID   string         `form_lookup:"/tenant/employeegroup/find|_id|Name"`
	EmploymentType    EmploymentType `form_required:"1" form_items:"PERMANENT|CONTRACT|OUTSOURCE|EXTERNAL|PROBATION"`
	CompanyID         string         `form_lookup:"/tenant/company/find|_id|Name"`
	OtherCompanyIDs   []string       `form_lookup:"/tenant/company/find|_id|Name"`
	Sites             []string       `form_lookup:"/bagong/sitesetup/find|_id|Name"`
	UserID            string         `form_read_only:"1"`
	JoinDate          time.Time      `form_kind:"date"`
	Dimension         Dimension      `grid:"hide"`
	IsLogin           bool           `form:"hide"`
	IsActive          bool
	//-- tolong ditambahkan field2 yang diperlukan, untuk dokumen ditaruh diattachment jangan di sini
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Employee) TableName() string {
	return "Employees"
}

func (o *Employee) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Employee) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Employee) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Employee) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Employee) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Employee) PostSave(dbflex.IConnection) error {
	return nil
}
