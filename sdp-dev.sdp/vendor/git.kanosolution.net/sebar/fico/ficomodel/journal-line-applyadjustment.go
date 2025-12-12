package ficomodel

type ApplyAdjustmentJournalLineForm struct {
	Text    string `form_label:"Name" form_required:"1"`
	Account SubledgerAccount
	TrxType string `form:"hide"`
	Amount  float64
}
