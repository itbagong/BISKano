package ficologic

import (
	"time"

	"github.com/sebarcode/codekit"
)

type PeriodClosing interface {
	Close(companyID string, dt time.Time) (codekit.M, error)
}

func ClosePeriod(companyID string, closeDate time.Time, closer PeriodClosing) (codekit.M, error) {
	return closer.Close(companyID, closeDate)
}
