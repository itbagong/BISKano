package tenantcoremodel

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemGroup struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string             `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name                 string             `form_required:"1" form_section:"General"`
	Alias                string             `form_section:"General" grid:"hide"`
	LastNo               int                `grid:"hide"` // for item numbering
	ItemType             ItemType           `form_required:"1" form_read_only_edit:"1" form_section:"General" form_items:"STOCK|VIRTUAL|SERVICE|MAKLOON"`
	LedgerAccountIDStock string             `form_required:"1" form_section:"General" form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	DefaultUnitID        string             `form_required:"1" form_section:"General" form_lookup:"/tenant/unit/find|_id|Name"`
	CostUnitCalcMethod   CostUnitCalcMethod `form_required:"1" form_section:"General" form_items:"MANUAL|STD_AVERAGE|FIFO|LIFO"`
	CostUnit             float64            `label:"Unit cost"`
	PhysicalDimension    ItemDimensionCheck `form_section:"Physical Dimension" grid:"hide" label:"Physical Dimension" form_section_size:"1" form_section_show_title:"1"`
	FinanceDimension     ItemDimensionCheck `form_section:"Finance Dimension" grid:"hide" label:"Finance Dimension" form_section_size:"1" form_section_show_title:"1"`
	Created              time.Time          `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time          `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ItemGroup) TableName() string {
	return "ItemGroups"
}

func (o *ItemGroup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ItemGroup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ItemGroup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ItemGroup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ItemGroup) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ItemGroup) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *ItemGroup) FormatItemID() string {
	outFormat := "[Alias]%06d"

	o.LastNo = o.LastNo + 1
	itemID := fmt.Sprintf(outFormat, o.LastNo)
	itemID = strings.ReplaceAll(itemID, "[Alias]", lo.Ternary(o.Alias != "", o.Alias, ""))

	return itemID
}
