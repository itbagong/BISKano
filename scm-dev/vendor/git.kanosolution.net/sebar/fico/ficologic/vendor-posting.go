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
	"github.com/leekchan/accounting"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type vendorPosting struct {
	opt     *PostingHubCreateOpt
	header  *ficomodel.VendorJournal
	jt      *ficomodel.VendorJournalType
	trxType string

	vendor   *tenantcoremodel.Vendor
	coas     *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	expenses *sebar.MapRecord[*tenantcoremodel.ExpenseType] //TODO: samakan dengan model expense

	trxs map[string][]orm.DataModel
}

func NewVendorPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.VendorJournal, ficomodel.JournalLine] {
	c := new(vendorPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	c.trxs = make(map[string][]orm.DataModel)
	pvd := PostingProvider[*ficomodel.VendorJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *vendorPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.VendorJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lines
}

func (p *vendorPosting) Header() (*ficomodel.VendorJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(ficomodel.VendorJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}
	p.header = j

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.VendorJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}
	p.trxType = jt.TransactionType
	if p.trxType == "" {
		p.trxType = "VendorJournal"
	}

	j.PostingProfileID = jt.PostingProfileID
	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.vendor, err = datahub.GetByID(p.opt.Db, new(tenantcoremodel.Vendor), j.VendorID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: vendor: %s", j.VendorID)
	}
	vendorGroup, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.VendorGroup), p.vendor.GroupID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: vendor group id: %s", p.vendor.GroupID)
	}
	p.vendor.Dimension = tenantcorelogic.TernaryDimension(p.vendor.Dimension, vendorGroup.Dimension)
	p.vendor.DepositAccount = tenantcorelogic.TernaryString(p.vendor.DepositAccount, vendorGroup.DepositAccount)
	p.vendor.MainBalanceAccount = tenantcorelogic.TernaryString(p.vendor.MainBalanceAccount, vendorGroup.MainBalanceAccount)
	p.jt, _ = datahub.GetByID(p.opt.Db, new(ficomodel.VendorJournalType), p.header.JournalTypeID)

	return j, pp, nil
}

func (p *vendorPosting) Lines() ([]ficomodel.JournalLine, error) {
	if len(p.header.Lines) == 0 {
		return nil, fmt.Errorf("no lines")
	}

	lines := make([]ficomodel.JournalLine, len(p.header.Lines))
	for idx, line := range p.header.Lines {
		if line.LineNo == 0 {
			line.LineNo = idx + 1
		}
		line.OffsetAccount = ficomodel.NewSubAccount(ficomodel.SubledgerVendor, p.header.VendorID)
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)
		line.TrxType = p.trxType
		line.CurrencyID = p.header.CurrencyID

		/*
			TODO: verify business requirements here
			Currently if taxable and no taxcodes, make taxcodes = header.taxcodes
		*/
		if line.Taxable {
			if len(line.TaxCodes) == 0 {
				line.TaxCodes = p.header.TaxCodes
			}
		}

		lines[idx] = line
	}

	p.header.Lines = lines
	return lines, nil
}

func (p *vendorPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.VendorJournal, lines []ficomodel.JournalLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {

	preview := tenantcoremodel.PreviewReport{
		Header: make(codekit.M),
	}
	totalAmount := float64(0)

	vendLedgerAccount, err := p.coas.Get(p.vendor.MainBalanceAccount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("missing: vendor ledger account: %s", p.vendor.MainBalanceAccount)
	}

	// create vendor transaction
	vendTrxs := []*ficomodel.VendorTransaction{}
	ledgTrxs := []*ficomodel.LedgerTransaction{}
	scheduleTrxs := []orm.DataModel{}
	for _, line := range lines {
		var (
			coaID         string
			expense       *tenantcoremodel.ExpenseType
			ledgerAccount *tenantcoremodel.LedgerAccount
			err           error
		)

		if !(line.Account.AccountType == ficomodel.SubledgerAccounting || line.Account.AccountType == ficomodel.SubledgerExpense) {
			return nil, nil, 0, fmt.Errorf("line %d: allowed account type only Ledger and Expense", line.LineNo)
		}

		if line.Account.AccountType == ficomodel.SubledgerExpense {
			expense, err = p.expenses.Get(line.Account.AccountID)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: expense: %s: %s", line.Account.AccountID, err.Error())
			}
			coaID = expense.LedgerAccountID
		} else {
			coaID = line.Account.AccountID
		}
		ledgerAccount, err = p.coas.Get(coaID)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("missing: ledger account: %s, line %d", coaID, line.LineNo)
		}

		vt := &ficomodel.VendorTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerVendor.String(),
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Vendor:            *p.vendor,
			Status:            ficomodel.AmountConfirmed,
			Amount:            line.Amount,
		}

		lt := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerVendor,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            line.Amount,
		}

		ltVendor := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerVendor,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Account:           *vendLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -line.Amount,
		}

		vendTrxs = append(vendTrxs, vt)
		ledgTrxs = append(ledgTrxs, lt, ltVendor)

		totalAmount += line.Amount
	}

	sumVT := &ficomodel.VendorTransaction{}
	*sumVT = *vendTrxs[0]
	sumVT.SourceLineNo = 0
	sumVT.Amount = lo.SumBy(vendTrxs, func(v *ficomodel.VendorTransaction) float64 {
		return v.Amount
	})
	if sumTax, err := p.CalcTax(); err != nil {
		return nil, nil, 0, err
	} else {
		sumVT.Amount += sumTax
	}
	p.trxs[sumVT.TableName()] = []orm.DataModel{sumVT}
	p.trxs[ledgTrxs[0].TableName()] = append(p.trxs[ledgTrxs[0].TableName()], ToDataModels(ledgTrxs)...)

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
			SourceType:        ficomodel.SubledgerVendor,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			CashBank:          *cb,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -totalAmount,
		}
		p.trxs[cashTrx.TableName()] = []orm.DataModel{cashTrx}

		cashLedgerTrx := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerVendor,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			Account:           *cbLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -totalAmount,
		}

		vendLedgerTrx := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerVendor,
			SourceJournalType: p.header.JournalTypeID,
			SourceJournalID:   p.header.ID,
			Dimension:         p.header.Dimension,
			TrxDate:           p.header.TrxDate,
			Text:              p.header.Text,
			Account:           *vendLedgerAccount,
			Status:            ficomodel.AmountConfirmed,
			Amount:            totalAmount,
		}
		clName := cashLedgerTrx.TableName()
		p.trxs[clName] = append(p.trxs[clName], cashLedgerTrx, vendLedgerTrx)
	} else {
		schedule := &ficomodel.CashSchedule{
			CompanyID:       p.header.CompanyID,
			SourceType:      ficomodel.SubledgerVendor,
			SourceJournalID: p.header.ID,
			SourceLineNo:    0,
			Text:            p.header.Text,
			Account:         ficomodel.NewSubAccount(ficomodel.SubledgerVendor, p.header.VendorID),
			Expected:        p.header.TrxDate,
			Status:          ficomodel.CashScheduled,
			Amount:          -totalAmount,
		}

		scheduleTax := &ficomodel.CashSchedule{
			CompanyID:       p.header.CompanyID,
			SourceType:      ficomodel.SubledgerVendor,
			SourceJournalID: p.header.ID,
			SourceLineNo:    0,
			Text:            "TAX/" + p.header.Text,
			Account:         ficomodel.NewSubAccount(ficomodel.SubledgerVendor, p.header.VendorID),
			Expected:        p.header.TrxDate,
			Status:          ficomodel.CashScheduled,
			Amount:          -p.header.TaxAmount,
		}

		scheduleTrxs = append(scheduleTrxs, schedule)
		scheduleTrxs = append(scheduleTrxs, scheduleTax)
		p.trxs[schedule.TableName()] = scheduleTrxs
	}

	// preview
	param := VendorPreviewParam{
		Hub:     p.opt.Db,
		Preview: &preview,
		Header:  p.header,
		Lines:   lines,
	}

	// generate general invoice preview
	err = p.GetPreview(&param)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error when get preview: %s", err.Error())
	}

	return &preview, p.trxs, totalAmount, nil
}

func (p *vendorPosting) CalcTax() (float64, error) {
	//-- get tax codes from lines
	taxSetups := map[string]*ficomodel.TaxSetup{}
	for _, line := range p.header.Lines {
		for _, code := range line.TaxCodes {
			_, ok := taxSetups[code]
			if ok {
				continue
			}

			setup, err := datahub.GetByID(p.opt.Db, new(ficomodel.TaxSetup), code)
			if err != nil {
				continue
			}

			taxSetups[code] = setup
		}
	}

	//-- calc tax for each of setup
	taxTrxs := []*ficomodel.TaxTransaction{}
	ledgerTrxs := []*ficomodel.LedgerTransaction{}

	for _, setup := range taxSetups {
		_, taxValue, err := CalcTax(*setup, &p.header.Lines, true)
		if err != nil {
			return 0, fmt.Errorf("calc tax: %s, %s: %s", setup.ID, setup.Name, err.Error())
		}

		trx := new(ficomodel.TaxTransaction)
		trx.CompanyID = p.header.CompanyID
		trx.TaxCode = setup.ID
		trx.InvoiceOperation = setup.InvoiceOperation
		trx.SourceType = ficomodel.SubledgerVendor.String()
		trx.SourceJournalID = p.header.ID
		trx.SourceAccount = ficomodel.SubledgerAccount{
			AccountType: ficomodel.SubledgerVendor,
			AccountID:   p.header.VendorID,
		}
		trx.SourceTrxType = p.header.TransactionType
		trx.FPDate = &p.header.TrxDate
		trx.Amount = taxValue

		// tax ledger entry
		ledgerAccount, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.LedgerAccount), setup.LedgerAccountID)
		if err != nil {
			return 0, fmt.Errorf("tax setup ledger acccount: %s, %s: %s", setup.ID, setup.LedgerAccountID, err.Error())
		}
		ltx := &ficomodel.LedgerTransaction{
			CompanyID:       p.header.CompanyID,
			TrxDate:         p.header.TrxDate,
			SourceType:      ficomodel.SubledgerVendor,
			SourceJournalID: p.header.ID,
			Account:         *ledgerAccount,
			Amount:          trx.Amount,
			Text:            fmt.Sprintf("[Tax %s] %s", setup.Name, p.header.Text),
		}
		ledgerTrxs = append(ledgerTrxs, ltx)

		vendorLedgerAccount, _ := datahub.GetByID(p.opt.Db, new(tenantcoremodel.LedgerAccount), p.vendor.MainBalanceAccount)
		ltx = &ficomodel.LedgerTransaction{
			CompanyID:       p.header.CompanyID,
			TrxDate:         p.header.TrxDate,
			SourceType:      ficomodel.SubledgerVendor,
			SourceJournalID: p.header.ID,
			Account:         *vendorLedgerAccount,
			Amount:          -trx.Amount,
			Text:            fmt.Sprintf("[Tax %s] %s", setup.Name, p.header.Text),
		}
		ledgerTrxs = append(ledgerTrxs, ltx)
		taxTrxs = append(taxTrxs, trx)
	}

	ltxName := new(ficomodel.LedgerTransaction).TableName()
	p.trxs[new(ficomodel.TaxTransaction).TableName()] = ToDataModels(taxTrxs)
	p.trxs[ltxName] = append(p.trxs[ltxName], ToDataModels(ledgerTrxs)...)

	taxAmt := lo.SumBy(taxTrxs, func(t *ficomodel.TaxTransaction) float64 {
		return t.Amount
	})
	return -taxAmt, nil
}

type VendorPreviewParam struct {
	Hub     *datahub.Hub
	Preview *tenantcoremodel.PreviewReport
	Header  *ficomodel.VendorJournal
	Lines   []ficomodel.JournalLine
}

func (p *vendorPosting) GetPreview(param *VendorPreviewParam) error {

	// get vendor name
	vnd := new(tenantcoremodel.Vendor)
	if err := param.Hub.GetByID(vnd, param.Header.VendorID); err != nil {
		return fmt.Errorf("invalid: vendor: %s", err.Error())
	}

	// set header
	var datas [][]string
	// datas = append(datas, []string{"", "","","","",param.Header.AddressAndTax})
	datas = append(datas, []string{fmt.Sprintf("No : %s", param.Header.ID), "", "", "", "", param.Header.TrxDate.Format("02 January 2006")})
	datas = append(datas, []string{fmt.Sprintf("PO : %s", param.Header.References.Get("PO Ref No", "").(string)), "", "", "", "", vnd.Name})
	datas = append(datas, []string{fmt.Sprintf("Text : %s", param.Header.Text), "", "", "", "", ""})

	var dataMobile [][]string
	dataMobile = append(dataMobile, []string{"No : ", param.Header.ID})
	dataMobile = append(dataMobile, []string{"PO : ", param.Header.References.Get("PO Ref No", "").(string)})
	dataMobile = append(dataMobile, []string{"Trx Date : ", param.Header.TrxDate.Format("02 January 2006")})
	dataMobile = append(dataMobile, []string{"Name : ", vnd.Name})
	dataMobile = append(dataMobile, []string{"Text : ", param.Header.Text})

	param.Preview.Header["Data"] = datas
	param.Preview.HeaderMobile.Data = dataMobile
	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := param.Hub.GetByID(curr, param.Header.CurrencyID); err != nil {
		return fmt.Errorf(fmt.Sprintf("currency not found: %s", param.Header.CurrencyID))
	}

	sectionLine := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Purchase Line",
		Items:       make([][]string, len(param.Lines)+1),
	}
	sectionLine.Items[0] = []string{"Trx Date", "Description", "Quantity:R", "Unit", "Unit Price:R", "Discount:R", "PPN:R", "PPH:R", "Total:R"}
	for i, l := range param.Lines {
		quantity := codekit.ToString(int(l.Qty))
		unitPrice := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.PriceEach, 2, ",", "."))
		amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.Amount, 2, ",", "."))
		discount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.Discount, 2, ",", "."))
		if l.DiscountType == "percent" {
			discount = fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber((l.Discount*l.PriceEach)/100, 2, ",", "."))
		}
		ppn := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.PPN, 2, ",", "."))
		pph := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.PPH, 2, ",", "."))

		sectionLine.Items[i+1] = []string{
			param.Header.TrxDate.Format("2006-02-01"),
			l.Text, quantity, l.UnitID, unitPrice, discount, ppn, pph, amount}
	}

	taxes := FromDataModels(p.trxs[new(ficomodel.TaxTransaction).TableName()], new(ficomodel.TaxTransaction))
	sectionTax := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Purchase Tax",
		Items:       make([][]string, len(taxes)+1),
	}
	sectionTax.Items[0] = []string{"Tax Code", "Amount:R"}
	for idx, tax := range taxes {
		invoiceOperation := 1
		if tax.InvoiceOperation == ficomodel.TaxDecreaseAmount {
			invoiceOperation = -1
		}
		sectionTax.Items[idx+1] = []string{
			tax.TaxCode,
			fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(tax.Amount*float64(invoiceOperation), 2, ",", ".")),
		}
	}

	ledgers := FromDataModels(p.trxs[new(ficomodel.LedgerTransaction).TableName()], new(ficomodel.LedgerTransaction))
	sectionLedger := tenantcoremodel.PreviewSection{
		SectionType:         tenantcoremodel.PreviewAsGrid,
		Title:               "Ledger Entry",
		Items:               make([][]string, len(ledgers)+1),
		RestrictedByFeature: true,
		FeatureIDs:          []string{"LedgerEntry"},
	}
	sectionLedger.Items[0] = []string{"Account", "Amount:R"}
	for idx, ledger := range ledgers {
		sectionLedger.Items[idx+1] = []string{
			fmt.Sprintf("%s - %s", ledger.Account.ID, ledger.Account.Name),
			fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(ledger.Amount, 2, ",", ".")),
		}
	}

	subTotal := fmt.Sprintf("%s%s%s", curr.Symbol2, accounting.FormatNumber(param.Header.SubtotalAmount, 2, ",", "."), ":R")
	lineDiscount := fmt.Sprintf("%s%s%s", curr.Symbol2, accounting.FormatNumber(param.Header.DiscountAmount, 2, ",", "."), ":R")
	taxAmount := fmt.Sprintf("%s%s%s", curr.Symbol2, accounting.FormatNumber(param.Header.TaxAmount, 2, ",", "."), ":R")
	headerDiscount := fmt.Sprintf("%s%s%s", curr.Symbol2, accounting.FormatNumber(param.Header.HeaderDiscountAmount, 2, ",", "."), ":R")
	totalAmount := fmt.Sprintf("%s%s%s", curr.Symbol2, accounting.FormatNumber(param.Header.TotalAmount, 2, ",", "."), ":R")

	var footer [][]string
	footer = append(footer, []string{"", "", "", "", "", "", "Subtotal Amount : ", subTotal})
	footer = append(footer, []string{"", "", "", "", "", "", "Line Discount Amount : ", lineDiscount})
	footer = append(footer, []string{"", "", "", "", "", "", "Tax Amount : ", taxAmount})
	footer = append(footer, []string{"", "", "", "", "", "", "Header Discount : ", headerDiscount})
	footer = append(footer, []string{"", "", "", "", "", "", "Total Amount : ", totalAmount})

	var footerMobile [][]string
	footerMobile = append(footerMobile, []string{"Subtotal Amount : ", subTotal})
	footerMobile = append(footerMobile, []string{"Line Discount Amount : ", lineDiscount})
	footerMobile = append(footerMobile, []string{"Tax Amount : ", taxAmount})
	footerMobile = append(footerMobile, []string{"Header Discount : ", headerDiscount})
	footerMobile = append(footerMobile, []string{"Total Amount : ", totalAmount})

	param.Preview.Header["Footer"] = footer
	param.Preview.HeaderMobile.Footer = footerMobile

	param.Preview.Sections = append(param.Preview.Sections, sectionLine, sectionTax, sectionLedger)

	return nil
}

func (p *vendorPosting) Post(opt PostingHubExecOpt, header *ficomodel.VendorJournal, lines []ficomodel.JournalLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		err error
		res string
	)

	err = sebar.Tx(opt.Db, false, func(tx *datahub.Hub) error {
		res, err = PostModelSave(p.opt.Db, p.header, "VendorVoucherNo", trxs)
		return err
	})
	return res, err
}

func (p *vendorPosting) Approved() error {
	return nil
}

func (p *vendorPosting) Rejected() error {
	return nil
}

func (p *vendorPosting) GetAccount() string {
	return p.vendor.ID
}

func (p *vendorPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Vendor Journal",
				Amount:                   p.header.TotalAmount,
				UserTo:                   app.UserID,
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
				IsApproval:               true,
			}

			switch p.header.TransactionType {
			case "Employee Expense":
				notification.Menu = "Employee Expense Submission"
			case "General Submission":
				notification.Menu = "General Submission"
			case "Site Entry Expense":
				notification.Menu = "Site Entry"
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

func (p *vendorPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
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
	latestNotif.IsApproval = false

	if op == PostOpApprove {
		latestNotif.Message = fmt.Sprintf("Your Submission ID. %s has been approved", p.header.ID)
	} else {
		latestNotif.Message = fmt.Sprintf("Your Submission ID. %s has been rejected", p.header.ID)
	}

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
	if len(approvals) >= pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line != pa.CurrentStage {
		for _, app := range pa.Approvals {
			if app.Line == pa.CurrentStage {
				latestNotif.ID = primitive.NewObjectID().Hex()
				latestNotif.UserTo = app.UserID
				latestNotif.Status = app.Status
				latestNotif.IsApproval = true

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
