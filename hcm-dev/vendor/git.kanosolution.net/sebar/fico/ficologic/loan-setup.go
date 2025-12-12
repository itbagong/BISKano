package ficologic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
)

type LoanSetupHandler struct {
}

func (obj *LoanSetupHandler) GenerateLoan(ctx *kaos.Context, fy *ficomodel.LoanSetup) (*ficomodel.LoanSetup, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	installmentAmount := fy.LoanAmount / float64(fy.Period)
	remainingLoan := fy.LoanAmount
	autodebetStartDate := fy.AutodebetStartDate
	autodebetStartDateTemp := fy.AutodebetStartDate
	isNextMonth := false
	lineLoans := []ficomodel.LoanLine{}
	for i := 0; i < fy.Period; i++ {
		line := ficomodel.LoanLine{}
		remainingLoan = remainingLoan - installmentAmount

		line.Period = i + 1
		line.InstallmentAmount = installmentAmount
		line.RemainingLoan = remainingLoan
		line.Status = ficomodel.LoanUnpaid

		// cek date
		y, m, _ := autodebetStartDate.Date()
		// firstDay := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
		// lastDay := time.Date(y, m+1, 1, 0, 0, 0, -1, time.UTC)
		// firstDayNextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC)
		lastDayNextMonth := time.Date(y, m+2, 1, 0, 0, 0, -1, time.UTC)
		if isNextMonth {
			autodebetStartDate = autodebetStartDateTemp
			isNextMonth = false
		} else {
			autodebetStartDateTemp = autodebetStartDate
			autodebetStartDate = autodebetStartDate.AddDate(0, 1, 0)
		}

		if autodebetStartDate.Before(lastDayNextMonth) || autodebetStartDate.Equal(lastDayNextMonth) {
			line.Date = autodebetStartDate
		} else {
			line.Date = lastDayNextMonth
			autodebetStartDateTemp = autodebetStartDateTemp.AddDate(0, 2, 0)
			isNextMonth = true
		}

		lineLoans = append(lineLoans, line)
	}

	// update line loan
	fy.Lines = lineLoans
	h.Save(fy)

	return fy, nil
}
