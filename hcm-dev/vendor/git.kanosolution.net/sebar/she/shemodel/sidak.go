package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Fatigue struct {
	SleepDuration       float64 `form_read_only:"1"`
	Sleep               string
	WakeUp              string
	Reporter            string `form_lookup:"/tenant/employee/find|_id|Name"`
	Fit                 bool
	MedicineConsumption bool
	MedicineDescription string                       `form_multi_row:"5"`
	Sign                []tenantcoremodel.Attachment `form_space_after:"1"`
	IsReadyWork         string                       `form_section:"Supervisor Action" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SPVAction|_id|Name"`
	Remark              string                       `form_multi_row:"5" form_section:"Supervisor Action"`
}

func (o *Fatigue) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 4},
			{Title: "Supervisor Action", ShowTitle: true, AutoCol: 1},
		}},
	}
}

type SpeedGun struct {
	LocationID     string                       `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name" grid_label:"Location" form_required:"1" form_section:"Speed Gun"`
	LocationDetail string                       `form_section:"Speed Gun"`
	Speed          int                          `form_section:"Speed Gun" label:"Speed (km/h)"`
	Reporter       string                       `form_section:"Speed Gun" form_lookup:"/tenant/employee/find|_id|Name"`
	Evidance       []tenantcoremodel.Attachment `form_section:"Speed Gun"`
	Remark         string                       `form_multi_row:"5" form_section:"Remark"`
}

func (o *SpeedGun) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Speed Gun", ShowTitle: false, AutoCol: 4},
			{Title: "Remark", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type Alcohol struct {
	Alcohol  float64                      `form_section:"Alcohol" label:"Alcohol (mg/l)"`
	Evidance []tenantcoremodel.Attachment `form_section:"Alcohol"`
	Sign     []tenantcoremodel.Attachment `form_section:"Alcohol"`
	Reporter string                       `form_section:"Alcohol" form_lookup:"/tenant/employee/find|_id|Name"`
	Remark   string                       `form_section:"Alcohol" form_multi_row:"5"`
}

func (o *Alcohol) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Alcohol", ShowTitle: false, AutoCol: 4},
		}},
	}
}

type Drug struct {
	Amphetam   bool                         `form_section:"Drug"`
	Morphin    bool                         `form_section:"Drug"`
	Marijuana  bool                         `form_section:"Drug"`
	Benzodiaze bool                         `form_section:"Drug"`
	Menthapet  bool                         `form_section:"Drug"`
	Cocain     bool                         `form_section:"Drug"`
	Evidance   []tenantcoremodel.Attachment `form_section:"Drug"`
	Sign       []tenantcoremodel.Attachment `form_section:"Drug"`
	Reporter   string                       `form_section:"Drug" form_lookup:"/tenant/employee/find|_id|Name"`
	Remark     string                       `form_section:"Drug" form_multi_row:"5" `
}

func (o *Drug) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Drug", ShowTitle: false, AutoCol: 4},
		}},
	}
}

type Sidak struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only:"1"`
	DateTime          time.Time
	EmployeeID        string `label:"Name" form_lookup:"/tenant/employee/find|_id|Name"`
	Position          string `form_read_only:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTE|_id|Name"`
	Mess              bool
	Penalty           SidakPenalty              `grid:"hide" form_section:"Penalty"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Fatigue           Fatigue                   `form:"hide"`
	SpeedGun          SpeedGun                  `form:"hide"`
	Alcohol           Alcohol                   `form:"hide"`
	Drug              Drug                      `form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *Sidak) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 2},
			{Title: "Penalty", ShowTitle: false, AutoCol: 2},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

type SidakPenalty struct {
	SP     []tenantcoremodel.Attachment
	PHK    []tenantcoremodel.Attachment
	Reason string `form_multi_row:"5" `
}

func (o *Sidak) TableName() string {
	return "SHESidaks"
}

func (o *Sidak) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Sidak) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Sidak) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Sidak) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Sidak) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Sidak) PostSave(dbflex.IConnection) error {
	return nil
}
