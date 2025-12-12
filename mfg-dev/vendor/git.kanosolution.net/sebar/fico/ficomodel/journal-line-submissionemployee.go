package ficomodel

type SubmissionemployeeJournalLineGrid struct {
	Account   SubledgerAccount
	PriceEach float64
	Qty       float64
	Amount    float64 `form_read_only:"1"`
	Text      string
	Critical  bool
}
