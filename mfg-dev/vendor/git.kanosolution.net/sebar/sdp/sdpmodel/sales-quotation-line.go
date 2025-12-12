package sdpmodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type SalesQuotationLineForm struct {
	Taxable   bool                      `form_section:"Tax" form_section_show_title:"1" form_read_only:"1"`
	TaxCodes  []string                  `form_section:"Tax" form_lookup:"/fico/taxsetup/find|_id|Name"`
	Dimension tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1" form_read_only:"1"`
}

type SalesQuotationLineGrid struct {
	Asset          string                    `form_lookup:"/tenant/asset/find|_id|Name"`
	Item           string                    `form_kind:"text" label:"Item Variant"`
	Shift          int64                     `form_kind:"number"`
	Description    string                    `form_kind:"text" form_multi_row:"1"`
	ContractPeriod int                       `form_kind:"number"`
	Uom            tenantcoremodel.UoM       `form_lookup:"/tenant/unit/find|_id|Name" form_allow_add:"1"`
	Qty            uint                      `form_kind:"number"`
	UnitPrice      uint64                    `form_kind:"number"`
	Amount         uint64                    `form_kind:"number" form_read_only:"1"`
	DiscountType   DiscountTypeSQ            `form_items:"fixed|percentage"`
	Discount       int                       `form_kind:"number" grid_label:"Discount"`
	Taxable        bool                      `form_kind:"checkbox" form_read_only:"1"`
	Dimension      tenantcoremodel.Dimension `form_read_only:"1"`
}

type SalesQuotationLinePreviewGrid struct {
	Item           float64
	Description    float64
	Qty            float64
	Uom            float64
	ContractPeriod float64
	UnitPrice      float64
	Amount         float64
}

type SalesQuotationLine struct {
	Asset          string
	Item           string
	Description    string
	Shift          int64
	Qty            uint
	UoM            string
	ContractPeriod int
	UnitPrice      int64
	Amount         int64
	DiscountType   DiscountTypeSQ
	Discount       int
	Taxable        bool
	TaxCodes       []string
	Spesifications []string
}
