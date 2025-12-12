package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ItemType string
type CostUnitCalcMethod string

const (
	ItemStock   ItemType = "STOCK"
	ItemVirtual ItemType = "VIRTUAL"
	ItemService ItemType = "SERVICE"
	ItemMakloon ItemType = "MAKLOON"

	CostUnitCalcMethodManual CostUnitCalcMethod = "MANUAL"
	CostUnitCalcMethodStdAvg CostUnitCalcMethod = "STD_AVERAGE"
	CostUnitCalcMethodFIFO   CostUnitCalcMethod = "FIFO"
	CostUnitCalcMethodLIFO   CostUnitCalcMethod = "LIFO"
)

type Item struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	ID                   string             `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"3"`
	Name                 string             `form_required:"1" form_section:"General"`
	OtherName            string 			`form_section:"General"`
	ItemType             ItemType           `form_required:"1" form_section:"General" form_items:"STOCK|VIRTUAL|SERVICE|MAKLOON"`
	ItemGroupID          string             `form_section:"General" form_lookup:"/tenant/itemgroup/find|_id|Name"`
	LedgerAccountIDStock string             `form_required:"1" form_section:"General" form_lookup:"/tenant/ledgeraccount/find|_id|Name"`
	DefaultUnitID        string             `form_required:"1" form_section:"General" form_lookup:"/tenant/unit/find|_id|Name"`
	CostUnitCalcMethod   CostUnitCalcMethod `form_required:"1" form_section:"General" form_items:"MANUAL|STD_AVERAGE|FIFO|LIFO"`
	CostUnit             float64            `label:"Unit cost"`
	PhysicalDimension    ItemDimensionCheck `form_section:"Physical Dimension" grid:"hide" form_section_auto_col:"2" form_section_show_title:"1"`
	FinanceDimension     ItemDimensionCheck `form_section:"Finance Dimension" grid:"hide" form_section_auto_col:"2" form_section_show_title:"1"`
	Created              time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type ItemDimensionCheck struct {
	IsEnabledSpecVariant       bool
	IsEnabledSpecSize          bool
	IsEnabledSpecGrade         bool
	IsEnabledItemBatch         bool
	IsEnabledItemSerial        bool
	IsEnabledLocationWarehouse bool
	IsEnabledLocationAisle     bool
	IsEnabledLocationSection   bool
	IsEnabledLocationBox       bool
}

func (o *Item) TableName() string {
	return "Items"
}

func (o *Item) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Item) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Item) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Item) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Item) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Item) PostSave(dbflex.IConnection) error {
	return nil
}
