package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PreviewSectionType string

const (
	PreviewAsGrid PreviewSectionType = "Grid"
	PreviewAsText PreviewSectionType = "Text"
	PreviewAsHTML PreviewSectionType = "HTML"
)

type PreviewSection struct {
	Title               string             `bson:"Title"`
	HideTitle           bool               `bson:"HideTitle"`
	HideHeader          bool               `bson:"HideHeader"`
	TemplateFile        string             `bson:"TemplateFile"`
	RestrictedByFeature bool               `bson:"RestrictedByFeature"`
	FeatureIDs          []string           `bson:"FeatureIDs"`
	SectionType         PreviewSectionType `bson:"SectionType"`
	Content             string             `bson:"Content"`
	Items               [][]string         `bson:"Items"`
}

type Signature struct {
	ID        string `bson:"_id"`
	Header    string `bson:"Header"`
	Footer    string `bson:"Footer"`
	Confirmed string `bson:"Confirmed"`
	Status    string `bson:"Status"`
}

type PreviewReport struct {
	Header       codekit.M                 `bson:"Header"`
	HeaderMobile PreviewReportHeaderMobile `bson:"HeaderMobile"`
	HideHeader   bool                      `bson:"HideHeader"`
	Sections     []PreviewSection          `bson:"Sections"`
	Signature    []Signature               `bson:"Signature"`
}

type Preview struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string
	SourceType        string
	SourceJournalID   string
	SourceTrxType     string
	VoucherNo         string
	PreviewReport     *PreviewReport
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type PreviewReportHeaderMobile struct {
	Data   [][]string
	Footer [][]string
}

func (o *Preview) TableName() string {
	return "Previews"
}

func (o *Preview) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Preview) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Preview) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Preview) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Preview) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Preview) PostSave(dbflex.IConnection) error {
	return nil
}
