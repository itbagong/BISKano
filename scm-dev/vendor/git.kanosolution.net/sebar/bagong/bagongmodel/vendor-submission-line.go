package bagongmodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type VendorSubmissionLineGrid struct {
	TagObjectID1  SubledgerAccount
	TagObjectID2  SubledgerAccount
	Text          string
	Qty           float64
	UnitID        string `form_lookup:"/tenant/uom/find|_id|Name" form_allow_add:"1"`
	PriceEach     float64
	Amount        float64 `form_read_only:"1"`
	OffsetAccount SubledgerAccount
	Taxable       bool
	Dimension     tenantcoremodel.Dimension
}

func (o *VendorSubmissionLineGrid) FromJournalLine(l *JournalLine) *VendorSubmissionLineGrid {
	return o
}

func (o *VendorSubmissionLineGrid) ToJournalLine() *JournalLine {
	line := new(JournalLine)
	return line
}

type VendorSubmissionLineForm struct {
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
	Taxable    bool                       `form_section:"Tax" form_section_show_title:"1" form_section_size:"2"`
	TaxCodes   []string                   `form_section:"Tax" form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1"`
	Dimension  tenantcoremodel.Dimension  `form_section:"Dimension" form_section_show_title:"1"`
	References tenantcoremodel.References `form_section:"References" form_section_show_title:"1"`
}

func (o *VendorSubmissionLineForm) FromJournalLine(l *JournalLine) *VendorSubmissionLineForm {
	return o
}

func (o *VendorSubmissionLineForm) ToJournalLine() *JournalLine {
	line := new(JournalLine)
	return line
}
