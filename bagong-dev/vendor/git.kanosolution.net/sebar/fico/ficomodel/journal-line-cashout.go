package ficomodel

type CashOutJournalLineGrid struct {
	Account      SubledgerAccount
	PriceEach    float64
	Qty          float64 `grid:"hide" form:"hide"`
	Amount       float64 `form_read_only:"1"`
	Text         string
	PaymentType  string `form_lookup:"/tenant/masterdata/find?MasterDataTypeID=PTY|_id|Name"`
	ChequeGiroID string `form_read_only:"1" grid_label:"Payment ID"`
	Ignore       bool   `form_read_only:"1"`
}
