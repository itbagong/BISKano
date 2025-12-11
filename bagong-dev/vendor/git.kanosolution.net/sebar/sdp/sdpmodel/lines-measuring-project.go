package sdpmodel

type LinesMeasuringProject struct {
	Budget        string `grid:"hide" form_section:"General" form_required:"1" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=MB|_id|Name"` // Create master budget from dynamic master and get value from there
	Year          int    `grid:"hide" form_kind:"number"`
	LedgerAccount string `form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	Month         map[string]float64
}
