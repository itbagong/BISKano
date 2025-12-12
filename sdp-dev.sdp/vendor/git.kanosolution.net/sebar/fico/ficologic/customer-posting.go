package ficologic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type customerPosting struct {
	opt     *PostingHubCreateOpt
	header  *ficomodel.CustomerJournal
	trxType string
	jt      *ficomodel.CustomerJournalType

	customer *tenantcoremodel.Customer
	coas     *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	expenses *sebar.MapRecord[*tenantcoremodel.ExpenseType] //TODO: samakan dengan model expense
}

func NewCustomerPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.CustomerJournal, ficomodel.JournalLine] {
	c := new(customerPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	pvd := PostingProvider[*ficomodel.CustomerJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *customerPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.CustomerJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lines
}

func (p *customerPosting) Header() (*ficomodel.CustomerJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(ficomodel.CustomerJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.CustomerJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}
	p.jt = jt
	p.trxType = jt.TransactionType
	if p.trxType == "" {
		p.trxType = "CustomerJournal"
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, jt.PostingProfileID)
	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile")
	}
	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.customer, err = datahub.GetByID(p.opt.Db, new(tenantcoremodel.Customer), j.CustomerID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: customer: %s", j.CustomerID)
	}
	customerGroup, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.CustomerGroup), p.customer.GroupID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: customer group id: %s", p.customer.GroupID)
	}
	p.customer.Dimension = tenantcorelogic.TernaryDimension(p.customer.Dimension, customerGroup.Dimension)
	p.customer.Setting.DepositAccount = tenantcorelogic.TernaryString(p.customer.Setting.DepositAccount, customerGroup.Setting.DepositAccount)
	p.customer.Setting.MainBalanceAccount = tenantcorelogic.TernaryString(p.customer.Setting.MainBalanceAccount, customerGroup.Setting.MainBalanceAccount)
	if j.DefaultOffset.AccountID == "" {
		j.DefaultOffset = jt.DefaultOffset
	}

	p.header = j
	return j, pp, nil
}

func (p *customerPosting) Lines() ([]ficomodel.JournalLine, error) {
	if len(p.header.Lines) == 0 {
		return nil, fmt.Errorf("no lines")
	}

	lines := make([]ficomodel.JournalLine, len(p.header.Lines))
	for idx, line := range p.header.Lines {
		if line.LineNo == 0 {
			line.LineNo = idx + 1
		}
		if line.OffsetAccount.AccountID == "" {
			line.OffsetAccount = ficomodel.NewSubAccount(ficomodel.SubledgerCustomer, p.header.CustomerID)
		}
		if line.Account.AccountID == "" {
			line.Account = p.header.DefaultOffset
		}
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)
		line.TrxType = p.trxType
		lines[idx] = line
	}

	return lines, nil
}

func (p *customerPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.CustomerJournal, lines []ficomodel.JournalLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {

	preview := tenantcoremodel.PreviewReport{
		Header: make(codekit.M),
	}
	trxs := map[string][]orm.DataModel{}
	totalAmount := float64(0)

	custLedgerAccount, err := p.coas.Get(p.customer.Setting.MainBalanceAccount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("missing: customer ledger account: %s", p.customer.Setting.MainBalanceAccount)
	}

	// create customer transaction
	custTrxs := []*ficomodel.CustomerTransaction{}
	ledgTrxs := []orm.DataModel{}

	for _, line := range lines {
		var (
			expense       *tenantcoremodel.ExpenseType
			ledgerAccount *tenantcoremodel.LedgerAccount
			err           error
		)

		if !(line.Account.AccountType == ficomodel.SubledgerAccounting || line.Account.AccountType == ficomodel.SubledgerExpense) {
			return nil, nil, 0, fmt.Errorf("line %d: allowed account type only Ledger and Expense", line.LineNo)
		}

		coaID := line.Account.AccountID
		if line.Account.AccountType == ficomodel.SubledgerExpense {
			expense, ledgerAccount, err = GetSubledgerFromMapRecord(line.Account.AccountID, "LedgerAccountID", p.expenses, p.coas)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: expense: %s: %s", line.Account.AccountID, err.Error())
			}
			coaID = expense.LedgerAccountID
		} else {
			ledgerAccount, err = p.coas.Get(line.Account.AccountID)
		}
		if err != nil {
			return nil, nil, 0, fmt.Errorf("missing: ledger account: %s, line %d", coaID, line.LineNo)
		}
		ct := &ficomodel.CustomerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerCustomer.String(),
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Customer:          *p.customer,
			Status:            ficomodel.AmountConfirmed,
			Amount:            line.Amount,
		}

		lt := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerCustomer,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            -line.Amount,
		}

		ltCustomer := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerCustomer,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Account:           *custLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            line.Amount,
		}

		custTrxs = append(custTrxs, ct)
		ledgTrxs = append(ledgTrxs, lt, ltCustomer)
		totalAmount += line.Amount
	}

	trxs[ledgTrxs[0].TableName()] = ledgTrxs

	sumCust := &ficomodel.CustomerTransaction{}
	*sumCust = *custTrxs[0]
	sumCust.Amount = totalAmount
	sumCust.SourceLineNo = 0
	sumCust.Text = p.header.Text
	trxs[sumCust.TableName()] = []orm.DataModel{sumCust}

	// create bank payment, if header cash payment is clicked
	if p.header.CashPayment {
		cb, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBank), p.header.CashBankID)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid: cash bank: %s", p.header.CashBankID)
		}
		if cb.MainBalanceAccount == "" {
			gcb, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBank), cb.CashBankGroupID)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: cash bank group: %s", cb.CashBankGroupID)
			}
			cb.MainBalanceAccount = gcb.MainBalanceAccount
		}
		cbLedgerAccount, err := p.coas.Get(cb.MainBalanceAccount)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("missing: cash ledger account: %s", cb.MainBalanceAccount)
		}

		cashTrx := &ficomodel.CashTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCustomer,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			CashBank:          *cb,
			Status:            ficomodel.AmountConfirmed,
			Amount:            totalAmount,
		}
		trxs[cashTrx.TableName()] = []orm.DataModel{cashTrx}

		cashLedgerTrx := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCustomer,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			Account:           *cbLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            totalAmount,
		}

		custLedgerTrx := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCustomer,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			Account:           *custLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -totalAmount,
		}
		clName := cashLedgerTrx.TableName()
		trxs[clName] = append(trxs[clName], cashLedgerTrx, custLedgerTrx)
	} else {
		schedule := &ficomodel.CashSchedule{
			CompanyID:       p.header.CompanyID,
			SourceType:      ficomodel.SubledgerCustomer,
			SourceJournalID: p.header.ID,
			SourceLineNo:    0,
			Text:            p.header.Text,
			Expected:        p.header.TrxDate,
			Status:          ficomodel.CashScheduled,
			Account:         ficomodel.NewSubAccount(ficomodel.SubledgerCustomer, p.header.CustomerID),
			Amount:          totalAmount,
		}
		trxs[schedule.TableName()] = []orm.DataModel{schedule}
	}

	journalType, err := datahub.GetByID(p.opt.Db, new(ficomodel.CustomerJournalType), p.header.JournalTypeID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("invalid: customer journal type: %s", p.header.JournalTypeID)
	}

	if journalType.TransactionType == "General Invoice" {
		param := CustomerPreviewParam{
			Hub:     p.opt.Db,
			Preview: &preview,
			Header:  header,
			Lines:   lines,
		}

		// generate general invoice preview
		err = p.GenerateGeneralInvoice(&param)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("error when generate general preview: %s", err.Error())
		}
	}

	return &preview, trxs, totalAmount, nil
}

func (p *customerPosting) Post(opt PostingHubExecOpt, header *ficomodel.CustomerJournal, lines []ficomodel.JournalLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	db, _ = p.opt.Db.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	res, err = PostModelSave(p.opt.Db, p.header, "CustomerVoucherNo", trxs)
	return res, err
}

type CustomerPreviewParam struct {
	Hub     *datahub.Hub
	Preview *tenantcoremodel.PreviewReport
	Header  *ficomodel.CustomerJournal
	Lines   []ficomodel.JournalLine
}

func (p *customerPosting) GenerateGeneralInvoice(param *CustomerPreviewParam) error {
	// set header
	param.Preview.Header["No"] = param.Header.ID
	param.Preview.Header["PO"] = param.Header.References
	param.Preview.Header["TrxDate"] = param.Header.TrxDate
	param.Preview.Header["AddressAndTax"] = param.Header.AddressAndTax
	param.Preview.Header["Text"] = param.Header.Text

	// get customer name
	cust := new(tenantcoremodel.Customer)
	if err := param.Hub.GetByID(cust, param.Header.CustomerID); err != nil {
		return fmt.Errorf("error when get customer : %s", err.Error())
	}
	param.Preview.Header["Name"] = cust.Name

	// get tax setup
	taxSetups := []ficomodel.TaxSetup{}
	err := param.Hub.Gets(new(ficomodel.TaxSetup), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("IsActive", true),
	), &taxSetups)
	if err != nil {
		return fmt.Errorf("error when get tax setup: %s", err.Error())
	}

	mapTaxSetup := lo.Associate(taxSetups, func(tax ficomodel.TaxSetup) (string, ficomodel.TaxSetup) {
		return tax.ID, tax
	})

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := param.Hub.GetByID(curr, param.Header.CurrencyID); err != nil {
		return fmt.Errorf(fmt.Sprintf("currency not found: %s", param.Header.CurrencyID))
	}

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       make([][]string, len(param.Lines)+1),
	}

	mapTaxLine := map[string]float64{}
	mapAmountTaxHeader := map[string]float64{}
	section.Items[0] = []string{"Item", "Name", "Trx Date", "Description", "Quantity", "UoM", "Unit Price", "Total"}
	for i, l := range param.Lines {
		quantity := codekit.ToString(int(l.Qty))
		unitPrice := fmt.Sprintf("%s%s", curr.Symbol2, codekit.ToString(int(l.PriceEach)))
		amount := fmt.Sprintf("%s%s", curr.Symbol2, codekit.ToString(int(l.Amount)))

		// calculate tax only if taxable is true
		if l.Taxable {
			for _, code := range l.TaxCodes {
				if v, ok := mapTaxSetup[code]; ok {
					if v.CalcMethod == ficomodel.TaxCalcLine {
						mapTaxLine[v.ID] += l.Amount * v.Rate
					} else {
						mapAmountTaxHeader[v.ID] += l.Amount
					}
				}
			}
		}

		section.Items[i+1] = []string{l.TagObjectID1.AccountID, cust.Name, param.Header.TrxDate.Format("2006-02-01"), l.Text, quantity, l.UnitID, unitPrice, amount}
	}

	grandTotal := param.Header.TotalAmount
	taxes := make([]codekit.M, len(mapAmountTaxHeader)+len(mapTaxLine))
	i := 0
	for k, v := range mapTaxLine {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = fmt.Sprintf("%s%d", curr.Symbol2, int(v))
				grandTotal += v
			} else {
				tax["Value"] = fmt.Sprintf("-%s%d", curr.Symbol2, int(v))
				grandTotal -= v
			}
		}

		taxes[i] = tax
		i++
	}

	for k, v := range mapAmountTaxHeader {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			val := v * taxSetup.Rate
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = fmt.Sprintf("%s%d", curr.Symbol2, int(val))
				grandTotal += val
			} else {
				tax["Value"] = fmt.Sprintf("-%s%d", curr.Symbol2, int(val))
				grandTotal -= val
			}
		}

		taxes[i] = tax
		i++
	}

	param.Preview.Header["TotalInvoice"] = fmt.Sprintf("%s%d", curr.Symbol2, int(param.Header.TotalAmount))
	param.Preview.Header["Taxes"] = taxes
	param.Preview.Header["GrandTotal"] = fmt.Sprintf("%s%d", curr.Symbol2, int(grandTotal))
	param.Preview.Sections = append(param.Preview.Sections, section)

	return nil
}

func (p *customerPosting) Approved() error {
	return nil
}

func (p *customerPosting) Rejected() error {
	return nil
}

func (p *customerPosting) GetAccount() string {
	return p.customer.ID
}

func (p *customerPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", p.opt.UserID),
	))
	if err != nil && err != io.EOF {
		return fmt.Errorf("error when get email employee : %s", err.Error())
	}

	for _, app := range pa.Approvals {
		if app.Line == pa.CurrentStage {
			notification := ficomodel.Notification{
				UserSubmitter:            p.opt.UserID,
				UserSubmitterEmail:       employee.Email,
				JournalID:                p.header.ID,
				JournalType:              p.header.JournalTypeID,
				PostingProfileApprovalID: pa.ID,
				TrxDate:                  p.header.TrxDate,
				TrxType:                  p.header.TransactionType,
				Text:                     p.header.Text,
				Menu:                     "Customer Journal",
				Amount:                   p.header.TotalAmount,
				UserTo:                   app.UserID,
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
			}

			employee := new(tenantcoremodel.Employee)
			err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
				dbflex.Eq("_id", app.UserID),
			))
			if err != nil && err != io.EOF {
				return fmt.Errorf("error when get employee : %s", err.Error())
			}

			notification.UserToEmail = employee.Email

			err = p.opt.Db.Save(&notification)
			if err != nil {
				return fmt.Errorf("error when save notification : %s", err.Error())
			}
		}
	}

	return nil
}

func (p *customerPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
	// get latest notification
	latestNotif := new(ficomodel.Notification)
	err := p.opt.Db.GetByParm(latestNotif, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("JournalID", p.header.ID),
			dbflex.Eq("UserTo", p.opt.UserID),
		),
	).SetSort("-Created"))
	if err != nil {
		return fmt.Errorf("error when get notification : %s", err.Error())
	}

	if op == PostOpApprove {
		latestNotif.Status = string(ficomodel.JournalStatusApproved)
	} else {
		latestNotif.Status = string(ficomodel.JournalStatusRejected)
	}

	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification : %s", err.Error())
	}

	// create notification for submitter user
	latestNotif.ID = primitive.NewObjectID().Hex()
	latestNotif.UserTo = latestNotif.UserSubmitter
	latestNotif.UserToEmail = latestNotif.UserSubmitterEmail
	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification for submitter user : %s", err.Error())
	}

	// get user approval stage
	userApprovals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.UserID == p.opt.UserID
	})

	approvals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.Line == userApprovals[0].Line && a.Status != "PENDING"
	})

	// check if need to send notification
	if len(approvals) == pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line != pa.CurrentStage {
		for _, app := range pa.Approvals {
			if app.Line == pa.CurrentStage {
				latestNotif.ID = primitive.NewObjectID().Hex()
				latestNotif.UserTo = app.UserID
				latestNotif.Status = app.Status

				employee := new(tenantcoremodel.Employee)
				err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
					dbflex.Eq("_id", app.UserID),
				))
				if err != nil && err != io.EOF {
					return fmt.Errorf("error when get employee : %s", err.Error())
				}

				latestNotif.UserToEmail = employee.Email

				err = p.opt.Db.Save(latestNotif)
				if err != nil {
					return fmt.Errorf("error when save notification for next approval : %s", err.Error())
				}
			}
		}
	}

	return nil
}
