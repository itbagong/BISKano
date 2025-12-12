package ficologic

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ekokurniadi/terbilang"
	"github.com/leekchan/accounting"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type cashPosting struct {
	header *ficomodel.CashJournal
	pp     *ficomodel.PostingProfile
	jt     *ficomodel.CashJournalType

	opt *PostingHubCreateOpt
	cb  *tenantcoremodel.CashBank

	coas           *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	expenses       *sebar.MapRecord[*tenantcoremodel.ExpenseType]
	banks          *sebar.MapRecord[*tenantcoremodel.CashBank]
	bankGroups     *sebar.MapRecord[*tenantcoremodel.CashBankGroup]
	vendors        *sebar.MapRecord[*tenantcoremodel.Vendor]
	customers      *sebar.MapRecord[*tenantcoremodel.Customer]
	customerGroups *sebar.MapRecord[*tenantcoremodel.CustomerGroup]
	vendorGroups   *sebar.MapRecord[*tenantcoremodel.VendorGroup]

	trxs map[string][]orm.DataModel

	calcHeader bool
}

type CashPreviewParam struct {
	Hub     *datahub.Hub
	Preview *tenantcoremodel.PreviewReport
	Header  *ficomodel.CashJournal
	Lines   []ficomodel.JournalLine
}

func NewCashPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.CashJournal, ficomodel.JournalLine] {
	c := new(cashPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	c.banks = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.CashBank))
	c.vendors = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.Vendor))
	c.customers = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.Customer))
	c.customerGroups = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.CustomerGroup))
	c.vendorGroups = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.VendorGroup))
	c.trxs = make(map[string][]orm.DataModel)
	pvd := PostingProvider[*ficomodel.CashJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *cashPosting) Header() (*ficomodel.CashJournal, *ficomodel.PostingProfile, error) {
	var (
		err error
	)

	if p.calcHeader {
		return p.header, p.pp, nil
	}

	p.header, err = datahub.GetByID(p.opt.Db, new(ficomodel.CashJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract header: %s: %s", p.opt.JournalID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.CashJournalType), p.header.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract journal type: %s: %s", p.header.JournalTypeID, err.Error())
	}
	p.jt = jt

	for idx, _ := range p.header.Lines {
		p.header.Lines[idx].TrxType = jt.TransactionType
	}
	p.header.PostingProfileID = jt.PostingProfileID

	p.pp, err = datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), p.header.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract posting profile: %s: %s", p.header.PostingProfileID, err.Error())
	}

	p.cb, err = datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBank), p.header.CashBookID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract cash bank: %s: %s", p.header.CashBookID, err.Error())
	}

	cbg, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBankGroup), p.cb.CashBankGroupID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract cash bank group: %s: %s", p.cb.CashBankGroupID, err.Error())
	}

	p.header.Dimension = tenantcorelogic.TernaryDimension(p.header.Dimension, jt.Dimension, p.cb.Dimension, cbg.Dimension)

	p.cb.MainBalanceAccount = tenantcorelogic.TernaryString(p.cb.MainBalanceAccount, cbg.MainBalanceAccount)
	_, err = p.coas.Get(p.cb.MainBalanceAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("cash main balance account: %s: %s", p.cb.MainBalanceAccount, err.Error())
	}

	p.calcHeader = true
	return p.header, p.pp, err
}

func (p *cashPosting) Lines() ([]ficomodel.JournalLine, error) {
	if len(p.header.Lines) == 0 {
		return nil, fmt.Errorf("no lines")
	}

	lines := []ficomodel.JournalLine{}
	for idx, line := range p.header.Lines {
		if line.Ignore {
			continue
		}
		if line.LineNo == 0 {
			line.LineNo = idx + 1
		}
		if line.OffsetAccount.AccountID == "" {
			line.OffsetAccount = ficomodel.NewSubAccount(ficomodel.SubledgerCashBank, p.header.CashBookID)
		}
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)

		if p.header.CashJournalType == strings.ToUpper(string(ficomodel.CashOut)) ||
			p.header.CashJournalType == strings.ToUpper("Submission Cash Out") {
			line.Amount = -line.Amount
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func (p *cashPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line ficomodel.JournalLine, index int) ficomodel.JournalLine {
		return line
	})
}

func (p *cashPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	offsetCoa, err := p.coas.Get(p.cb.MainBalanceAccount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("ledger account: %s: %s", p.cb.MainBalanceAccount, err.Error())
	}

	allowedModule := []tenantcoremodel.TrxModule{
		ficomodel.SubledgerCashBank,
		ficomodel.SubledgerCustomer, ficomodel.SubledgerVendor,
		ficomodel.SubledgerAccounting, ficomodel.SubledgerExpense}

	for _, line := range lines {
		needCashSchedule := !(line.Account.AccountType == ficomodel.SubledgerCashBank && line.Account.AccountID != "")
		if line.Amount == 0 {
			continue
		}
		if line.Ignore {
			continue
		}
		if !codekit.HasMember(allowedModule, line.Account.AccountType) {
			return nil, nil, 0, fmt.Errorf("line %d: allowed account type: %v", line.LineNo, allowedModule)
		}
		ct := &ficomodel.CashTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			CashBank:          *p.cb,
			TrxDate:           p.header.TrxDate,
			TrxType:           line.TrxType,
			Amount:            line.Amount,
			Status:            ficomodel.AmountConfirmed,
			Text:              line.Text,
		}

		var cs *ficomodel.CashSchedule
		if needCashSchedule {
			cs = &ficomodel.CashSchedule{
				CompanyID:       p.header.CompanyID,
				SourceType:      ficomodel.SubledgerCashBank,
				SourceJournalID: p.header.ID,
				SourceLineNo:    line.LineNo,
				Expected:        p.header.TrxDate,
				Text:            line.Text,
				Account:         ficomodel.NewSubAccount(line.Account.AccountType, line.Account.AccountID),
				Amount:          line.Amount,
				Status:          ficomodel.CashScheduled,
			}
		}

		var (
			expense   *tenantcoremodel.ExpenseType
			mainCoa   *tenantcoremodel.LedgerAccount
			customer  *tenantcoremodel.Customer
			vendor    *tenantcoremodel.Vendor
			bankLawan *tenantcoremodel.CashBank
			cashSched *ficomodel.CashSchedule
			err       error
		)

		switch line.Account.AccountType {
		case ficomodel.SubledgerExpense:
			expense, mainCoa, err = GetSubledgerFromMapRecord(line.Account.AccountID, "LedgerAccountID", p.expenses, p.coas)

			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          -line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
			}

		case ficomodel.SubledgerAccounting:
			mainCoa, err = p.coas.Get(line.Account.AccountID)

		case ficomodel.SubledgerCustomer:
			customer, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID,
				"Setting.MainBalanceAccount", "GroupID",
				p.customers, p.customerGroups, p.coas)
			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          -line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}

				custTrx := ficomodel.CustomerTransaction{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.AmountConfirmed,
					Customer:        *customer,
					Amount:          -line.Amount,
					TrxDate:         p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting.String(),
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
				trxs[custTrx.TableName()] = append(trxs[custTrx.TableName()], &custTrx)
			}

		case ficomodel.SubledgerVendor:
			vendor, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID, "MainBalanceAccount", "GroupID", p.vendors, p.vendorGroups, p.coas)
			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}

				vendTrx := ficomodel.VendorTransaction{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.AmountConfirmed,
					Vendor:          *vendor,
					Amount:          line.Amount,
					TrxDate:         p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting.String(),
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
				trxs[vendTrx.TableName()] = append(trxs[vendTrx.TableName()], &vendTrx)
			}

		case ficomodel.SubledgerCashBank:
			bankLawan, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID,
				"MainBalanceAccount", "CashBankGroupID",
				p.banks, p.bankGroups, p.coas)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: line account bank: %s, %s", line.Account.AccountID, err.Error())
			}
			ctLawan := &ficomodel.CashTransaction{
				CompanyID:         p.header.CompanyID,
				SourceType:        ficomodel.SubledgerCashBank,
				SourceJournalID:   p.header.ID,
				SourceLineNo:      line.LineNo,
				SourceJournalType: p.header.JournalTypeID,
				Dimension:         line.Dimension,
				CashBank:          *bankLawan,
				TrxDate:           p.header.TrxDate,
				TrxType:           line.TrxType,
				Amount:            -line.Amount,
				Status:            ficomodel.AmountConfirmed,
			}
			trxs[ctLawan.TableName()] = append(trxs[ctLawan.TableName()], ctLawan)

		default:
			return nil, nil, 0, fmt.Errorf("invalid: line account type: %d, %s", line.LineNo, line.Account.AccountType)
		}

		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid: account: %s, %s: %s", line.Account.AccountType, line.Account.AccountID, err.Error())
		}

		if cashSched != nil {
			trxs[cashSched.TableName()] = append(trxs[cashSched.TableName()], cashSched)
		}

		ledgerTrxOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceTrxType:     line.TrxType,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			Account:           *offsetCoa,
			TrxDate:           p.header.TrxDate,
			Amount:            line.Amount,
			Status:            ficomodel.AmountConfirmed,
		}

		ledgerTrxMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceTrxType:     line.TrxType,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			Account:           *mainCoa,
			Expense:           expense,
			TrxDate:           p.header.TrxDate,
			Amount:            -line.Amount,
			Status:            ficomodel.AmountConfirmed,
		}

		trxs[ct.TableName()] = append(trxs[ct.TableName()], ct)
		if cs != nil {
			trxs[cs.TableName()] = append(trxs[cs.TableName()], cs)
		}
		trxs[ledgerTrxMain.TableName()] = append(trxs[ledgerTrxMain.TableName()], ledgerTrxMain, ledgerTrxOffset)
	}

	p.trxs = trxs

	// preview
	param := CashPreviewParam{
		Hub:     p.opt.Db,
		Preview: &preview,
		Header:  p.header,
		Lines:   lines,
	}

	// generate general invoice preview
	respPreview, err := p.GetPreview(&param)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error when get preview: %s", err.Error())
	}

	return respPreview.Preview, trxs, header.Amount, nil
}

func (p *cashPosting) GetPreview(param *CashPreviewParam) (*CashPreviewParam, error) {

	// get cash out journal
	journal := new(ficomodel.CashJournal)
	if err := param.Hub.GetByID(journal, param.Header.ID); err != nil {
		return nil, fmt.Errorf("error when get journal: %s", err.Error())
	}
	// generatedPreview := CashPreviewParam{}
	if journal.JournalTypeID == "CJT-004" {
		generatedPreview, err := p.generateCashOutSite(param, journal)
		if err != nil {
			return nil, fmt.Errorf("error when generate cash out site: %s", err.Error())
		}
		param.Preview = generatedPreview.Preview
	} else if journal.JournalTypeID == "CJT-003" {
		generatedPreview, err := p.generatePettyCash(param, journal)
		if err != nil {
			return nil, fmt.Errorf("error when generate petty cash: %s", err.Error())
		}
		param.Preview = generatedPreview.Preview
	} else if journal.JournalTypeID == "CJTCASHOUTVENDOR" {
		generatedPreview, err := p.generateCashOutDefaultVendor(param, journal)
		if err != nil {
			return nil, fmt.Errorf("error when generate cash vendor: %s", err.Error())
		}
		param.Preview = generatedPreview.Preview
	}

	// Save the preview report
	preview, _ := tenantcorelogic.GetPreviewBySource(param.Hub, "CASHBANK", param.Header.ID, "", "Default")
	preview.SourceJournalTypeID = journal.JournalTypeID
	param.Hub.Save(preview)

	return param, nil
}

func (p *cashPosting) generateCashOutDefaultVendor(param *CashPreviewParam, journal *ficomodel.CashJournal) (*CashPreviewParam, error) {
	accountID := ""
	vnd := new(tenantcoremodel.Vendor)

	for _, l := range journal.Lines {
		if l.Account.AccountType == ficomodel.SubledgerVendor {
			// get vendor
			if err := param.Hub.GetByID(vnd, l.Account.AccountID); err != nil {
				return nil, fmt.Errorf("error when get vendor: %s", err.Error())
			}
			accountID = vnd.MainBalanceAccount
		} else if l.Account.AccountType == ficomodel.SubledgerExpense {
			// get expense
			exp := new(tenantcoremodel.ExpenseType)
			if err := param.Hub.GetByID(exp, l.Account.AccountID); err != nil {
				return nil, fmt.Errorf("error when get vendor: %s", err.Error())
			}
			accountID = exp.LedgerAccountID
		} else {
			accountID = l.Account.AccountID
		}
	}

	// get main balance account
	ledgerAccount := new(tenantcoremodel.LedgerAccount)
	if err := param.Hub.GetByID(ledgerAccount, accountID); err != nil {
		return nil, fmt.Errorf("error when get main balance account: %s", err.Error())
	}

	// get vendor
	cashBank := new(tenantcoremodel.CashBank)
	if err := param.Hub.GetByID(cashBank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := param.Hub.GetByID(curr, journal.CurrencyID); err != nil {
		return nil, fmt.Errorf("error when get currency: %s", err.Error())
	}

	// get site
	site := new(tenantcoremodel.DimensionMaster)
	if err := param.Hub.GetByID(site, journal.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	// get posting profile
	pa := new(ficomodel.PostingApproval)
	if journal.Status != ficomodel.JournalStatusDraft {
		if err := param.Hub.GetByFilter(pa, dbflex.Eq("SourceID", journal.ID)); err != nil {
			return nil, fmt.Errorf("error when get posting approval: %s", err.Error())
		}
	}

	userIDs := make([]string, 0)
	userIDs = append(userIDs, journal.CreatedBy)
	for _, appr := range pa.Approvals {
		userIDs = append(userIDs, appr.UserID)
	}

	for _, post := range pa.Postingers {
		userIDs = append(userIDs, post.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err := param.Hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", userIDs...),
	), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	preview, _ := tenantcorelogic.GetPreviewBySource(param.Hub, "CASH OUT", journal.ID, "", journal.JournalTypeID)
	preview.Name = journal.JournalTypeID
	preview.PreviewReport = &tenantcoremodel.PreviewReport{}
	preview.PreviewReport.Header = codekit.M{}
	preview.PreviewReport.Sections = make([]tenantcoremodel.PreviewSection, 0)

	headers := make([][]string, 4)
	if vnd.Name == "" {
		headers[0] = []string{"", "", "", "", "No Voucher :", journal.ID}
	} else {
		headers[0] = []string{"Nama Vendor :", vnd.Name, "", "", "No Voucher :", journal.ID}
	}

	headers[1] = []string{"Nama Bank :", cashBank.BankName, "", "", "Tgl Jatuh Tempo :", ""}

	// set due date
	if len(journal.Lines) > 0 {
		applies := []ficomodel.CashApply{}
		err := param.Hub.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("Source.JournalID", journal.ID),
				dbflex.Eq("ApplyTo.Module", ficomodel.SubledgerVendor),
			),
		).SetSelect("ApplyTo.JournalID"), &applies)
		if err != nil {
			return nil, fmt.Errorf("error when get cash apply: %s", err.Error())
		}

		if len(applies) > 0 {
			vendorJournalIDs := make([]string, 0)
			for _, a := range applies {
				vendorJournalIDs = append(vendorJournalIDs, a.ApplyTo.JournalID)
			}

			vendors := []ficomodel.VendorJournal{}
			err := param.Hub.Gets(new(ficomodel.VendorJournal), dbflex.NewQueryParam().SetWhere(
				dbflex.In("_id", vendorJournalIDs...),
			).SetSelect("ExpectedDate").SetSort("-ExpectedDate").SetTake(1), &vendors)
			if err != nil {
				return nil, fmt.Errorf("error when get vendor journal: %s", err.Error())
			}

			if len(vendors) == 1 {
				if vendors[0].ExpectedDate != nil {
					headers[1][5] = vendors[0].ExpectedDate.Format("02-01-2006")
				}
			}
		}
	}

	headers[2] = []string{"No Rekening :", cashBank.BankAccountNo, "", "", "Bank/Cash :", cashBank.Name}
	headers[3] = []string{"Nama Rekening :", cashBank.BankAccountName, "", "", "Site :", site.Label}
	preview.PreviewReport.Header["Data"] = headers

	previewDetail := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       make([][]string, 4),
	}
	previewDetail.Items[0] = []string{"No Akun", "Keterangan", "Jumlah"}
	amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(journal.Amount, 2, ",", "."))
	previewDetail.Items[1] = []string{fmt.Sprintf("%s | %s", ledgerAccount.ID, ledgerAccount.Name), journal.Text, amount}
	previewDetail.Items[2] = []string{"", "Total", amount}

	convertFirstChar := func(word string) string {
		words := strings.Split(word, " ")
		for i := range words {
			words[i] = cases.Title(language.Indonesian).String(words[i])
		}

		return strings.Join(words, " ")
	}
	format := terbilang.Init()
	previewDetail.Items[3] = []string{"Terbilang", "", convertFirstChar(format.Convert(int64(journal.Amount)))}

	// signature
	preview.PreviewReport.MultipleRowSignature = make([][]tenantcoremodel.Signature, 3)
	// Prepared by
	preview.PreviewReport.MultipleRowSignature[0] = append(preview.PreviewReport.MultipleRowSignature[0], tenantcoremodel.Signature{
		ID:        journal.CreatedBy,
		Header:    "Prepared By",
		Footer:    mapUser[journal.CreatedBy],
		Confirmed: "Tgl. " + journal.Created.Format("02-January-2006 15:04:05"),
		Status:    "",
	})

	// checked by
	for _, appr := range pa.Approvals {
		if appr.Confirmed != nil {
			preview.PreviewReport.MultipleRowSignature[1] = append(preview.PreviewReport.MultipleRowSignature[1], tenantcoremodel.Signature{
				ID:        appr.UserID,
				Header:    "Checked By",
				Footer:    mapUser[appr.UserID],
				Confirmed: "Tgl. " + appr.Confirmed.Format("02-January-2006 15:04:05"),
				Status:    "",
			})
		}
	}

	// approved by
	for _, post := range pa.Postingers {
		for _, userID := range post.UserIDs {
			preview.PreviewReport.MultipleRowSignature[2] = append(preview.PreviewReport.MultipleRowSignature[2], tenantcoremodel.Signature{
				ID:        userID,
				Header:    "Approved By",
				Footer:    mapUser[userID],
				Confirmed: "",
				Status:    "",
			})
		}
	}

	preview.PreviewReport.Sections = append(preview.PreviewReport.Sections, previewDetail)
	// param.Hub.Save(preview)
	param.Preview = preview.PreviewReport
	return param, nil
}

func (p *cashPosting) generateCashOutSite(param *CashPreviewParam, journal *ficomodel.CashJournal) (*CashPreviewParam, error) {
	accountID := ""
	vnd := new(tenantcoremodel.Vendor)

	for _, l := range journal.Lines {
		if l.Account.AccountType == ficomodel.SubledgerVendor {
			// get vendor
			if err := param.Hub.GetByID(vnd, l.Account.AccountID); err != nil {
				return nil, fmt.Errorf("error when get vendor: %s", err.Error())
			}
			accountID = vnd.MainBalanceAccount
		} else if l.Account.AccountType == ficomodel.SubledgerExpense {
			// get expense
			exp := new(tenantcoremodel.ExpenseType)
			if err := param.Hub.GetByID(exp, l.Account.AccountID); err != nil {
				return nil, fmt.Errorf("error when get vendor: %s", err.Error())
			}
			accountID = exp.LedgerAccountID
		} else {
			accountID = l.Account.AccountID
		}
	}

	// get main balance account
	ledgerAccount := new(tenantcoremodel.LedgerAccount)
	if err := param.Hub.GetByID(ledgerAccount, accountID); err != nil {
		return nil, fmt.Errorf("error when get main balance account: %s", err.Error())
	}

	// bank
	bank := new(tenantcoremodel.CashBank)
	if err := param.Hub.GetByID(bank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get vendor
	cashBank := new(tenantcoremodel.CashBank)
	if err := param.Hub.GetByID(cashBank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := param.Hub.GetByID(curr, journal.CurrencyID); err != nil {
		return nil, fmt.Errorf("error when get currency: %s", err.Error())
	}

	// get site
	site := new(tenantcoremodel.DimensionMaster)
	if err := param.Hub.GetByID(site, journal.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	// get posting profile
	pa := new(ficomodel.PostingApproval)
	if journal.Status != ficomodel.JournalStatusDraft {
		if err := param.Hub.GetByFilter(pa, dbflex.Eq("SourceID", journal.ID)); err != nil {
			return nil, fmt.Errorf("error when get posting approval: %s", err.Error())
		}
	}

	userIDs := make([]string, 0)
	userIDs = append(userIDs, journal.CreatedBy)
	for _, appr := range pa.Approvals {
		userIDs = append(userIDs, appr.UserID)
	}

	for _, post := range pa.Postingers {
		userIDs = append(userIDs, post.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err := param.Hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", userIDs...),
	), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	preview, _ := tenantcorelogic.GetPreviewBySource(param.Hub, "CASH OUT", journal.ID, "", journal.JournalTypeID)
	preview.Name = journal.JournalTypeID
	preview.PreviewReport = &tenantcoremodel.PreviewReport{}
	preview.PreviewReport.Header = codekit.M{}
	preview.PreviewReport.Sections = make([]tenantcoremodel.PreviewSection, 0)

	headers := make([][]string, 3)
	headers[0] = []string{"Dibayarkan kepada :", "", "", "", "No Voucher :", journal.ID}
	headers[1] = []string{"Site :", site.Label, "", "", "Tanggal :", journal.TrxDate.Format("2006-01-02")}
	headers[2] = []string{"No. Journal :", "", "", "", "Bank/Cash :", bank.Name}

	vendors := []ficomodel.VendorJournal{}
	// set due date
	if len(journal.Lines) > 0 {
		applies := []ficomodel.CashApply{}
		err := param.Hub.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("Source.JournalID", journal.ID),
				dbflex.Eq("ApplyTo.Module", ficomodel.SubledgerVendor),
			),
		), &applies)
		if err != nil {
			return nil, fmt.Errorf("error when get cash apply: %s", err.Error())
		}

		applyIds := lo.Map(applies, func(m ficomodel.CashApply, index int) string {
			return m.ApplyTo.RecordID
		})

		schedules := []ficomodel.CashSchedule{}
		err = param.Hub.Gets(new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.In("_id", applyIds...),
				dbflex.Eq("Status", ficomodel.CashSettled),
			),
		), &schedules)
		if err != nil {
			return nil, fmt.Errorf("error when get cash schedule: %s", err.Error())
		}

		if len(schedules) > 0 {
			vendorJournalIDs := lo.Map(schedules, func(m ficomodel.CashSchedule, index int) string {
				return m.SourceJournalID
			})

			err := param.Hub.Gets(new(ficomodel.VendorJournal), dbflex.NewQueryParam().SetWhere(
				dbflex.In("_id", vendorJournalIDs...),
			).SetSort("-Created"), &vendors)
			if err != nil {
				return nil, fmt.Errorf("error when get vendor journal: %s", err.Error())
			}
		}
	}

	// headers[2] = []string{"No Rekening :", bank.BankAccountNo, "", "", "Bank/Cash :", cashBank.Name}
	// headers[3] = []string{"Nama Rekening :", bank.BankAccountName, "", "", "Site :", site.Label}
	preview.PreviewReport.Header["Data"] = headers

	listExpense := []string{}
	for _, c := range vendors {
		expenseIds := lo.Map(c.Lines, func(m ficomodel.JournalLine, index int) string {
			return m.Account.AccountID
		})
		listExpense = append(listExpense, expenseIds...)
	}

	expenses := []tenantcoremodel.ExpenseType{}
	err = param.Hub.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", listExpense...),
		),
	), &expenses)
	if err != nil {
		return nil, fmt.Errorf("error when get cash schedule: %s", err.Error())
	}

	//get list of ledger account by list of ids
	ledgerAccountIds := lo.Map(expenses, func(m tenantcoremodel.ExpenseType, index int) string {
		return m.LedgerAccountID
	})
	ledgerAccounts := []tenantcoremodel.LedgerAccount{}
	e := param.Hub.GetsByFilter(new(tenantcoremodel.LedgerAccount), dbflex.In("_id", ledgerAccountIds...), &ledgerAccounts)
	if e != nil {
		fmt.Errorf("Failed populate data ledger accounts: %s", e.Error())
	}

	previewDetail := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       make([][]string, 0),
	}

	previewDetail.Items = append(previewDetail.Items, []string{"No.", "No Akun", "Keterangan", "Jumlah"})
	no := 1
	totalAmount := 0
	for _, c := range vendors {
		for _, d := range c.Lines {
			accountName := ""
			for _, e := range expenses {
				if e.ID == d.Account.AccountID {
					for _, f := range ledgerAccounts {
						if e.LedgerAccountID == f.ID {
							accountName = f.ID + " - " + f.Name
							break
						}
					}
				}
			}

			amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(d.Amount, 2, ",", "."))
			previewDetail.Items = append(previewDetail.Items, []string{fmt.Sprintf("%d", no), accountName, c.ID + " | " + d.Text, amount})
			totalAmount += int(d.Amount)
			no++
		}
	}
	// amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(journal.Amount, 2, ",", "."))
	// previewDetail.Items[1] = []string{fmt.Sprintf("%s | %s", ledgerAccount.ID, ledgerAccount.Status), journal.Text, amount}
	totalAmountFormatted := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(totalAmount, 2, ",", "."))
	previewDetail.Items = append(previewDetail.Items, []string{"", "Total", "", fmt.Sprintf("%s", totalAmountFormatted)})

	convertFirstChar := func(word string) string {
		words := strings.Split(word, " ")
		for i := range words {
			words[i] = cases.Title(language.Indonesian).String(words[i])
		}

		return strings.Join(words, " ")
	}
	format := terbilang.Init()
	totalFormatted := convertFirstChar(format.Convert(int64(totalAmount)))
	previewDetail.Items = append(previewDetail.Items, []string{"Terbilang", "", "", totalFormatted})

	nosig := 0

	// Ensure the outer slice is not nil or empty
	if len(preview.PreviewReport.MultipleRowSignature) <= nosig {
		// Grow the outer slice to accommodate the required index
		preview.PreviewReport.MultipleRowSignature = append(preview.PreviewReport.MultipleRowSignature, make([]tenantcoremodel.Signature, 0))
	}

	// Append "Prepared By" signature
	preview.PreviewReport.MultipleRowSignature[0] = append(preview.PreviewReport.MultipleRowSignature[0], tenantcoremodel.Signature{
		ID:        journal.CreatedBy,
		Header:    "Prepared By",
		Footer:    mapUser[journal.CreatedBy],
		Confirmed: "Tgl. " + journal.Created.Format("02-January-2006 15:04:05"),
		Status:    "",
	})

	// Increment nosig for the next set of signatures
	nosig++

	// Add "Approved By" signatures from pa.Postingers
	for _, post := range pa.Postingers {
		for _, userID := range post.UserIDs {
			// Ensure the outer slice has the required index
			if len(preview.PreviewReport.MultipleRowSignature) <= nosig {
				preview.PreviewReport.MultipleRowSignature = append(preview.PreviewReport.MultipleRowSignature, make([]tenantcoremodel.Signature, 0))
			}

			// Append "Approved By" signature
			preview.PreviewReport.MultipleRowSignature[0] = append(preview.PreviewReport.MultipleRowSignature[0], tenantcoremodel.Signature{
				ID:        userID,
				Header:    "Approved By",
				Footer:    mapUser[userID],
				Confirmed: "",
				Status:    "",
			})
		}

		// Move to the next index for the next type of signature
		nosig++
	}

	// Ensure the slice has the final index for "Received By"
	if len(preview.PreviewReport.MultipleRowSignature) <= nosig {
		preview.PreviewReport.MultipleRowSignature = append(preview.PreviewReport.MultipleRowSignature, make([]tenantcoremodel.Signature, 0))
	}

	// Append "Received By" signature
	preview.PreviewReport.MultipleRowSignature[0] = append(preview.PreviewReport.MultipleRowSignature[0], tenantcoremodel.Signature{
		ID:        "",
		Header:    "Received By",
		Footer:    "",
		Confirmed: "",
		Status:    "",
	})

	// Add the preview detail to Sections
	preview.PreviewReport.Sections = append(preview.PreviewReport.Sections, previewDetail)
	preview.SourceJournalTypeID = journal.JournalTypeID
	// Save the preview report
	// param.Hub.Save(preview)
	param.Preview = preview.PreviewReport

	return param, nil
}

func (p *cashPosting) generatePettyCash(param *CashPreviewParam, journal *ficomodel.CashJournal) (*CashPreviewParam, error) {

	// get cash bank
	cashbank := new(tenantcoremodel.CashBank)
	if err := param.Hub.GetByID(cashbank, param.Header.CashBookID); err != nil {
		return nil, fmt.Errorf("invalid: CashBank: %s", err.Error())
	}

	// get site
	site := new(tenantcoremodel.DimensionMaster)
	if err := param.Hub.GetByID(site, param.Header.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("invalid: site: %s", err.Error())
	}

	// get posting profile
	pa := new(ficomodel.PostingApproval)
	if journal.Status != ficomodel.JournalStatusDraft {
		if err := param.Hub.GetByFilter(pa, dbflex.Eq("SourceID", journal.ID)); err != nil {
			return nil, fmt.Errorf("error when get posting approval: %s", err.Error())
		}
	}

	userIDs := make([]string, 0)
	userIDs = append(userIDs, journal.CreatedBy)
	for _, appr := range pa.Approvals {
		userIDs = append(userIDs, appr.UserID)
	}

	for _, post := range pa.Postingers {
		userIDs = append(userIDs, post.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err := param.Hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", userIDs...),
	), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	// set header
	var datas [][]string
	// datas = append(datas, []string{"", "","","","",param.Header.AddressAndTax})
	datas = append(datas, []string{fmt.Sprintf("No Reff : %s", param.Header.ID), "", fmt.Sprintf("Nama Bank : %s", cashbank.BankName)})
	datas = append(datas, []string{fmt.Sprintf("Tanggal Pengajuan : %s", param.Header.TrxDate.Format("02 January 2006")), "", fmt.Sprintf("No Rekening : %s", cashbank.BankAccountNo)})
	datas = append(datas, []string{fmt.Sprintf("Site : %s", site.Label), "", fmt.Sprintf("Nama Rekening : %s", cashbank.BankAccountName)})

	var dataMobile [][]string
	dataMobile = append(dataMobile, []string{"No Reff : ", param.Header.ID})
	dataMobile = append(dataMobile, []string{"Nama Bank : ", cashbank.BankName})
	dataMobile = append(dataMobile, []string{"Tanggal Pengajuan : ", param.Header.TrxDate.Format("02 January 2006")})
	dataMobile = append(dataMobile, []string{"No Rekening : ", cashbank.BankAccountNo})
	dataMobile = append(dataMobile, []string{"Nama Rekening : ", cashbank.BankName})
	dataMobile = append(dataMobile, []string{"Site : ", site.Label})

	param.Preview.HeaderMobile.Data = dataMobile
	if param.Preview.Header == nil {
		param.Preview.Header = codekit.M{}
	}
	param.Preview.Header["Data"] = datas

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := param.Hub.GetByID(curr, param.Header.CurrencyID); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("currency not found: %s", param.Header.CurrencyID))
	}

	sectionLine := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Purchase Line",
		Items:       make([][]string, len(param.Lines)+1),
	}
	sectionLine.Items[0] = []string{"No:R", "Keterangan", "Nominal:R"}
	var subTotalAmount float64
	for i, l := range param.Lines {
		amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.Amount, 2, ",", "."))
		sectionLine.Items[i+1] = []string{strconv.Itoa(i + 1), l.Text, amount}
		subTotalAmount += l.Amount
	}

	totalAmount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(subTotalAmount, 2, ",", "."))

	var footer [][]string
	footer = append(footer, []string{"", "Jumlah", totalAmount})

	var footerMobile [][]string
	footerMobile = append(footerMobile, []string{"Jumlah : ", totalAmount})

	param.Preview.Header["Footer"] = footer
	param.Preview.HeaderMobile.Footer = footerMobile

	// signature
	param.Preview.MultipleRowSignature = make([][]tenantcoremodel.Signature, 3)
	// Prepared by
	param.Preview.MultipleRowSignature[0] = append(param.Preview.MultipleRowSignature[0], tenantcoremodel.Signature{
		ID:        journal.CreatedBy,
		Header:    "Prepared By",
		Footer:    mapUser[journal.CreatedBy],
		Confirmed: "Tgl. " + journal.Created.Format("02-January-2006 15:04:05"),
		Status:    "",
	})

	// checked by
	for _, appr := range pa.Approvals {
		if appr.Confirmed != nil {
			param.Preview.MultipleRowSignature[0] = append(param.Preview.MultipleRowSignature[0], tenantcoremodel.Signature{
				ID:        appr.UserID,
				Header:    "Approved By",
				Footer:    mapUser[appr.UserID],
				Confirmed: "Tgl. " + appr.Confirmed.Format("02-January-2006 15:04:05"),
				Status:    "",
			})
		}
	}

	// approved by
	for _, post := range pa.Postingers {
		for _, userID := range post.UserIDs {
			param.Preview.MultipleRowSignature[0] = append(param.Preview.MultipleRowSignature[0], tenantcoremodel.Signature{
				ID:        userID,
				Header:    "Approved By",
				Footer:    mapUser[userID],
				Confirmed: "",
				Status:    "",
			})
		}
	}

	param.Preview.Sections = append(param.Preview.Sections, sectionLine)

	return param, nil
}

func (p *cashPosting) Post(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine, trxs map[string][]orm.DataModel) (string, error) {
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

	// check balance
	cts := FromDataModels(p.trxs[new(ficomodel.CashTransaction).TableName()], new(ficomodel.CashTransaction))
	if len(cts) == 0 {
		return "", fmt.Errorf("missing: cash transaction")
	}
	for _, ct := range cts {
		ct.Status = ficomodel.AmountConfirmed
	}
	ctTotals := lo.MapEntries(lo.GroupBy(cts, func(ct *ficomodel.CashTransaction) string {
		return ct.CashBank.ID
	}), func(bankID string, values []*ficomodel.CashTransaction) (string, float64) {
		return bankID, lo.SumBy(values, func(c *ficomodel.CashTransaction) float64 {
			return c.Amount
		})
	})
	for bankID, amount := range ctTotals {
		if amount > 0 {
			continue
		}

		bals, _ := NewCashBalanceHub(db).Get(nil, CashBalanceOpt{
			CompanyID:  p.header.CompanyID,
			AccountIDs: []string{bankID},
			Dimension:  p.header.Dimension,
		})

		total := lo.SumBy(bals, func(bal *ficomodel.CashBalance) float64 {
			return bal.Balance
		})

		if total < amount {
			bank, _ := datahub.GetByID(db, new(tenantcoremodel.CashBank), bankID)
			if !bank.AllowNegative {
				return "", fmt.Errorf("balance not available: %s, %s", bankID, bank.Name)
			}
		}
	}

	// delete reserved or planned cash transaction
	db.DeleteByFilter(new(ficomodel.CashTransaction),
		dbflex.Eqs("CompanyID", p.header.CompanyID,
			"SourceType", cts[0].SourceType, "SourceJournalID", cts[0].SourceJournalID))

	// save all trxs
	res, err = PostModelSave(db, p.header, "CustomerVoucherNo", trxs)
	return res, err
}

func (p *cashPosting) Approved() error {
	cts := FromDataModels(p.trxs[new(ficomodel.CashTransaction).TableName()], new(ficomodel.CashTransaction))
	db := p.opt.Db
	for _, ct := range cts {
		ct.Status = lo.Ternary(ct.Amount > 0, ficomodel.AmountPlanned, ficomodel.AmountReserved)
		db.Save(ct)

		_, e := NewCustomerBalanceHub(db).Sync(nil, CustomerBalanceOpt{
			CompanyID:  ct.CompanyID,
			AccountIDs: []string{ct.CashBank.ID},
		})
		if e != nil {
			return fmt.Errorf("fail: approved calc balance: %s", e.Error())
		}
	}
	return nil
}

func (p *cashPosting) Rejected() error {
	return nil
}

func (p *cashPosting) GetAccount() string {
	return p.cb.ID
}

func (p *cashPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				TrxType:                  p.header.CashJournalType,
				Text:                     p.header.Text,
				Menu:                     p.header.CashJournalType,
				Amount:                   p.header.Amount,
				UserTo:                   app.UserID,
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
				IsApproval:               true,
			}

			switch p.header.CashJournalType {
			case "SubmissionTypePettyCash":
				notification.Menu = "Pety Cash Submission"
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

func (p *cashPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
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
