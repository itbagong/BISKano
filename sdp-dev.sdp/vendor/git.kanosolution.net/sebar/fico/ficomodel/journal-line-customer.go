package ficomodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type CustomerJournalLineGrid struct {
	TagObjectID1 SubledgerAccount `grid_label:"Asset"`
	Text         string
	Qty          float64
	UnitID       string `form_lookup:"/tenant/unit/find|_id|Name" form_allow_add:"1"`
	PriceEach    float64
	DiscountType string  `form_items:"fixed|percent"`
	Discount     float64 `label:"Discount"`
	Amount       float64 `form_read_only:"1"`
	Account      SubledgerAccount
	Taxable      bool
	Dimension    tenantcoremodel.Dimension
}

func (o *CustomerJournalLineGrid) FromJournalLine(l *JournalLine) *CustomerJournalLineGrid {
	return o
}

func (o *CustomerJournalLineGrid) ToJournalLine() *JournalLine {
	line := new(JournalLine)
	return line
}

type CustomerJournalLineForm struct {
	/*
		TagObjectID1  SubledgerAccount `form_section:"General" form_section_show_title:"1" form_section_size:"3"`
		TagObjectID2  SubledgerAccount
		Text          string
		OffsetAccount SubledgerAccount           `form_lookup:"/tenant/ledgeraccount/find|_id|Name"`
		Qty           float64                    `form_section:"Amount" form_section_show_title:"1"`
		UnitID        string                     `form_section:"Amount" form_lookup:"/tenant/uom/find|_id|Name" form_allow_add:"1"`
		PriceEach     float64                    `form_section:"Amount"`
		Amount        float64                    `form_section:"Amount" form_read_only:"1"`
	*/
	Taxable   bool                      `form_section:"Tax" form_section_show_title:"1" form_section_size:"2"`
	TaxCodes  []string                  `form_section:"Tax" form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1"`
	Dimension tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1"`
}

func (o *CustomerJournalLineForm) FromJournalLine(l *JournalLine) *CustomerJournalLineForm {
	return o
}

func (o *CustomerJournalLineForm) ToJournalLine() *JournalLine {
	line := new(JournalLine)
	return line
}
