package ficomodel

type SubmissionemployeeJournalLineGrid struct {
	TagObjectID1 SubledgerAccount `label:"Employee"`
	Account      SubledgerAccount
	Qty          float64
	UnitID       string `form_lookup:"/tenant/unit/find|_id|Name" form_allow_add:"1"`
	PriceEach    float64
	DiscountType string  `form_items:"fixed|percent"`
	Discount     float64 `label:"Discount"`
	Amount       float64 `form_read_only:"1"`
	Text         string
	Critical     bool
	Taxable      bool
	PPN          float64 `grid_label:"PPN" form_read_only:"1"`
	PPH          float64 `grid_label:"PPH" form_read_only:"1"`
}
