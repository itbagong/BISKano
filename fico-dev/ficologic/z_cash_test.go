package ficologic_test

import (
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestCashTrxAndBalance(t *testing.T) {
	convey.Convey("prepare and sync nil", t, conveyCashTrxAndBalance)
}

func conveyCashTrxAndBalance() {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "cash-user-01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	dt5Jan := time.Date(2023, 1, 5, 0, 0, 0, 0, time.Local)

	InsertModel(db, []*ficomodel.CashTransaction{
		{CompanyID: testCoID1, SourceType: "Cash Journal", SourceJournalID: "CJX01", SourceLineNo: 1, TrxDate: dt5Jan,
			CashBank:  tenantcoremodel.CashBank{ID: "C01"},
			Dimension: tenantcoremodel.Dimension{}.Set("PC", "MINIG").Set("Site", "KJA"),
			Status:    ficomodel.AmountConfirmed, Amount: 1000, Text: "CJX01 Line 01"},
		{CompanyID: testCoID1, SourceType: "Cash Journal", SourceJournalID: "CJX01", SourceLineNo: 2, TrxDate: dt5Jan,
			CashBank:  tenantcoremodel.CashBank{ID: "C01"},
			Dimension: tenantcoremodel.Dimension{}.Set("PC", "MINIG").Set("Site", "KJA"),
			Status:    ficomodel.AmountConfirmed, Amount: 2000, Text: "CJX01 Line 01"},
		{CompanyID: testCoID1, SourceType: "Cash Journal", SourceJournalID: "CJX02", SourceLineNo: 1, TrxDate: dt5Jan,
			CashBank:  tenantcoremodel.CashBank{ID: "C01"},
			Dimension: tenantcoremodel.Dimension{}.Set("PC", "MINIG").Set("Site", "HO"),
			Status:    ficomodel.AmountConfirmed, Amount: 2000, Text: "CJX02 Line 01"},
	})

	balHub := ficologic.NewCashBalanceHub(db)
	_, err := balHub.Sync(&dt5Jan, ficologic.CashBalanceOpt{CompanyID: testCoID1,
		GroupByDimension: []string{"PC", "CC", "Site", "CashBookID"},
		//Dimension:        tenantcoremodel.Dimension{}.Set("Site", "KJA"),
	})
	convey.SoMsg("sync process should return nil", err, convey.ShouldBeNil)
	balC01, _ := datahub.GetByFilter(db, new(ficomodel.CashBalance),
		dbflex.And(
			dbflex.Eqs("CompanyID", testCoID1, "CashBookID", "C01", "BalanceDate", dt5Jan),
			tenantcoremodel.Dimension{}.Set("Site", "KJA").Where(),
		))
	convey.SoMsg("balance = 3000 with date", balC01.Balance, convey.ShouldEqual, 3000)
	balC01Nil, _ := datahub.GetByFilter(db, new(ficomodel.CashBalance),
		dbflex.And(
			dbflex.Eqs("CompanyID", testCoID1, "CashBookID", "C01", "BalanceDate", nil),
			tenantcoremodel.Dimension{}.Set("Site", "HO").Where(),
		))
	convey.SoMsg("balance = 3000 with date = nil", balC01Nil.Balance, convey.ShouldEqual, 2000)

	convey.Convey("reserve amount and split", conveyCashTrxSplit)
}

func conveyCashTrxSplit() {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "cash-user-01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	dt8Jan := time.Date(2023, 1, 8, 0, 0, 0, 0, time.Local)

	InsertModel(db, []*ficomodel.CashTransaction{
		{CompanyID: testCoID1, SourceType: "Cash Journal", SourceJournalID: "CJX10", SourceLineNo: 1, TrxDate: dt8Jan,
			CashBank:  tenantcoremodel.CashBank{ID: "C01"},
			Dimension: tenantcoremodel.Dimension{}.Set("PC", "MINIG").Set("Site", "KJA"),
			Status:    ficomodel.AmountReserved, Amount: -2500, Text: "CJX10 Line 01"},
		{CompanyID: testCoID1, SourceType: "Cash Journal", SourceJournalID: "CJX11", SourceLineNo: 1, TrxDate: dt8Jan,
			CashBank:  tenantcoremodel.CashBank{ID: "C01"},
			Dimension: tenantcoremodel.Dimension{}.Set("PC", "MINIG").Set("Site", "HO"),
			Status:    ficomodel.AmountReserved, Amount: -6000, Text: "CJX11 Line 01"},
	})

	splitter := ficologic.NewCashTrxSplitter(db)
	_, _, err := splitter.Split(-1000, string(ficomodel.AmountConfirmed), ficologic.CashSplitOpt{
		CompanyID:       testCoID1,
		SourceType:      "Cash Journal",
		SourceJournalID: "CJX10",
		SourceLineNo:    1,
		SourceStatus:    string(ficomodel.AmountReserved),
	})

	convey.So(err, convey.ShouldBeNil)
	cjx02res, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction),
		dbflex.Eqs("SourceJournalID", "CJX10", "Status", ficomodel.AmountReserved))
	convey.So(cjx02res.Amount, convey.ShouldEqual, -1500)
	cjx02cfm, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction),
		dbflex.Eqs("SourceJournalID", "CJX10", "Status", ficomodel.AmountConfirmed))
	convey.So(cjx02cfm.Amount, convey.ShouldEqual, -1000)

	balHub := ficologic.NewCashBalanceHub(db)
	balHub.Sync(nil, ficologic.CashBalanceOpt{CompanyID: testCoID1,
		GroupByDimension: []string{"PC", "CC", "Site", "CashBookID"},
		//Dimension:        tenantcoremodel.Dimension{}.Set("Site", "KJA"),
	})
	balC01, _ := datahub.GetByFilter(db, new(ficomodel.CashBalance),
		dbflex.And(
			dbflex.Eqs("CompanyID", testCoID1, "CashBookID", "C01", "BalanceDate", nil),
			tenantcoremodel.Dimension{}.Set("Site", "KJA").Where(),
		))
	convey.SoMsg("balance = 2000", balC01.Balance, convey.ShouldEqual, 2000)

	convey.Convey("balance get", conveyCashGetBalance)
}

func conveyCashGetBalance() {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "cash-user-01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	dt6Jan := time.Date(2023, 1, 6, 0, 0, 0, 0, time.Local)
	dt9Jan := time.Date(2023, 1, 9, 0, 0, 0, 0, time.Local)

	bal := ficologic.NewCashBalanceHub(db)

	balBySites, err := bal.Get(&dt6Jan, ficologic.CashBalanceOpt{
		CompanyID:        testCoID1,
		GroupByDimension: []string{"Site"},
	})
	convey.So(err, convey.ShouldBeNil)
	kja6jan, _ := lo.Find(balBySites, func(b *ficomodel.CashBalance) bool {
		return b.Dimension.Get("Site") == "KJA"
	})
	convey.SoMsg("balance at 6 jan = 3000", kja6jan.Balance, convey.ShouldEqual, 3000)

	balBySites, err = bal.Get(&dt9Jan, ficologic.CashBalanceOpt{
		CompanyID:        testCoID1,
		GroupByDimension: []string{"Site"},
	})
	convey.So(err, convey.ShouldBeNil)
	kja9jan, _ := lo.Find(balBySites, func(b *ficomodel.CashBalance) bool {
		return b.Dimension.Get("Site") == "KJA"
	})
	convey.SoMsg("balance at 9 jan = 2000", kja9jan.Balance, convey.ShouldEqual, 2000)
}
