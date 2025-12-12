package ficologic_test

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

func prepareCtxData(ctx *kaos.Context, userid, companyid string) *kaos.Context {
	if userid != "" {
		ctx.Data().Set("jwt_reference_id", userid)
	}

	if companyid == "" {
		companyid = testCoID1
	}
	ctx.Data().Set("jwt_data", codekit.M{}.Set("CompanyID", companyid))
	return ctx
}

func injectLedgerBankMasterData(ctx *kaos.Context) error {
	prepareCtxData(ctx, "", "")
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("missing: db")
	}

	cos := []tenantcoremodel.Company{
		{ID: testCoID1, Name: "Test company 1"},
		{ID: testCoID2, Name: "Test company 2"}}

	for _, co := range cos {
		if e := h.Insert(&co); e != nil {
			return fmt.Errorf("company: %s", e.Error())
		}
	}

	ledgers := []tenantcoremodel.LedgerAccount{
		{ID: "100000", Name: "Asset", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "150000", Name: "Cash & Bank", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "151000", Name: "BRI-01", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "200000", Name: "Customer", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "300000", Name: "Liability", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "310100", Name: "Utang umum", AccountType: tenantcoremodel.BalanceSheetAccount},
		{ID: "320000", Name: "Utang jangka pendek", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "340000", Name: "Utang jangka panjang", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "399999", Name: "Modal", AccountType: tenantcoremodel.BalanceSheetAccount},
		{ID: "400000", Name: "Pendapatan", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "400001", Name: "Pendapatan Mining", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "400002", Name: "Pendapatan Trayek", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "400003", Name: "Pendapatan Pariwisata", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "400004", Name: "Pendapatan BTS", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "400005", Name: "Pendapatan Trading", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: true},
		{ID: "499998", Name: "Pendapatan Lain-lain", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "499999", Name: "Pendapatan Lain-lain", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "500000", Name: "Biaya langsung", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "510000", Name: "BBM", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "521000", Name: "Perbaikan dan Perawatan - Reguler", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "522000", Name: "Perbaikan dan Perawatan - Kerusakan", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "531000", Name: "Tenaga Kerja - Unit", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "532000", Name: "Tenaga Kerja - Mekanik", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "541000", Name: "Material - Oli", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "542000", Name: "Material - Ban", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "543000", Name: "Material - Lain lain", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "580000", Name: "Depresiasi Unit", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "599991", Name: "Biaya Mining", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "599992", Name: "Biaya Pariwisata", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "599993", Name: "Biaya AKDP", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "599994", Name: "Biaya BTS", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "599999", Name: "Biaya lain-lain unit", AccountType: tenantcoremodel.ProfitLossAccount, BlockedFromGJ: false},
		{ID: "610000", Name: "Overhead Site", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "620000", Name: "Overhead Office", AccountType: tenantcoremodel.HeaderAccount},
		{ID: "799999", Name: "Other P&L", AccountType: tenantcoremodel.ProfitLossAccount},
		{ID: "210001", Name: "Customer Transaction", AccountType: tenantcoremodel.BalanceSheetAccount},
		{ID: "899999", Name: "Other B&S", AccountType: tenantcoremodel.BalanceSheetAccount}}
	for _, ledger := range ledgers {
		ledger.IsActive = true
		if e := h.Insert(&ledger); e != nil {
			return fmt.Errorf("ledger: %s, %s: %s", ledger.ID, ledger.Name, e.Error())
		}
	}

	// dimensions
	dims := []tenantcoremodel.DimensionMaster{
		{DimensionType: "CC", ID: "CXO", Label: "CXO", IsActive: true},
		{DimensionType: "CC", ID: "OPS", Label: "OPS", IsActive: true},
		{DimensionType: "CC", ID: "SLS", Label: "SLS", IsActive: true},
		{DimensionType: "CC", ID: "MTG", Label: "MTG", IsActive: true},
		{DimensionType: "CC", ID: "FIN", Label: "FIN", IsActive: true},
		{DimensionType: "CC", ID: "SCM", Label: "SCM", IsActive: true},
		{DimensionType: "CC", ID: "MFG", Label: "MFG", IsActive: true},
		{DimensionType: "CC", ID: "CC-NONE", Label: "NONE", IsActive: true},
		{DimensionType: "Site", ID: "M1", Label: "M1", IsActive: true},
		{DimensionType: "Site", ID: "M2", Label: "M2", IsActive: true},
		{DimensionType: "Site", ID: "T1", Label: "T1", IsActive: true},
		{DimensionType: "Site", ID: "T2", Label: "T2", IsActive: true},
		{DimensionType: "Site", ID: "ST-NONE", Label: "NONE", IsActive: true},
		{ID: "PC-Mining", DimensionType: "PC", Label: "Mining", IsActive: true},
		{ID: "PC-Trayek", DimensionType: "PC", Label: "Trayek", IsActive: true},
		{ID: "PC-Other", DimensionType: "PC", Label: "Other", IsActive: true},
		{ID: "PC-None", DimensionType: "PC", Label: "NONE", IsActive: true}}
	for _, obj := range dims {
		if e := h.Insert(&obj); e != nil {
			return fmt.Errorf("%s: %s: %s", obj.TableName(), obj.ID, e.Error())
		}
	}

	// expense type
	if e := InsertModel(h, []*tenantcoremodel.ExpenseType{
		{ID: "Mining", Name: "BIaya Mining", LedgerAccountID: "599991"},
		{ID: "Pariwisata", Name: "BIaya Pariwisata", LedgerAccountID: "599992"},
		{ID: "ADKP", Name: "BIaya ADKP", LedgerAccountID: "599993"},
		{ID: "BTS", Name: "BIaya BTS", LedgerAccountID: "599994"},
	}); e != nil {
		return e
	}

	// cash banks
	banks := []orm.DataModel{
		&tenantcoremodel.CashBankGroup{ID: "OPS", Name: "OPS"},
		&tenantcoremodel.CashBank{ID: "BRI01", Name: "BRI 01", CashBankGroupID: "OPS", CurrencyID: "IDR", MainBalanceAccount: "151000", IsActive: true},
		&tenantcoremodel.CashBank{ID: "BRI02", Name: "BRI 02", CashBankGroupID: "OPS", CurrencyID: "IDR", MainBalanceAccount: "151000", IsActive: true}}
	for _, obj := range banks {
		if e := h.Insert(obj); e != nil {
			return fmt.Errorf("%s: %s: %s", obj.TableName(), modelID(obj), e.Error())
		}
	}

	return nil
}

func modelID(d orm.DataModel) string {
	_, ids := d.GetID(nil)
	if len(ids) == 0 {
		return "NoID"
	}
	return fmt.Sprintf("%v", ids[0])
}

func injectLedgerBankConfigData(ctx *kaos.Context) error {
	prepareCtxData(ctx, "", testCoID1)
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("missing: db")
	}

	// number sequence
	nseqs := []*tenantcoremodel.NumberSequence{
		{ID: "VCH", Name: "Ledger Voucher", OutFormat: "VCH%08d"},
		{ID: "LGJ", Name: "General Journal", OutFormat: "GNJ%08d"},
		{ID: "CBJ01", Name: "Cash Bank Journal Co01", OutFormat: "CBJ01%04d"},
		{ID: "CSJ01", Name: "Customer Journal Co01", OutFormat: "CSJ01%04d"},
		{ID: "VJ23", Name: "Vendor Journal 2023", OutFormat: "VJ23%04d"},
		{ID: "CSJ02", Name: "Cash Journal Co02", OutFormat: "CSJ02%04d"},
	}
	if e := InsertModel(h, nseqs); e != nil {
		return e
	}

	nsetups := []*tenantcoremodel.NumberSequenceSetup{
		{Kind: "LedgerVoucher", NumSeqID: "VCH"},
		//{Kind: "CashBankJournal", NumSeqID: "GJ"},
		{Kind: "LedgerJournal", NumSeqID: "LGJ"},
		{Kind: "CashJournal", NumSeqID: "CBJ01", CompanyID: testCoID1},
		{Kind: "VendorJournal", NumSeqID: "VJ23"},
		{Kind: "CustomerJournal", NumSeqID: "CSJ01", CompanyID: testCoID1},
		{Kind: "CustomerJournal", NumSeqID: "CSJ02", CompanyID: testCoID2},
	}
	if e := InsertModel(h, nsetups); e != nil {
		return e
	}

	// fiscal year
	startYear := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Now().Local().Location())
	endYear := startYear.AddDate(1, 0, 0).Add(-1 * time.Second)
	yr, e := new(ficologic.FiscalYearHandler).Create(ctx,
		&ficomodel.FiscalYear{ID: "FY2023", Name: "FY20023", CompanyID: testCoID1, From: startYear, To: endYear, IsActive: true})
	if e != nil {
		return fmt.Errorf("create year: %s", e.Error())
	}
	periods, _ := datahub.FindByFilter(h, new(ficomodel.FiscalPeriod), dbflex.Eq("FiscalYearID", yr.ID))
	if len(periods) == 0 {
		return errors.New("missing: period")
	}
	for _, period := range periods {
		period.Modules = map[string]ficomodel.PeriodStatus{
			"Finance":   ficomodel.PeriodOpen,
			"Inventory": ficomodel.PeriodOpen}
		h.Save(period)
	}

	// posting profile
	pps := []*ficomodel.PostingProfile{
		{ID: "DEF", Name: "Default Posting Profile", NeedApproval: true, IsActive: true},
		{ID: "NAPV", Name: "No Approval", NeedApproval: false, IsActive: true},
		{ID: "NADP", Name: "No Approval, Direct Posting", NeedApproval: false, DirectPosting: true, IsActive: true},
		{ID: "WADP", Name: "With Approval, Direct Posting, limited", LimitSubmission: true, NeedApproval: true, DirectPosting: true, IsActive: true},
	}
	if e := InsertModel(h, pps); e != nil {
		return e
	}
	ppApprovers := []*ficomodel.PostingProfilePIC{
		{Name: "PJO KJA", PostingProfileID: "DEF", Priority: 1,
			Dimension: tenantcoremodel.Dimension{{Key: "CC", Value: "OPS"}, {Key: "Site", Value: "KJA"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"pjo-kja"}}}},
		{Name: "PJO TBG", PostingProfileID: "DEF", Priority: 1,
			Dimension: tenantcoremodel.Dimension{{Key: "CC", Value: "OPS"}, {Key: "Site", Value: "TBG"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"pjo-kja"}}}},
		{Name: "Cost checking", PostingProfileID: "DEF", Priority: 10,
			Account:   ficomodel.PostingProfileAccount{AccountType: "AccountGroup", AccountIDs: []string{"CostOps"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"ap-analyst"}}}},
		{Name: "Cost HCM", PostingProfileID: "DEF", Priority: 10,
			Account:   ficomodel.PostingProfileAccount{AccountType: "AccountGroup", AccountIDs: []string{"CostHCM"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"hcm-analyst"}}}},
		{Name: "Cost IT", PostingProfileID: "DEF", Priority: 10,
			Account:   ficomodel.PostingProfileAccount{AccountType: "AccountGroup", AccountIDs: []string{"CostIT"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"it-analyst"}}}},
		{Name: "C&B BRI01", PostingProfileID: "DEF", Priority: 10,
			Account:   ficomodel.PostingProfileAccount{AccountType: ficomodel.SubledgerCashBank, AccountIDs: []string{"BRI01"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"ho-treasury"}}}},
		{Name: "Revenue", PostingProfileID: "DEF", Priority: 10,
			Account:   ficomodel.PostingProfileAccount{AccountType: "AccountGroup", AccountIDs: []string{"Revenue"}},
			Approvers: []ficomodel.PostingUsers{{UserIDs: []string{"{revenue-analyst"}}}},
		{Name: "FI_CXO", PostingProfileID: "DEF", Priority: 999,
			Postingers: []ficomodel.PostingUsers{{UserIDs: []string{"finance", "finance-exec"}}},
			Approvers: []ficomodel.PostingUsers{
				{UserIDs: []string{"finance"}},
				{UserIDs: []string{"cxo"}}}},
		{Name: "FI", PostingProfileID: "NADP", Priority: 900, Postingers: []ficomodel.PostingUsers{{UserIDs: []string{"finance"}}}},
		{Name: "FI_CXO", PostingProfileID: "WADP", Priority: 999,
			Submitters: []ficomodel.PostingUsers{{UserIDs: []string{"finance", "finance-exec"}}},
			Approvers: []ficomodel.PostingUsers{
				{UserIDs: []string{"finance"}},
				{UserIDs: []string{"cxo"}}}},
	}
	if e := InsertModel(h, ppApprovers); e != nil {
		return e
	}
	// ledger journal types
	ljts := []*ficomodel.LedgerJournalType{
		{ID: "DEF", Name: "Default", PostingProfileID: "NADP"},
		{ID: "GEN", Name: "General Journal", PostingProfileID: "DEF"}}
	if e := InsertModel(h, ljts); e != nil {
		return e
	}

	// cash journal types
	if e := InsertModel(h, []*ficomodel.CashJournalType{
		{ID: "DEF", Name: "Default", PostingProfileID: "NADP"},
		{ID: "OPENING", Name: "Opening Balance"},
		{ID: "CO-OPS", Name: "Cash out - operational site"},
		{ID: "CO-PAYROLL", Name: "Cash out - payroll"},
	}); e != nil {
		return e
	}

	// customer journal types
	if e := InsertModel(h, []*ficomodel.CustomerJournalType{
		{ID: "DEF", Name: "Customer Default Journal Type", PostingProfileID: "NADP"},
	}); e != nil {
		return e
	}

	// vendor journal types
	if e := InsertModel(h, []*ficomodel.VendorJournalType{
		{ID: "DEF", Name: "Vendor Journal Type", PostingProfileID: "NADP"},
	}); e != nil {
		return e
	}

	// project
	if e := InsertModel(h, []*tenantcoremodel.Project{
		{ID: "Project1", Name: "Project Satue"},
		{ID: "Project2", Name: "Project Dua"},
	}); e != nil {
		return e
	}

	// expense
	if e := InsertModel(h, []*tenantcoremodel.ExpenseType{
		{ID: "Las", LedgerAccountID: "532000"},
		{ID: "Assembly", LedgerAccountID: "532000"},
		{ID: "General", LedgerAccountID: "532000"},
		{ID: "ATK", LedgerAccountID: "532000"},
	}); e != nil {
		return e
	}

	// customer group
	if e := InsertModel(h, []*tenantcoremodel.CustomerGroup{{ID: "Std", Name: "Customer Default Group",
		Setting: tenantcoremodel.CustomerSetting{
			MainBalanceAccount: "210001",
		}}}); e != nil {
		return e
	}

	// customer
	if e := InsertModel(h, []*tenantcoremodel.Customer{
		{ID: "C01", Name: "Customer 01", GroupID: "Std"},
		{ID: "C03", Name: "Customer 03 - Apply Test", GroupID: "Std"},
	}); e != nil {
		return e
	}

	// vendor group
	if e := InsertModel(h, []*tenantcoremodel.VendorGroup{{ID: "Std", Name: "Vendor Default Group",
		MainBalanceAccount: "310100",
	}}); e != nil {
		return e
	}

	// vendor
	if e := InsertModel(h, []*tenantcoremodel.Vendor{
		{ID: "V01", Name: "Vendor 01", GroupID: "Std"},
		{ID: "V03", Name: "Vendor 03", GroupID: "Std"},
	}); e != nil {
		return e
	}

	return nil
}

func InsertModel[T orm.DataModel](h *datahub.Hub, records []T) error {
	tableName := ""
	for _, obj := range records {
		if e := h.Insert(obj); e != nil {
			return fmt.Errorf("%s: %s: %s", obj.TableName(), modelID(obj), e.Error())
		}
		if tableName == "" {
			tableName = obj.TableName()
		}
	}
	return nil
}
