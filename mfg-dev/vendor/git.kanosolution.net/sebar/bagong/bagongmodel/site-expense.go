package bagongmodel

import "git.kanosolution.net/sebar/fico/ficomodel"

type SiteExpense struct {
	LineNo          int
	ID              string
	Name            string  `label:"Expense Name"`
	ExpenseCategory string  `grid:"hide" form:"hide"`
	ExpenseTypeID   string  `grid_label:"Expense Type" form_lookup:"/tenant/expensetype/find|_id|Name"`
	UnitID          string  `form_lookup:"/tenant/unit/find|_id|Name"`
	Amount          float64 `grid_label:"Price Each"`
	Value           float64 `grid_label:"Qty"`
	TotalAmount     float64 `form_read_only:"1" grid_label:"Amount"`
	Notes           string
	Vendor          string                  `grid:"hide" form_lookup:"/tenant/vendor/find|_id|Name" label:"Vendor"`
	CashBankID      string                  `grid:"hide" grid_label:"Cash Bank" form_lookup:"/tenant/cashbank/find|_id|Name"`
	Urgent          bool                    `grid:"hide"`
	ApprovalStatus  ficomodel.JournalStatus `form_read_only:"1"`
	JournalID       string                  `grid_label:"Journal No" form_read_only:"1"`
	VoucherID       string                  `grid_label:"Voucher No" form_read_only:"1"`
}

type SiteExpenseTrayekGrid struct {
	ID          string
	Name        string  `label:"Expense Name"`
	Amount      float64 `grid_label:"Price Each"`
	TotalAmount float64 `form_read_only:"1" grid_label:"Amount"`
	Notes       string
}
type SiteExpenseTrayekReadGrid struct {
	LineNo         int                     `form_read_only:"1"`
	Name           string                  `label:"Expense Name"`
	Amount         float64                 `grid_label:"Price Each"`
	TotalAmount    float64                 `form_read_only:"1"  grid_label:"Amount"`
	Notes          string                  ``
	ApprovalStatus ficomodel.JournalStatus `form_read_only:"1"`
	JournalID      string                  `grid_label:"Journal No" form_read_only:"1"`
	VoucherID      string                  `grid_label:"Voucher No" form_read_only:"1"`
}

type SiteExpenseGrid struct {
	ExpenseTypeID string  `grid_label:"Expense Type" form_lookup:"/tenant/expensetype/find|_id|Name"`
	Name          string  `label:"Expense Name"`
	Amount        float64 `grid_label:"Price Each"`
	Value         float64 `grid_label:"Qty"`
	TotalAmount   float64 `form_read_only:"1" grid_label:"Amount"`
	Notes         string
}

type SiteExpenseReadGrid struct {
	LineNo         int                     `form_read_only:"1"`
	ExpenseTypeID  string                  `form_read_only:"1" grid_label:"Expense Type" form_lookup:"/tenant/expensetype/find|_id|Name"`
	Name           string                  `label:"Expense Name"`
	Amount         float64                 `grid_label:"Price Each"`
	Value          float64                 `grid_label:"Qty"`
	TotalAmount    float64                 `form_read_only:"1" grid_label:"Amount"`
	Notes          string                  ``
	ApprovalStatus ficomodel.JournalStatus `form_read_only:"1"`
	JournalID      string                  `form_read_only:"1" grid_label:"Journal No"  `
	VoucherID      string                  `form_read_only:"1" grid_label:"Voucher No" `
}
