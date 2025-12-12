package sdpmodel

type CompetitorOpportunity struct {
	// Competitor            string  `form_kind:"text"` // get from master competitor
	Competitor            string  `form_required:"1" label:"Competitor" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOCM|_id|_id,Name"`
	BiddingAmount         float64 `form_kind:"number"`
	Strength              string  `form_kind:"text" form_multi_row:"2"`
	Weakness              string  `form_kind:"text" form_multi_row:"2"`
	OpportunityPercentage string  `label:"Opportunity"  form_multi_row:"2"`
	Threats               string  `form_kind:"text" form_multi_row:"2"`
}
