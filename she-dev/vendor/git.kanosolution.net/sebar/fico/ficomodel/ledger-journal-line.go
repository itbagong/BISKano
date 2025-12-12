package ficomodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type LedgerJournalLineGrid struct {
	Account       SubledgerAccount
	Text          string
	Debit         float64
	Credit        float64
	Amount        float64 `grid:"hide" form:"hide"`
	OffsetAccount SubledgerAccount
	//Dimension     tenantcoremodel.Dimension
}

type LedgerJournalLineForm struct {
	References tenantcoremodel.References
	Dimension  tenantcoremodel.Dimension
}
