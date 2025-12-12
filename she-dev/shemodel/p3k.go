package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type P3k struct {
// 	orm.DataModelBase `bson:"-" json:"-"`
// 	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_label:"No."`
// 	Date              time.Time `form_section:"General"`
// 	PatientType       string    `form_section:"General"`
// 	PatientName       string    `form_read_only:"1" form_lookup:"/tenant/employee/find|_id|Name"`
// 	Age               string    `form_section:"General"`
// 	Gender            string    `form_section:"General"`
// 	TimeIn            time.Time `form_section:"General"`
// 	TimeOut           time.Time `form_section:"General"`
// 	Doctor            string    `form_section:"General"`
// 	Purpose           string    `form_section:"General"`
// 	PC                string    `form_section:"General"`
// 	CC                string    `form_section:"General"`
// 	Site              string    `form_section:"General"`
// 	Asset             string    `form_section:"General"`

// 	Condition    string `form_section:"Consultancy"`
// 	MedicalNotes string `form_section:"Consultancy"`
// 	Diagnose     string `form_read_only:"1" form_lookup:"/tenant/diagnose/find|_id|Name"`
// 	VisitResult  string `form_read_only:"1" form_lookup:"/tenant/visitresult/find|_id|Name"`

// 	Medicines []MedLine `form_section:"Line" grid:"hide"`

// 	ConculPrice float32 `form_section:"Price"`
// 	MedPRice    float32 `form_section:"Price"`
// 	TotalPrice  float32 `form_section:"Price"`

// 	Height             float32 `form_section:"Measurement"`
// 	Weight             float32 `form_section:"Measurement"`
// 	Bmi                float32 `form_section:"Measurement"`
// 	Pulse              float32 `form_section:"Measurement"`
// 	ButaWarna          string  `form_section:"Measurement"`
// 	AlcoholTest        bool    `form_section:"Measurement"`
// 	DrugTest           string  `form_section:"Measurement"`
// 	WaistCircumference float32 `form_section:"Measurement"`
// 	Diastolic          float32 `form_section:"Measurement"`
// 	Systolic           float32 `form_section:"Measurement"`
// 	Cholesterol        float32 `form_section:"Measurement"`
// 	Ekg                float32 `form_section:"Measurement"`
// 	CholesterolTotal   float32 `form_section:"Measurement"`
// 	Ldl                float32 `form_section:"Measurement"`
// 	UricACid           float32 `form_section:"Measurement"`
// 	RandomBloodSugar   float32 `form_section:"Measurement"`
// 	FastingBloodSugar  float32 `form_section:"Measurement"`
// 	Hdl                float32 `form_section:"Measurement"`
// 	Tligriseride       float32 `form_section:"Measurement"`
// 	GulaDarah          float32 `form_section:"Measurement"`

// 	Status     string                    `form:"hide"`
// 	Dimension  tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
// 	Created    time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info" form_section_auto_col:"2"`
// 	LastUpdate time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form:"hide" form_section:"Time Info"`
// }

type P3k struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General1" label:"No." form_section_auto_col:"3" form_section_direction:"row"`
	Date              time.Time                 `form_kind:"date" form_section:"General1"`
	PatientType       bool                      `grid:"hide" form_section:"General1"` // Internal: true, External: false
	PatientName       string                    `form_section:"General1" form_lookup:"/tenant/employee/find|_id|Name"`
	Age               string                    `grid:"hide" form_section:"General1"`
	Gender            string                    `grid:"hide" form_section:"General1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=GEME|_id|Name"`
	TimeIn            time.Time                 `grid:"hide" form_kind:"time" form_section:"General2"`
	TimeOut           time.Time                 `grid:"hide" form_kind:"time" form_section:"General2"`
	Doctor            string                    `form_section:"General2" label:"Doctor/Paramedic"`
	Purpose           string                    `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=Purpose|_id|Name"`
	ReffNo            string                    `grid:"hide" form_section:"General2" form_read_only:"1" label:"Ref No"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension"`
	P3kDetail         `grid:"hide"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string    `form_read_only:"1" form_section:"Info"`
}

type P3kDetail struct {
	Condition    string    `grid:"hide" form_multi_row:"5" form_section:"Concultancy1"  form_section_auto_col:"3" form_section_direction:"row"`
	MedicalNotes string    `grid:"hide" form_multi_row:"5" form_section:"Concultancy1"`
	Diagnose     string    `form_section:"Concultancy2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=Diagnose|_id|Name"`
	VisitResult  string    `grid:"hide" form_section:"Concultancy2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=VisitResult|_id|Name"`
	Medicines    []MedLine `form_section:"LineMedicine" grid:"hide"`
	ConculPrice  float32   `grid:"hide" form:"hide" form_section:"Price"`
	MedPRice     float32   `grid:"hide" form:"hide" form_section:"Price"`
	TotalPrice   float32   `grid:"hide" form:"hide" form_section:"Price"`

	Height             float32 `grid:"hide" form_section:"Measurement1" label:"Height (cm)"`
	Weight             float32 `grid:"hide" form_section:"Measurement1" label:"Weight (kg)"`
	Bmi                float32 `grid:"hide" form_section:"Measurement1" label:"BMI (kg/m2)"`
	Pulse              float32 `grid:"hide" form_section:"Measurement1" label:"Pulse (nadi)"`
	ButaWarna          string  `grid:"hide" form_section:"Measurement1" label:"Buta Warna"`
	AlcoholTest        bool    `grid:"hide" form_section:"Measurement1" label:"Breath alcohol test"`
	DrugTest           `grid:"hide" form_section:"Measurement1" label:"Drug Test"`
	WaistCircumference float32 `grid:"hide" form_section:"Measurement2" label:"Waist Circumference (cm)"`
	Diastolic          float32 `grid:"hide" form_section:"Measurement2" label:"Diastolic (mmHg) / Systolic (mmHg)"`
	Systolic           float32 `grid:"hide" form:"hide" form_section:"Measurement2" label:"Systolic"`
	Cholesterol        float32 `grid:"hide" form_section:"Measurement2" label:"Cholesterol (mg/dl)"`
	Ekg                float32 `grid:"hide" form_section:"Measurement2" label:"EKG"`
	CholesterolTotal   float32 `grid:"hide" form_section:"Measurement2" label:"Cholestrol Total"`
	Ldl                float32 `grid:"hide" form_section:"Measurement2" label:"LDL"`
	UricACid           float32 `grid:"hide" form_section:"Measurement3" label:"Uric Acid (mg/dl)"`
	RandomBloodSugar   float32 `grid:"hide" form_section:"Measurement3" label:"Random Blood Sugar (mg/dl)"`
	FastingBloodSugar  float32 `grid:"hide" form_section:"Measurement3" label:"Fasting Blood Sugar (mmHg)"`
	Hdl                float32 `grid:"hide" form_section:"Measurement3" label:"HDL"`
	Tligriseride       float32 `grid:"hide" form_section:"Measurement3" label:"Tligriseride"`
	GulaDarah          float32 `grid:"hide" form_section:"Measurement3" label:"Gula darah PP (2 jam setelah puasa)"`
}

func (o *P3k) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *P3kDetail) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Concultancy", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Concultancy1", ShowTitle: false, AutoCol: 1},
			{Title: "Concultancy2", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Treatment/Medicine", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "LineMedicine", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Health Measurement", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Measurement1", ShowTitle: false, AutoCol: 1},
			{Title: "Measurement2", ShowTitle: false, AutoCol: 1},
			{Title: "Measurement3", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type MedLine struct {
	ID     string `grid:"hide" bson:"_id" json:"_id" `
	MedNo  int    `label:"No" form_read_only:"1"`
	Name   string `form_lookup:"/tenant/item/find?ItemGroupID=GRP0016|_id|Name"`
	Qty    int
	Price  float32
	Remark string
}

type DrugTest struct {
	Amphetamine bool `grid:"hide" `
	Morphin     bool `grid:"hide"`
	Menthapet   bool `grid:"hide"`
	Cocain      bool `grid:"hide"`
	Marijuana   bool `grid:"hide"`
	Benzodiaze  bool `grid:"hide"`
}

func (o *P3k) TableName() string {
	return "SHEP3k"
}

func (o *P3k) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *P3k) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *P3k) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *P3k) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *P3k) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Status == "" {
		o.Status = string(SHEStatusDraft)
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *P3k) PostSave(dbflex.IConnection) error {
	return nil
}
