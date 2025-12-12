package bagongmodel

type SiteIncome struct {
	LineNo         int
	ID             string
	Name           string
	Amount         float64
	Notes          string
	CashBankID     string `grid_label:"Cash Bank" form_lookup:"/tenant/cashbank/find|_id|Name" grid:"hide"`
	ApprovalStatus string `grid_label:"Approval Status"  form_read_only:"1"`
	JournalID      string `grid_label:"Journal No"  form_read_only:"1"`
	VoucherID      string `grid_label:"Voucher No"  form_read_only:"1"`
}

type SiteIncomeGrid struct {
	Name   string
	Amount float64
	Notes  string
}

type SiteIncomeReadGrid struct {
	LineNo         int     `form_read_only:"1"`
	Name           string  ``
	Amount         float64 ``
	Notes          string  ``
	ApprovalStatus string  `grid_label:"Approval Status"  form_read_only:"1"`
	JournalID      string  `grid_label:"Journal No"  form_read_only:"1"`
	VoucherID      string  `grid_label:"Voucher No"  form_read_only:"1"`
}
