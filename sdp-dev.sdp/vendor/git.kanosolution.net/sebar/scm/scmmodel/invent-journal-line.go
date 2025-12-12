package scmmodel

import (
	"time"

	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type BatchSN struct {
	BatchID      string
	SerialNumber string
	Qty          float64
}

type InventJournalLine struct {
	LineNo       int                       `form_read_only_edit:"1" grid:"hide" form_read_only:"1" grid_read_only:"1" form_section:"General" form_section_size:"3" `
	ItemID       string                    `form_section:"General" form_lookup:"/tenant/item/find|_id|Name" label:"Item"`
	SKU          string                    `form_section:"General" label:"SKU" form_lookup:"/tenant/itemspec/find|SKU|SKU"`
	Qty          float64                   `form_section:"General"`
	RemainingQty float64                   `form_section:"General"`
	UnitID       string                    `form_section:"General" form_section:"General" form_lookup:"/tenant/unit/find|_id|Name" label:"UoM"`
	Text         string                    `form_section:"General" form:"hide" grid_label:"Description" form_label:"Description" grid:"hide"`
	UnitCost     float64                   `form_section:"General"`
	Remarks      string                    `form_section:"General"`
	BatchSerials []BatchSN                 `grid:"hide" form_section:"General"`
	Dimension    tenantcoremodel.Dimension `form_section:"Dimension"`
	InventDim    InventDimension           `grid:"hide" form_section:"InventDim"`
}

type InventTrxLine struct {
	InventJournalLine
	JournalID   string
	TrxType     InventTrxType
	TrxDate     time.Time
	Item        *tenantcoremodel.Item
	InventQty   float64
	CostPerUnit float64
}
