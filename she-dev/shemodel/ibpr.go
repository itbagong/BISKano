package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBPR struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section_auto_col:"4" form_section_direction:"row"`
	Name              string                    `form_section:"General"`
	IBPRTeam          []string                  `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name" label:"IBPR Team"`
	Location          string                    `form_section:"General1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail    string                    `form_section:"General1" form_multi_row:"3"`
	JournalTypeID     string                    `grid:"hide" form_section:"General" form_lookup:"/fico/shejournaltype/find|_id|_id,Name"`
	PostingProfileID  string                    `form:"hide" grid:"hide"`
	Lines             []IBPRLine                `grid:"hide" form:"hide"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy         string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy      string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

type IBPRLine struct {
	ID                        string        `grid:"hide"`
	LineNo                    string        `label:"No" form_read_only:"1"`
	SituasiAktivitas          string        `label:"Situasi / Aktivitas"`
	RutinNonRutin             IBPRRutinitas `label:"Rutin / NonRutin" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IBPRRoutine|_id|Name"`
	NormalAbnormalEmergency   IBPRCondition `label:"Normal / Abnormal / Emergency" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IBPRStatus|_id|Name"`
	PeraturanTerkait          string        `form_lookup:"/she/legalregister/find|_id|LegalNo"`
	Lingkup                   IBPRLingkup   `label:"Lingkup (K3/L)" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IBPRType|_id|Name"`
	BahayaK3LAspekLingkungan  string        `label:"Bahaya K3 / Aspek Lingkungan" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IBPRAspek|_id|Name"`
	ResikoK3LDampakLingkungan string        `label:"Risiko  K3L / Dampak Lingkungan" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IBPRDampak|_id|Name"`
	ParentId                  string        `grid:"hide"`
	CreatedBy                 string        `form_read_only:"1"`
	CreatedTime               time.Time     `grid:"hide"`
	UpdatedBy                 string        `form_read_only:"1"`
	UpdatedTime               time.Time     `grid:"hide"`
	IsUpdated                 bool          `grid:"hide"`
}

func (o *IBPR) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *IBPR) TableName() string {
	return "SHEIBPRs"
}

func (o *IBPR) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *IBPR) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *IBPR) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *IBPR) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *IBPR) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *IBPR) PostSave(dbflex.IConnection) error {
	return nil
}
