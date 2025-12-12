package tenantcoremodel

import (
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemSpec struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"Specification" form_section_auto_col:"2"`
	ItemID            string
	SKU               string `form_required:"1" form_section:"Specification" label:"SKU"`
	SpecVariantID     string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/specvariant/find|_id|Name"`
	SpecSizeID        string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/specsize/find|_id|Name"`
	SpecGradeID       string `form_required:"1" form_section:"Specification" form_lookup:"/tenant/specgrade/find|_id|Name"`
	SpecID            string
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
