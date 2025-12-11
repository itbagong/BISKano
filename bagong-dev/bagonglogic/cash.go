package bagonglogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ekokurniadi/terbilang"
	"github.com/leekchan/accounting"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CashHandler struct {
}

type GenerateCashOutVendorPayload struct {
	JournalID string
	ApplyTo   []GetApplyToResponse
}

type GetApplyToResponse struct {
	Module    string
	JournalID string
}

func (m *CashHandler) GenerateCashOutVendor(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := GetURLQueryParams(ctx)

	// get cash out journal
	journal := new(ficomodel.CashJournal)
	if err := h.GetByID(journal, param["JournalID"]); err != nil {
		return nil, fmt.Errorf("error when get journal: %s", err.Error())
	}

	if journal.JournalTypeID == "CJT-004" {
		// if SITE
		return m.generateCashOutSite(ctx)
	} else {
		// if VENDOR
		return m.generateCashOutDefaultVendor(ctx)
	}
}

func (m *CashHandler) generateCashOutDefaultVendor(ctx *kaos.Context) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := GetURLQueryParams(ctx)

	// get cash out journal
	journal := new(ficomodel.CashJournal)
	if err := h.GetByID(journal, param["JournalID"]); err != nil {
		return nil, fmt.Errorf("error when get journal: %s", err.Error())
	}

	vendorID := ""
	bankID := ""
	for _, l := range journal.Lines {
		vendorID = l.Account.AccountID
		bankID = l.ChequeGiroID
	}
	// get vendor
	vnd := new(tenantcoremodel.Vendor)
	if err := h.GetByID(vnd, vendorID); err != nil {
		return nil, fmt.Errorf("error when get vendor: %s", err.Error())
	}

	// get main balance account
	ledgerAccount := new(tenantcoremodel.LedgerAccount)
	if err := h.GetByID(ledgerAccount, vnd.MainBalanceAccount); err != nil {
		return nil, fmt.Errorf("error when get main balance account: %s", err.Error())
	}

	// get vendor
	bgVendor := new(bagongmodel.BGVendor)
	if err := h.GetByID(bgVendor, vnd.ID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}
	// bank
	bank := bagongmodel.VendorBank{}
	for _, v := range bgVendor.VendorBank {
		if v.ID == bankID {
			bank = v
			break
		}
	}

	// get vendor
	cashBank := new(tenantcoremodel.CashBank)
	if err := h.GetByID(cashBank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := h.GetByID(curr, journal.CurrencyID); err != nil {
		return nil, fmt.Errorf("error when get currency: %s", err.Error())
	}

	// get site
	site := new(tenantcoremodel.DimensionMaster)
	if err := h.GetByID(site, journal.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	// get posting profile
	pa := new(ficomodel.PostingApproval)
	if journal.Status != ficomodel.JournalStatusDraft {
		if err := h.GetByFilter(pa, dbflex.Eq("SourceID", journal.ID)); err != nil {
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
	err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", userIDs...),
	), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	preview, _ := tenantcorelogic.GetPreviewBySource(h, "CASH OUT", journal.ID, "", journal.JournalTypeID)
	preview.Name = journal.JournalTypeID
	preview.PreviewReport = &tenantcoremodel.PreviewReport{}
	preview.PreviewReport.Header = codekit.M{}
	preview.PreviewReport.Sections = make([]tenantcoremodel.PreviewSection, 0)

	headers := make([][]string, 4)
	headers[0] = []string{"Nama Vendor :", vnd.Name, "", "", "No Voucher :", journal.ID}
	headers[1] = []string{"Nama Bank :", bank.BankName, "", "", "Tgl Jatuh Tempo :", ""}

	// set due date
	if len(journal.Lines) > 0 {
		applies := []ficomodel.CashApply{}
		err := h.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
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
			err := h.Gets(new(ficomodel.VendorJournal), dbflex.NewQueryParam().SetWhere(
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

	headers[2] = []string{"No Rekening :", bank.BankAccountNo, "", "", "Bank/Cash :", cashBank.Name}
	headers[3] = []string{"Nama Rekening :", bank.BankAccountName, "", "", "Site :", site.Label}
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
	h.Save(preview)

	return preview.PreviewReport, nil

}

func (m *CashHandler) generateCashOutSite(ctx *kaos.Context) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := GetURLQueryParams(ctx)

	// get cash out journal
	journal := new(ficomodel.CashJournal)
	if err := h.GetByID(journal, param["JournalID"]); err != nil {
		return nil, fmt.Errorf("error when get journal: %s", err.Error())
	}

	vendorID := ""
	for _, l := range journal.Lines {
		vendorID = l.Account.AccountID
	}
	// get vendor
	vnd := new(tenantcoremodel.Vendor)
	if err := h.GetByID(vnd, vendorID); err != nil {
		return nil, fmt.Errorf("error when get vendor: %s", err.Error())
	}

	// get main balance account
	ledgerAccount := new(tenantcoremodel.LedgerAccount)
	if err := h.GetByID(ledgerAccount, vnd.MainBalanceAccount); err != nil {
		return nil, fmt.Errorf("error when get main balance account: %s", err.Error())
	}

	// get vendor
	bgVendor := new(bagongmodel.BGVendor)
	if err := h.GetByID(bgVendor, vnd.ID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// bank
	bank := new(tenantcoremodel.CashBank)
	if err := h.GetByID(bank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get vendor
	cashBank := new(tenantcoremodel.CashBank)
	if err := h.GetByID(cashBank, journal.CashBookID); err != nil {
		return nil, fmt.Errorf("error when get vendor detail: %s", err.Error())
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := h.GetByID(curr, journal.CurrencyID); err != nil {
		return nil, fmt.Errorf("error when get currency: %s", err.Error())
	}

	// get site
	site := new(tenantcoremodel.DimensionMaster)
	if err := h.GetByID(site, journal.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	// get posting profile
	pa := new(ficomodel.PostingApproval)
	if journal.Status != ficomodel.JournalStatusDraft {
		if err := h.GetByFilter(pa, dbflex.Eq("SourceID", journal.ID)); err != nil {
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
	err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", userIDs...),
	), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(detail tenantcoremodel.Employee) (string, string) {
		return detail.ID, detail.Name
	})

	preview, _ := tenantcorelogic.GetPreviewBySource(h, "CASH OUT", journal.ID, "", journal.JournalTypeID)
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
		err := h.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
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
		err = h.Gets(new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(
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

			err := h.Gets(new(ficomodel.VendorJournal), dbflex.NewQueryParam().SetWhere(
				dbflex.In("_id", vendorJournalIDs...),
			).SetSort("-Created"), &vendors)
			if err != nil {
				return nil, fmt.Errorf("error when get vendor journal: %s", err.Error())
			}

			// if len(vendors) == 1 {
			// 	if vendors[0].ExpectedDate != nil {
			// 		headers[1][5] = vendors[0].ExpectedDate.Format("02-01-2006")
			// 	}
			// }
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
	err = h.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(
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
	e := h.GetsByFilter(new(tenantcoremodel.LedgerAccount), dbflex.In("_id", ledgerAccountIds...), &ledgerAccounts)
	if e != nil {
		ctx.Log().Errorf("Failed populate data ledger accounts: %s", e.Error())
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

	// Save the preview report
	h.Save(preview)

	return preview.PreviewReport, nil
}
