package ficomodel

import (
	"fmt"

	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
)

type JournalLine struct {
	//JournalID        string
	LineNo           int `form_read_only:"1"`
	Account          SubledgerAccount
	OffsetAccount    SubledgerAccount
	OffsetTransRefID string
	TagObjectID1     SubledgerAccount
	TagObjectID2     SubledgerAccount
	CurrencyID       string
	LedgerDirection  tenantcoremodel.LedgerDirection
	TrxType          string
	PaymentType      string
	Qty              float64
	UnitID           string
	PriceEach        float64
	Amount           float64
	DiscountType     string
	Discount         float64
	ApproveAmount    float64
	Text             string
	Critical         bool
	Taxable          bool
	TaxCodes         []string
	Locked           bool
	Ignore           bool
	ChequeGiroID     string
	ChecklistTemp    tenantcoremodel.Checklists
	References       tenantcoremodel.References
	Dimension        tenantcoremodel.Dimension
	PPN              float64
	PPH              float64
}

type JournalLines []JournalLine

func (lines JournalLines) Update(lineNo int, updater func(line *JournalLine)) ([]JournalLine, error) {
	_, ok := lo.Find(lines, func(l JournalLine) bool {
		return l.LineNo == lineNo
	})
	if !ok {
		return nil, fmt.Errorf("not found: journal line: %d", lineNo)
	}

	lines = lo.Map(lines, func(l JournalLine, index int) JournalLine {
		if l.LineNo == lineNo {
			updater(&l)
		}
		return l
	})

	return lines, nil
}
