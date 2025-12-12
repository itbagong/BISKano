package ficologic

import (
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/codekit"
)

func CalcTax(setup ficomodel.TaxSetup, linesRef *[]ficomodel.JournalLine, injectReferences bool) ([]float64, float64, error) {
	var err error

	taxLines := make([]float64, len(*linesRef))
	taxSum := float64(0)
	baseSum := float64(0)

	for index, line := range *linesRef {
		if !line.Taxable || !codekit.HasMember(line.TaxCodes, setup.ID) {
			continue
		}

		if setup.CalcMethod == ficomodel.TaxCalcHeader {
			baseSum += line.Amount
			continue
		}
		taxAmt := setup.Rate * line.Amount
		taxLines[index] = taxAmt
		taxSum += taxAmt

		if injectReferences {
			if line.References == nil {
				line.References = tenantcoremodel.References{}
			}
			line.References = line.References.Set("Tax_"+setup.ID, taxAmt)
		}
		(*linesRef)[index] = line
	}

	if setup.CalcMethod == ficomodel.TaxCalcHeader {
		taxSum = setup.Rate * baseSum
	}

	return taxLines, taxSum, err
}
