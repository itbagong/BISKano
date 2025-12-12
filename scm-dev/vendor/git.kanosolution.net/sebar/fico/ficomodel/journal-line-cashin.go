package ficomodel

type CashInJournalLineGrid struct {
	Account   SubledgerAccount
	PriceEach float64
	Qty       float64 `grid:"hide" form:"hide"`
	Amount    float64 `form_read_only:"1"`
	Text      string
	Ignore    bool `form_read_only:"1"`
}
