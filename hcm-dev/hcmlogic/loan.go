package hcmlogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type LoanHandler struct {
}

func (obj *LoanHandler) GenerateLoan(ctx *kaos.Context, l *hcmmodel.Loan) (*hcmmodel.Loan, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	remainingLoan := l.ApprovedLoan
	lineLoans := make([]hcmmodel.LoanLine, l.ApprovedLoanPeriod)
	for i := 0; i < l.ApprovedLoanPeriod; i++ {
		line := hcmmodel.LoanLine{}
		remainingLoan -= l.ApprovedInstallment

		line.Period = i + 1
		line.InstallmentAmount = l.ApprovedInstallment
		line.RemainingLoan = remainingLoan
		line.Status = hcmmodel.LoanUnpaid
		line.Date = l.AutoDebitStartDate.AddDate(0, i, 0)

		lineLoans[i] = line
	}

	// update line loan
	l.Lines = lineLoans
	h.Save(l)

	return l, nil
}
