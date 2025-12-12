package ficologic_test

import (
	"testing"

	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/smartystreets/goconvey/convey"
)

func TestCalcTax(t *testing.T) {
	convey.Convey("test tax line", t, func() {
		setup := ficomodel.TaxSetup{
			ID:         "PPN",
			Rate:       0.11,
			CalcMethod: ficomodel.TaxCalcLine,
		}

		lines := []ficomodel.JournalLine{
			{Amount: 1000, Taxable: true, TaxCodes: []string{"PPN"}},
			{Amount: 2000, Taxable: true, TaxCodes: []string{"PPN"}},
			{Amount: 500},
		}

		_, sum, err := ficologic.CalcTax(setup, &lines, true)
		convey.So(err, convey.ShouldBeNil)
		convey.So(sum, convey.ShouldEqual, 3000*setup.Rate)

		total := float64(0)
		for _, line := range lines {
			total += line.References.ToM().GetFloat64("Tax_PPN")
		}
		convey.So(total, convey.ShouldEqual, 3000*setup.Rate)
	})
}
