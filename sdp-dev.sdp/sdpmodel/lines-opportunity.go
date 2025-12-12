package sdpmodel

type LinesOpportunity struct {
	// Item           string `form_kind:"text"` // get from master item type virtual
	// Description    string `form_kind:"text"`
	// Quantity       int    `form_kind:"number"`
	// UoM            string `form_kind:"text"` // get from master UoM
	// ContractPeriod int    `form_kind:"number"`

	Asset          string   `form_lookup:"/tenant/asset/find|_id|Name"`
	Item           string   `form_kind:"text" form_lookup:"/tenant/item/find|_id|Name"`
	Spesifications []string `form_lookup:"/tenant/specvariant/find|_id|Name" label:"Specification"`
	Description    string   `form_kind:"text" form_multi_row:"1"`
	ContractPeriod int      `form_kind:"number"`
	Uom            string   `label:"UoM" form_kind:"text" form_lookup:"/tenant/unit/find|_id|Name"`
	Qty            uint     `form_kind:"number"`
}
