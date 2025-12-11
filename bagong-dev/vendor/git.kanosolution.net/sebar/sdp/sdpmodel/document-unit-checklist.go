package sdpmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DocumentUnitChecklistGrid struct {
	WONo     string `grid_label:"WO No"`
	SUNID    string `grid_label:"SUN ID"`
	AssetID  string
	ChasisNo string `grid_label:"Chasis No"`
	Status   string
}

type DocumentUnitChecklist struct {
	orm.DataModelBase             `bson:"-" json:"-"`
	ID                            string                    `bson:"_id" json:"_id" key:"1" key:"1" form_read_only_edit:"1" form:"hide"`
	WONo                          string                    `form:"hide" grid_label:"WO No"`
	SUNID                         string                    `form_section_size:"2" grid_label:"SUN ID" form_section_direction:"row" form_section:"General"`
	InvoiceNo                     string                    `form_section_size:"2" form_section_direction:"row" form_section:"General"`
	EngineNo                      string                    `form_section_size:"2" form_section_direction:"row" form_section:"General3"`
	AssetID                       string                    `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_label:"Asset ID" form_lookup:"/tenant/item/find|_id|Name"`
	AssetName                     string                    `form:"hide"`
	ChasisNo                      string                    `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_label:"Chasis No" grid_label:"Chasis No"`
	SKRBDate                      time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_kind:"date" form_label:"SKRB Date"`
	SRUTNO                        string                    `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_label:"SRUT No"`
	AreaRekomPeruntukan           string                    `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MDP|_id|Name" form_section_direction:"row" form_section:"General2" form_label:"Area Rekom Peruntukan"`
	SubmissionDateUjiKIR          time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_label:"Submission Date Uji KIR" form_kind:"date"`
	SKRBNo                        string                    `form_section_size:"2" form_section_direction:"row" form_section:"General3" form_label:"SKRB No"`
	HullNo                        string                    `form_section_size:"2" form_section_direction:"row" form_section:"General"`
	SubmissionDateSRUT            time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General" form_label:"Submission Date SRUT" form_kind:"date"`
	SRUTDate                      time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General3" form_label:"SRUT Date" form_kind:"date"`
	SubmissionDateRekomPeruntukan time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General" form_label:"Submission Date Rekom Peruntukan" form_kind:"date"`
	SubmissionDateSamsat          time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General" form_label:"Submission Date Samsat" form_kind:"date"`
	RoutePermitDate               time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General3" form_label:"Route Permit Date" form_kind:"date"`
	SubmissionDatePolres          time.Time                 `form_section_size:"2" form_section_direction:"row" form_section:"General2" form_label:"Submission Date Polres" form_kind:"date"`
	Dimension                     tenantcoremodel.Dimension `form_section:"Dimension" form_section_size:"2" form_section_direction:"row" form_label:"Finance Dimension" grid:"hide"`
	StatusRekomPeruntukan         string                    `form_read_only:"1" form:"hide" grid:"hide"`
	StatusSamsat                  string                    `form_read_only:"1" form:"hide" grid:"hide"`
	StatusSRUT                    string                    `form_read_only:"1" form:"hide" grid:"hide"`
	StatusUjiKIR                  string                    `form_read_only:"1" form:"hide" grid:"hide"`
	StatusRoutePermit             string                    `form_read_only:"1" form:"hide" grid:"hide"`
	StatusFinal                   string                    `form_read_only:"1" form:"hide" grid:"hide"`
	Created                       time.Time                 `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate                    time.Time                 `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info"`
}

func (o *DocumentUnitChecklist) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *DocumentUnitChecklist) TableName() string {
	return "DocumentUnitChecklist"
}

func (o *DocumentUnitChecklist) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *DocumentUnitChecklist) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *DocumentUnitChecklist) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *DocumentUnitChecklist) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *DocumentUnitChecklist) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() &&
		o.SubmissionDateSamsat.IsZero() &&
		o.SubmissionDateUjiKIR.IsZero() &&
		o.RoutePermitDate.IsZero() &&
		o.SubmissionDatePolres.IsZero() {
		o.StatusSRUT = "Pengajuan SRUT"
	} else if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() == false &&
		o.SubmissionDateSamsat.IsZero() &&
		o.SubmissionDateUjiKIR.IsZero() &&
		o.RoutePermitDate.IsZero() &&
		o.SubmissionDatePolres.IsZero() {
		o.StatusRekomPeruntukan = "Pengajuan Peruntukan"
	} else if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() == false &&
		o.SubmissionDateSamsat.IsZero() == false &&
		o.SubmissionDateUjiKIR.IsZero() &&
		o.RoutePermitDate.IsZero() &&
		o.SubmissionDatePolres.IsZero() {
		o.StatusSamsat = "Pengajuan Samsat"
	} else if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() == false &&
		o.SubmissionDateSamsat.IsZero() == false &&
		o.SubmissionDateUjiKIR.IsZero() == false &&
		o.RoutePermitDate.IsZero() &&
		o.SubmissionDatePolres.IsZero() {
		o.StatusUjiKIR = "Pengajuan Uji KIR"
	} else if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() == false &&
		o.SubmissionDateSamsat.IsZero() == false &&
		o.SubmissionDateUjiKIR.IsZero() == false &&
		o.RoutePermitDate.IsZero() == false &&
		o.SubmissionDatePolres.IsZero() {
		o.StatusRoutePermit = "Ijin Trayek"
	} else if o.SubmissionDateSRUT.IsZero() == false &&
		o.SubmissionDateRekomPeruntukan.IsZero() == false &&
		o.SubmissionDateSamsat.IsZero() == false &&
		o.SubmissionDateUjiKIR.IsZero() == false &&
		o.RoutePermitDate.IsZero() == false &&
		o.SubmissionDatePolres.IsZero() == false {
		o.StatusFinal = "Pengajuan Polres"
	} else {
		o.StatusFinal = "Need Action"
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *DocumentUnitChecklist) PostSave(dbflex.IConnection) error {
	return nil
}
