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

type OvertimeLine struct {
	EmployeeID     string `label:"Employee Name" form_lookup:"/tenant/employee/find|_id|Name"`
	Position       string
	Task           string
	ActualOvertime float64
	OffDay         bool
}

type Overtime struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	CompanyID          string                    `grid:"hide" form:"hide"`
	JournalTypeID      string                    `grid:"hide" form_required:"1" form_section:"General" form_lookup:"/hcm/journaltype/find?TransactionType=Overtime|_id|Name"`
	PostingProfileID   string                    `grid:"hide" form:"hide"`
	RequestorID        string                    `form_required:"1" form_label:"Requestor Name" form_section:"General" form_lookup:"/tenant/employee/find|_id|Name"`
	OvertimeDate       time.Time                 `form_section:"General" form_kind:"date" form_required:"1"`
	EstimatedStartTime string                    `form_section:"General2" form_kind:"time"`
	EstimatedEndTime   string                    `form_section:"General3" form_kind:"time"`
	Status             ficomodel.JournalStatus   `form:"hide"`
	Lines              []OvertimeLine            `grid:"hide" form:"hide"`
	Dimension          tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Overtime) FormSections() []suim.FormSectionGroup {
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

func (o *Overtime) TableName() string {
	return "HCMOvertimes"
}

func (o *Overtime) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Overtime) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Overtime) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Overtime) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Overtime) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Overtime) PostSave(dbflex.IConnection) error {
	return nil
}
