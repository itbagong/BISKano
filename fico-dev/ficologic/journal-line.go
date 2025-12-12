package ficologic

import (
	"fmt"

	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

func JournalLinesToLedgerTrxs(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.CashSchedule, []*ficomodel.LedgerTransaction, error) {
	css := []*ficomodel.CashSchedule{}
	lts := []*ficomodel.LedgerTransaction{}
	expTypes := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.ExpenseType))

	mul := lo.Ternary(negate, float64(-1), float64(1))
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))

	buildLedgerTrx := func(line ficomodel.JournalLine, accountType tenantcoremodel.TrxModule, accountID string, amt float64, offset bool) error {
		ledgerAccountID := accountID
		if accountType == ficomodel.SubledgerExpense {
			expType, err := expTypes.Get(accountID)
			if err != nil {
				return fmt.Errorf("invalid expense: %s, line %d: %s", accountID, line.LineNo, err.Error())
			}
			ledgerAccountID = expType.LedgerAccountID
		}

		coa, err := coas.Get(ledgerAccountID)
		if err != nil {
			return fmt.Errorf("invalid ledger account: %s: %s", accountID, err.Error())
		}
		lt := &ficomodel.LedgerTransaction{
			Dimension:     line.Dimension,
			Account:       *coa,
			CurrencyID:    line.CurrencyID,
			Amount:        amt,
			Status:        ficomodel.AmountConfirmed,
			Text:          line.Text,
			SourceTrxType: line.TrxType,
			SourceLineNo:  line.LineNo,
			Offset:        offset,
		}
		lts = append(lts, lt)

		if accountType == ficomodel.SubledgerExpense {
			cs := &ficomodel.CashSchedule{
				Account:   ficomodel.NewSubAccount(ficomodel.SubledgerExpense, accountID),
				Amount:    amt,
				Status:    ficomodel.CashScheduled,
				Text:      line.Text,
				Direction: ficomodel.CashExpense,
			}
			css = append(css, cs)
		}
		return nil
	}

	for _, line := range lines {
		if line.Account.IsValid(ficomodel.SubledgerAccounting) || line.Account.IsValid(ficomodel.SubledgerExpense) {
			if err := buildLedgerTrx(line, line.Account.AccountType, line.Account.AccountID, mul*line.Amount, false); err != nil {
				return nil, nil, err
			}
		}

		if line.OffsetAccount.IsValid(ficomodel.SubledgerAccounting) || line.OffsetAccount.IsValid(ficomodel.SubledgerExpense) {
			if err := buildLedgerTrx(line, line.OffsetAccount.AccountType, line.OffsetAccount.AccountID, -mul*line.Amount, true); err != nil {
				return nil, nil, err
			}
		}
	}

	return css, lts, nil
}

func JournalLinesToCashTrxs(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.CashTransaction, error) {
	var (
		cts0 []*ficomodel.CashTransaction
	)

	cts := []*ficomodel.CashTransaction{}
	cbs := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.CashBank))
	cbgs := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.CashBankGroup))

	buildTrx := func(line ficomodel.JournalLine, accountID string, amt float64) error {
		ct, err := reflector.CopyAttributes(line, new(ficomodel.CashTransaction))
		if err != nil {
			return fmt.Errorf("journal line to cash: %s", err.Error())
		}
		cb, err := cbs.Get(accountID)
		if err != nil {
			return fmt.Errorf("missing: cash bank: %s: %s", line.Account.AccountID, err.Error())
		}
		if cb.MainBalanceAccount == "" {
			cbg, _ := cbgs.Get(cb.CashBankGroupID)
			if cbg.MainBalanceAccount == "" {
				return fmt.Errorf("missing: cash bank main balance account: %s, %s", cb.ID, cb.Name)
			}
			cb.MainBalanceAccount = cbg.MainBalanceAccount
		}
		ct.CashBank = *cb
		cts = append(cts, ct)

		return nil
	}

	mul := float64(1)
	if negate {
		mul = -1
	}
	for _, line := range lines {
		if line.Account.AccountType == ficomodel.SubledgerCashBank && line.Account.AccountID != "" {
			if err := buildTrx(line, line.Account.AccountID, mul*line.Amount); err != nil {
				return cts0, err
			}
		}

		if line.OffsetAccount.AccountType == ficomodel.SubledgerCashBank && line.OffsetAccount.AccountID != "" {
			if err := buildTrx(line, line.OffsetAccount.AccountID, -mul*line.Amount); err != nil {
				return cts0, err
			}
		}
	}
	return cts, nil
}

func JournalLineToCustomerTrxs(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.LedgerTransaction, []*ficomodel.CustomerTransaction, error) {
	var err error

	lts := []*ficomodel.LedgerTransaction{}
	cts := []*ficomodel.CustomerTransaction{}
	lts0 := []*ficomodel.LedgerTransaction{}
	cts0 := []*ficomodel.CustomerTransaction{}

	mul := lo.Ternary(negate, float64(-1), float64(1))
	custs := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Customer))
	custGroups := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.CustomerGroup))
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))

	buildTrx := func(line ficomodel.JournalLine, accountID string, amt float64) error {
		ct, err := reflector.CopyAttributes(line, new(ficomodel.CustomerTransaction))
		if err != nil {
			return err
		}
		cust, err := custs.Get(accountID)
		if err != nil {
			return fmt.Errorf("invalid: customer: %s: %s", accountID, err.Error())
		}
		ct.Customer = *cust
		if ct.Customer.Setting.MainBalanceAccount == "" {
			cg, err := custGroups.Get(ct.Customer.GroupID)
			if err == nil {
				ct.Customer.Setting.MainBalanceAccount = cg.Setting.MainBalanceAccount
			}
		}
		ct.Amount = amt
		cts = append(cts, ct)

		if ct.Customer.Setting.MainBalanceAccount == "" {
			return fmt.Errorf("invalid customer ledger account: customer %s: %s", ct.Customer.ID, ct.Customer.Name)
		}

		coa, err := coas.Get(ct.Customer.Setting.MainBalanceAccount)
		if err != nil {
			return fmt.Errorf("invalid customer ledger account: %s: %s", ct.Customer.Setting.MainBalanceAccount, err.Error())
		}
		lt := &ficomodel.LedgerTransaction{
			CompanyID:     ct.CompanyID,
			Dimension:     ct.Dimension,
			Amount:        ct.Amount,
			Text:          ct.Text,
			SourceTrxType: line.TrxType,
			SourceLineNo:  line.LineNo,
			Account:       *coa,
			CurrencyID:    line.CurrencyID,
			Status:        ficomodel.AmountConfirmed,
		}
		lts = append(lts, lt)
		return nil
	}

	for _, line := range lines {
		if line.Account.IsValid(ficomodel.SubledgerCustomer) {
			if err = buildTrx(line, line.Account.AccountID, mul*line.Amount); err != nil {
				return lts0, cts0, fmt.Errorf("journal lines to customer trxs: %s", err.Error())
			}
		}

		if line.OffsetAccount.IsValid(ficomodel.SubledgerCustomer) {
			if err = buildTrx(line, line.OffsetAccount.AccountID, -mul*line.Amount); err != nil {
				return lts0, cts0, fmt.Errorf("journal lines to customer trxs: %s", err.Error())
			}
		}
	}

	return lts, cts, nil
}

func JournalLineToVendorTrxs(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.LedgerTransaction, []*ficomodel.VendorTransaction, error) {
	var (
		vts0 []*ficomodel.VendorTransaction
		lts0 []*ficomodel.LedgerTransaction
	)

	vts := []*ficomodel.VendorTransaction{}
	lts := []*ficomodel.LedgerTransaction{}

	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))
	vendors := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Vendor))

	buildTrx := func(line ficomodel.JournalLine, accountID string, amt float64) error {
		vt, err := reflector.CopyAttributes(line, new(ficomodel.VendorTransaction))
		if err != nil {
			return fmt.Errorf("journal line to cash: %s", err.Error())
		}
		vnd, err := vendors.Get(accountID)
		if err != nil {
			return fmt.Errorf("missing: cash bank: %s: %s", line.Account.AccountID, err.Error())
		}
		vt.Vendor = *vnd
		vts = append(vts, vt)

		if vt.Vendor.MainBalanceAccount == "" {
			return fmt.Errorf("missing: main balance ledger account for vendor")
		}

		coa, err := coas.Get(vt.Vendor.MainBalanceAccount)
		if err != nil {
			return fmt.Errorf("invalid account: %s: %s", vt.Vendor.MainBalanceAccount, err.Error())
		}
		lt, _ := reflector.CopyAttributes(vt, new(ficomodel.LedgerTransaction))
		lt.ID = ""
		lt.Account = *coa
		lt.Status = ficomodel.AmountConfirmed
		lts = append(lts, lt)

		return nil
	}

	mul := float64(1)
	if negate {
		mul = -1
	}
	for _, line := range lines {
		if line.Account.AccountType == ficomodel.SubledgerCashBank && line.Account.AccountID != "" {
			if err := buildTrx(line, line.Account.AccountID, mul*line.Amount); err != nil {
				return lts0, vts0, err
			}
		}

		if line.OffsetAccount.AccountType == ficomodel.SubledgerCashBank && line.OffsetAccount.AccountID != "" {
			if err := buildTrx(line, line.OffsetAccount.AccountID, -mul*line.Amount); err != nil {
				return lts0, vts0, err
			}
		}
	}
	return lts, vts, nil
}

func JournalLineToAssetTrxs(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.LedgerTransaction, []*ficomodel.AssetTransaction, error) {
	ledgderTrxs := []*ficomodel.LedgerTransaction{}
	assetTrxs := []*ficomodel.AssetTransaction{}

	return ledgderTrxs, assetTrxs, nil
}

func CashTrxToLedgerTrx(db *datahub.Hub, trxs []*ficomodel.CashTransaction) ([]*ficomodel.LedgerTransaction, error) {
	res := []*ficomodel.LedgerTransaction{}
	empty := []*ficomodel.LedgerTransaction{}
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))
	for _, trx := range trxs {
		coa, err := coas.Get(trx.CashBank.MainBalanceAccount)
		if err != nil {
			return empty, fmt.Errorf("invalid: ledger account: %s: %s", trx.CashBank.MainBalanceAccount, err.Error())
		}
		ltx, _ := reflector.CopyAttributes(trx, new(ficomodel.LedgerTransaction))
		ltx.ID = ""
		ltx.Account = *coa
		res = append(res, ltx)
	}
	return res, nil
}

func CustomerTrxToLedgerTrx(db *datahub.Hub, trxs []*ficomodel.CustomerTransaction) ([]*ficomodel.LedgerTransaction, error) {
	res := []*ficomodel.LedgerTransaction{}
	empty := []*ficomodel.LedgerTransaction{}
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))
	for _, trx := range trxs {
		coa, err := coas.Get(trx.Customer.Setting.MainBalanceAccount)
		if err != nil {
			return empty, fmt.Errorf("invalid: ledger account: %s: %s", trx.Customer.Setting.MainBalanceAccount, err.Error())
		}
		ltx, _ := reflector.CopyAttributes(trx, new(ficomodel.LedgerTransaction))
		ltx.ID = ""
		ltx.Account = *coa
		res = append(res, ltx)
	}
	return res, nil
}

func VendorTrxToLedgerTrx(db *datahub.Hub, trxs []*ficomodel.VendorTransaction) ([]*ficomodel.LedgerTransaction, error) {
	res := []*ficomodel.LedgerTransaction{}
	empty := []*ficomodel.LedgerTransaction{}
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))
	for _, trx := range trxs {
		coa, err := coas.Get(trx.Vendor.MainBalanceAccount)
		if err != nil {
			return empty, fmt.Errorf("invalid: ledger account: %s: %s", trx.Vendor.MainBalanceAccount, err.Error())
		}
		ltx, _ := reflector.CopyAttributes(trx, new(ficomodel.LedgerTransaction))
		ltx.ID = ""
		ltx.Account = *coa
		res = append(res, ltx)
	}
	return res, nil
}

func AssetTrxToLedgerTrx(db *datahub.Hub, trxs []*ficomodel.AssetTransaction) ([]*ficomodel.LedgerTransaction, error) {
	res := []*ficomodel.LedgerTransaction{}
	empty := []*ficomodel.LedgerTransaction{}
	coas := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.LedgerAccount))
	for _, trx := range trxs {
		var (
			coa             *tenantcoremodel.LedgerAccount
			err             error
			ledgerAccountID string
		)
		switch trx.TrxType {
		case string(ficomodel.AssetAcquistion):
			ledgerAccountID = trx.Asset.AcquisitionAccount

		case string(ficomodel.AssetDeprecation):
			ledgerAccountID = trx.Asset.DepreciationAccount

		case string(ficomodel.AssetDisposal):
			ledgerAccountID = trx.Asset.DisposalAccount

		case string(ficomodel.AssetAdjustment):
			ledgerAccountID = trx.Asset.AdjustmentAccount
		}
		if ledgerAccountID == "" {
			return empty, fmt.Errorf("invalid: ledger account: %s: mandatory", trx.TrxType)
		}
		coa, err = coas.Get(ledgerAccountID)
		if err != nil {
			return empty, fmt.Errorf("invalid: ledger account: %s, %s: %s", trx.TrxType, ledgerAccountID, err.Error())
		}
		ltx, _ := reflector.CopyAttributes(trx, new(ficomodel.LedgerTransaction))
		ltx.ID = ""
		ltx.Account = *coa
		res = append(res, ltx)
	}
	return res, nil
}

func VendorTrxToCashSchedule(db *datahub.Hub, trxs []*ficomodel.VendorTransaction) ([]*ficomodel.CashSchedule, error) {
	res := []*ficomodel.CashSchedule{}

	for _, trx := range trxs {
		sched := NewCashSchedule(trx.CompanyID, tenantcoremodel.TrxModule(trx.SourceJournalType), trx.SourceJournalID, 0, -trx.Amount)
		sched.Expected = trx.TrxDate
		sched.Text = trx.Text
		sched.Direction = ficomodel.CashExpense
		sched.Status = ficomodel.CashScheduled
		res = append(res, sched)
	}

	return res, nil
}

func CashTrxToCashSchedule(db *datahub.Hub, trxs []*ficomodel.CashTransaction) ([]*ficomodel.CashSchedule, error) {
	res := []*ficomodel.CashSchedule{}

	for _, trx := range trxs {
		sched := NewCashSchedule(trx.CompanyID, tenantcoremodel.TrxModule(trx.SourceJournalType), trx.SourceJournalID, 0, trx.Amount)
		sched.Expected = trx.TrxDate
		sched.Text = trx.Text
		sched.Direction = ficomodel.CashReceive
		sched.Status = ficomodel.CashScheduled
		res = append(res, sched)
	}

	return res, nil
}

func JournalLinesToExpense(db *datahub.Hub, lines []ficomodel.JournalLine, negate bool) ([]*ficomodel.CashSchedule, error) {
	empty := []*ficomodel.CashSchedule{}
	res := []*ficomodel.CashSchedule{}

	buildTrx := func(line ficomodel.JournalLine, accountID string, amount float64) error {
		cs, err := reflector.CopyAttributes(line, new(ficomodel.CashSchedule))
		if err != nil {
			return err
		}
		cs.ID = ""
		cs.Amount = amount
		cs.Direction = lo.Ternary(amount > 0, ficomodel.CashReceive, ficomodel.CashExpense)
		cs.Status = ficomodel.CashScheduled
		cs.Account = ficomodel.NewSubAccount(ficomodel.SubledgerExpense, accountID)
		res = append(res, cs)
		return nil
	}

	mul := float64(1)
	if negate {
		mul = -1
	}

	for _, line := range lines {
		if line.Account.IsValid(ficomodel.SubledgerExpense) {
			if err := buildTrx(line, line.Account.AccountID, mul*line.Amount); err != nil {
				return empty, fmt.Errorf("create cash schedule from expense: line %d: %s", line.LineNo, err.Error())
			}
		}

		if line.OffsetAccount.IsValid(ficomodel.SubledgerExpense) {
			if err := buildTrx(line, line.OffsetAccount.AccountID, -mul*line.Amount); err != nil {
				return empty, fmt.Errorf("create cash schedule from expense: line %d: %s", line.LineNo, err.Error())
			}
		}
	}

	return res, nil
}
