package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type InventTrxReceipt struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                     `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General1" form_section_size:"4"`
	TrxDate           time.Time                  `form_section:"General1"`
	SourceType        tenantcoremodel.TrxModule  `form_section:"General1"`
	SourceTrxType     string                     `form_section:"General1"`
	SourceJournalID   string                     `form_section:"General1"`
	SourceLineNo      int                        `form_section:"General2"`
	Item              tenantcoremodel.Item       `form_section:"General2"`
	SKU               string                     `form_section:"General2" label:"SKU"`
	SKUName           string                     `form_section:"General2" label:"SKU"`
	Text              string                     `form_section:"General2"`
	Qty               float64                    `form_section:"General3"`
	TrxQty            float64                    `form_section:"General3"`
	TrxUnitID         string                     `form_section:"General3"`
	SettledQty        float64                    `form_section:"General3"` // sum of TrxQty dari source yang sama dengan status=confirmed
	OriginalQty       float64                    `form_section:"General3"` // SettledQty + TrxQty
	VendorID          string                     `form_section:"General3"`
	VendorName        string                     `form_section:"General3"`
	InventJournalLine InventJournalLine          `form:"hide" grid:"hide" json:"InventJournalLine" bson:"InventJournalLine"`
	InventDim         InventDimension            `form_section:"Dimension1"  grid:"hide"`
	Dimension         tenantcoremodel.Dimension  `form_section:"Dimension2"  grid:"hide"`
	References        tenantcoremodel.References `form:"hide" grid:"hide"`
	UnitCost          float64                    `form:"hide" grid:"hide"`
	ItemName          string                     `form:"hide" grid:"hide"`
}
