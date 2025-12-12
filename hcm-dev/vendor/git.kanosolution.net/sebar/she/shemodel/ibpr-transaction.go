package shemodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBPRTransaction struct {
	orm.DataModelBase      `bson:"-" json:"-"`
	ID                     string                    `bson:"_id" json:"_id" key:"1" label:"No" form_read_only:"1" form_section_auto_col:"4" form_section_direction:"row"`
	Name                   string                    `form_section:"General"`
	TemplateID             string                    `form_section:"General" form_lookup:"/she/masteribpr/find|_id|Name" label:"Template"`
	Location               string                    `form_section:"General1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=LOC|_id|Name"`
	LocationDetail         string                    `form_section:"General1" form_multi_row:"3"`
	IBPRTeam               []string                  `form_section:"General1" form_lookup:"/tenant/employee/find|_id|Name" label:"IBPR Team"`
	Lines                  []IBPRLine                `grid:"hide" form:"hide"`
	InitialRisks           []InitialRisk             `grid:"hide" form:"hide"`
	ResidualRisks          []ResidualRisk            `grid:"hide" form:"hide"`
	OpportunityAssessments []OpportunityAssessment   `grid:"hide" form:"hide"`
	Dimension              tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_size:"4"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	CreatedBy 		string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	LastUpdateBy 	string                    `grid:"hide" form_read_only:"1"  form_section:"Info"`
	Status            string                    `form_read_only:"1" form_section:"Info"`
}

func (o *IBPRTransaction) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General1", ShowTitle: false, AutoCol: 1},
			{Title: "Info", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		// {Sections: []suim.FormSection{
		// 	{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		// }},
	}
}

type InitialRisk struct {
	CurrentActions CurrentAction `label:"Kontrol yang ada saat ini"`
	Severity       string        `label:"Keparahan / Severity"`
	Probability    string        `label:"Kemungkinan / Probability"`
	RiskRating     string        `label:"Peringkat Resiko / Risk Rating"`
}

type GridConfigInitialRisk struct {
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
	UpdatedBy                 string        `form_read_only:"1"`
	IsUpdated                 bool          `grid:"hide"`
	CurrentActions            CurrentAction `label:"Kontrol yang ada saat ini"`
	Severity                  string        `label:"Keparahan / Severity" form_lookup:"/bagong/severity/find?Type=IBPR|_id|Level"`
	Probability               string        `label:"Kemungkinan / Probability" form_lookup:"/bagong/likelihood/find?Type=IBPR|_id|Level"`
	RiskRating                string        `label:"Peringkat Resiko / Risk Rating" form_lookup:"/bagong/riskmatrix/find?Type=IBPR|_id|RiskID"`
}

type ResidualRisk struct {
	CurrentActions     CurrentAction `label:"Tindakan Pengendalian Yang Direkomendasikan"`
	Severity           string        `label:"Keparahan / Severity"`
	Probability        string        `label:"Kemungkinan / Probability"`
	RiskRating         string        `label:"Peringkat Resiko / Risk Rating" form_items:"Normal|Abnormal|Emergency"`
	PIC                string
	DueDate            time.Time
	StatusPenyelesaian string
	Validasi           string
}

type GridConfigResidualRisk struct {
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
	UpdatedBy                 string        `form_read_only:"1"`
	IsUpdated                 bool          `grid:"hide"`
	CurrentActions            CurrentAction `label:"Tindakan Pengendalian Yang Direkomendasikan"`
	Severity                  string        `label:"Keparahan / Severity" form_lookup:"/bagong/severity/find?Type=IBPR|_id|Level"`
	Probability               string        `label:"Kemungkinan / Probability" form_lookup:"/bagong/likelihood/find?Type=IBPR|_id|Level"`
	RiskRating                string        `label:"Peringkat Resiko / Risk Rating" form_lookup:"/bagong/riskmatrix/find?Type=IBPR|_id|RiskID"`
	PIC                       string        `form_lookup:"/tenant/employee/find|_id|Name"`
	DueDate                   time.Time     `form_kind:"date"`
	StatusPenyelesaian        string        `form_items:"Selesai|Belum Selesai"`
	Validasi                  string        `label:"Validasi Penyelesaian (bukti, dll)"`
}

type OpportunityAssessment struct {
	PeluangTeridentifikasi bool `label:"Peluang teridentifikasi"`
	Possibility            string
	Severity               string `label:"Keparahan / Severity"`
	Probability            string `label:"Kemungkinan / Probability"`
	RiskRating             string `label:"Peringkat Resiko / Risk Rating"`
	RekomendasiRencanaAksi string
	PIC                    string
	DueDate                time.Time
	StatusPenyelesaian     string
	Validasi               string
}

type GridConfigOpportunityAssessment struct {
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
	UpdatedBy                 string        `form_read_only:"1"`
	IsUpdated                 bool          `grid:"hide"`
	PeluangTeridentifikasi    bool          `label:"Peluang teridentifikasi"`
	Possibility               string        `label:"Peluang" form_multi_row:"3"`
	Severity                  string        `label:"Keparahan / Severity" form_lookup:"/bagong/severity/find?Type=IBPR|_id|Level"`
	Probability               string        `label:"Kemungkinan / Probability" form_lookup:"/bagong/likelihood/find?Type=IBPR|_id|Level"`
	RiskRating                string        `label:"Peringkat Resiko / Risk Rating" form_lookup:"/bagong/riskmatrix/find?Type=IBPR|_id|RiskID"`
	RekomendasiRencanaAksi    string        `form_multi_row:"3"`
	PIC                       string        `form_lookup:"/tenant/employee/find|_id|Name"`
	DueDate                   time.Time     `form_kind:"date"`
	StatusPenyelesaian        string
	Validasi                  string `label:"Validasi Penyelesaian (bukti, dll)"`
}

type CurrentAction struct {
	Engineering      string `form_section:"General" form_multi_row:"2"`
	Administrasi     string `form_section:"General" form_multi_row:"2"`
	APD              string `form_section:"General" form_multi_row:"2"`
	Proaktif         string `form_section:"General" form_multi_row:"2"`
	PencegahanLimbah string `form_section:"General" form_multi_row:"2"`
	PengolahanLimbah string `form_section:"General" form_multi_row:"2"`
	Dilusi           string `form_section:"General" form_multi_row:"2"`
}

func (o *CurrentAction) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
		}},
	}
}

func (o *IBPRTransaction) TableName() string {
	return "SHEIBPRTransactions"
}

func (o *IBPRTransaction) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *IBPRTransaction) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *IBPRTransaction) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *IBPRTransaction) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *IBPRTransaction) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *IBPRTransaction) PostSave(dbflex.IConnection) error {
	return nil
}
