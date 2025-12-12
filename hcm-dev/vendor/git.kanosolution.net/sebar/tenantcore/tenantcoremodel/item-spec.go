package tenantcoremodel

import (
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/sebar"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemSpec struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"Specification" form_section_auto_col:"2"`
	ItemID            string `grid:"hide"`
	SKU               string `form_required:"1" form_section:"Specification" label:"SKU"`
	OtherName         string `form_section:"Specification"`
	SpecVariantID     string `form_required:"1" label:"Variant" form_section:"Specification" form_lookup:"/tenant/specvariant/find|_id|Name"`
	SpecSizeID        string `form_required:"1" label:"Size" form_section:"Specification" form_lookup:"/tenant/specsize/find|_id|Name"`
	SpecGradeID       string `form_required:"1" label:"Grade" form_section:"Specification" form_lookup:"/tenant/specgrade/find|_id|Name"`
	SpecID            string
	IsActive          bool      `form_section:"Specification" label:"Is Active"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemSpec) TableName() string {
	return "ItemSpecs"
}

func (o *ItemSpec) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemSpec) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemSpec) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemSpec) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemSpec) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	o.Calc()
	return nil
}

func (o *ItemSpec) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *ItemSpec) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "ItemID", Fields: []string{"ItemID"}},
		{Name: "SpecVariantID", Fields: []string{"SpecVariantID"}},
		{Name: "SpecSizeID", Fields: []string{"SpecSizeID"}},
		{Name: "SpecGradeID", Fields: []string{"SpecGradeID"}},
		{Name: "SKU", Fields: []string{"SKU"}},
		{Name: "OtherName", Fields: []string{"OtherName"}},
	}
}

func (o *ItemSpec) SpecHash() string {
	fieldToHash := []string{
		o.ItemID,
		o.SKU,
		o.SpecVariantID,
		o.SpecSizeID,
		o.SpecGradeID}

	return codekit.MD5String(strings.Join(fieldToHash, "|"))
}

func (o *ItemSpec) Calc() *ItemSpec {
	o.SpecID = o.SpecHash()
	return o
}

func (o *ItemSpec) GenerateName(
	itemORM *sebar.MapRecord[*Item],
	specVariantORM *sebar.MapRecord[*SpecVariant],
	specSizeORM *sebar.MapRecord[*SpecSize],
	specGradeORM *sebar.MapRecord[*SpecGrade],
) string {
	separator := " - "
	texts := []string{}

	if data, _ := itemORM.Get(o.ItemID); data != nil && data.Name != "" {
		texts = append(texts, data.Name)
	}

	if o.SKU != "" {
		texts = append([]string{o.SKU}, texts...)
	}

	if o.OtherName != "" {
		texts = append(texts, o.OtherName)
	}

	if data, _ := specVariantORM.Get(o.SpecVariantID); data != nil && data.Name != "" {
		texts = append(texts, data.Name)
	}

	if data, _ := specSizeORM.Get(o.SpecSizeID); data != nil && data.Name != "" {
		texts = append(texts, data.Name)
	}

	if data, _ := specGradeORM.Get(o.SpecGradeID); data != nil && data.Name != "" {
		texts = append(texts, data.Name)
	}

	return strings.Join(texts, separator)
}
