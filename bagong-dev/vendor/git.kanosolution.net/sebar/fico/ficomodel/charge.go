package ficomodel

type Charge struct {
	ChargeCode      string
	CurrencyID      string
	LedgerAccountID string
	Rate            float64
	Amount          float64
}
