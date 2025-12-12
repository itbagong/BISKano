package ficologic_test

import (
	"fmt"
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

func TestPostingProfileApprovers(t *testing.T) {
	convey.Convey("prepare line", t, func() {
		ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "user01", testCoID1)
		db := sebar.GetTenantDBFromContext(ctx)

		dim := tenantcoremodel.Dimension{{Key: "CC", Value: "OPS"}, {Key: "Site", Value: "KJA"}}
		lines := []ficomodel.JournalLine{
			{LineNo: 1, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "555555"}, Amount: 250000},
			{LineNo: 2, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "888888"}, Amount: 320000},
		}
		approvers := ficologic.GetApproversByJournalLine(db, "DEF", dim, lines)
		convey.So(len(approvers), convey.ShouldEqual, 3)
		convey.So(approvers[0].UserIDs[0], convey.ShouldEqual, "pjo-kja")
		convey.So(approvers[1].UserIDs[0], convey.ShouldEqual, "finance")
		convey.So(approvers[2].UserIDs[0], convey.ShouldEqual, "cxo")
		// convey.Println()
		// convey.Println("========")
		// convey.Println("approvers:", codekit.JsonString(approvers))
		// convey.Println("========")
	})
}

func TestLedgerJournalRejectPost(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	prepareCtxData(ctx, "user01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)

	convey.Convey("create journal", t, func() {
		lj := &ficomodel.LedgerJournal{
			CompanyID:     testCoID1,
			TrxDate:       time.Date(2023, 1, 10, 0, 0, 0, 0, time.Now().Location()),
			JournalTypeID: "GEN",
			Text:          "Test ledger journal 1",
			DefaultOffset: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "899999"},
			Dimension: tenantcoremodel.Dimension{
				{Key: "PC", Value: "PC-NONE"},
				{Key: "CC", Value: "OPS"},
				{Key: "Site", Value: "KJA"},
			},
			Lines: []ficomodel.JournalLine{
				{LineNo: 1, Amount: 350000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "510000"}},
				{LineNo: 2, Amount: 500000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "521000"}}},
		}

		lj, e := insertLedgerJournal(lj)
		convey.So(e, convey.ShouldBeNil)
		convey.So(lj.ID, convey.ShouldStartWith, "GNJ00")

		convey.Convey("submit journal for approval", func() {
			prepareCtxData(ctx, "user01", testCoID1)
			pph := new(ficologic.PostingProfileHandler)
			_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpSubmit}})
			convey.So(err, convey.ShouldBeNil)
			submitted, _ := datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, lj.CompanyID, string(ficomodel.SubledgerAccounting), lj.ID, true)
			convey.So(pa, convey.ShouldNotBeNil)
			convey.So(len(pa.Approvers), convey.ShouldEqual, 3)

			convey.Convey("reject lv1", func() {
				prepareCtxData(ctx, "pjo-kja", testCoID1)
				_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpReject, Text: "test reject"}})
				convey.So(err, convey.ShouldBeNil)
				submitted, _ = datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
				convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusRejected)
			})
		})
	})
}

func TestLedgerJournalApprovePost(t *testing.T) {
	convey.Convey("working with ledger journal", t, func() {
		ctx := kaos.NewContextFromService(svc, nil)
		prepareCtxData(ctx, "user01", testCoID1)
		db := sebar.GetTenantDBFromContext(ctx)

		convey.Convey("create journal", func() {
			lj := &ficomodel.LedgerJournal{
				CompanyID:        testCoID1,
				TrxDate:          time.Date(2023, 1, 10, 0, 0, 0, 0, time.Now().Location()),
				JournalTypeID:    "GEN",
				PostingProfileID: "DEF",
				Text:             "Test ledger journal 1",
				DefaultOffset:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "899999"},
				Dimension: tenantcoremodel.Dimension{
					{Key: "PC", Value: "PC-NONE"},
					{Key: "CC", Value: "OPS"},
					{Key: "Site", Value: "KJA"},
				},
				Lines: []ficomodel.JournalLine{
					{LineNo: 1, Amount: 350000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "510000"}},
					{LineNo: 2, Amount: 500000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "521000"}}},
			}

			lj, e := insertLedgerJournal(lj)
			convey.So(e, convey.ShouldBeNil)
			convey.So(lj.ID, convey.ShouldStartWith, "GNJ00")
			// convey.Println()
			// convey.Println("======================================")
			// convey.Println("ledger journal:", lj.ID)
			// convey.Println("ledger journal:", lj.Dimension.ToString())
			// convey.Println("======================================")

			convey.Convey("submit journal for approval", func() {
				prepareCtxData(ctx, "user01", testCoID1)
				pph := new(ficologic.PostingProfileHandler)
				_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpSubmit}})
				convey.So(err, convey.ShouldBeNil)
				submitted, _ := datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
				convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

				pa, _ := ficologic.GetPostingApprovalBySource(db, lj.CompanyID, string(ficomodel.SubledgerAccounting), lj.ID, true)
				convey.So(pa, convey.ShouldNotBeNil)
				convey.So(len(pa.Approvers), convey.ShouldEqual, 3)

				convey.Convey("approve lv1", func() {
					prepareCtxData(ctx, "pjo-kja", testCoID1)
					_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpApprove}})
					convey.So(err, convey.ShouldBeNil)
					submitted, _ = datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
					convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

					pa, _ := ficologic.GetPostingApprovalBySource(db, lj.CompanyID, string(ficomodel.SubledgerAccounting), lj.ID, true)
					convey.So(pa, convey.ShouldNotBeNil)
					convey.So(pa.Approvals[0].Status, convey.ShouldEqual, string(ficomodel.JournalStatusApproved))
					convey.So(len(pa.Approvals), convey.ShouldEqual, 2)
					convey.So(pa.Approvals[1].UserID, convey.ShouldEqual, "finance")

					convey.Convey("approve lv2", func() {
						prepareCtxData(ctx, "finance", testCoID1)
						_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpApprove}})
						convey.So(err, convey.ShouldBeNil)
						submitted, _ = datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
						convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

						pa, _ := ficologic.GetPostingApprovalBySource(db, lj.CompanyID, string(ficomodel.SubledgerAccounting), lj.ID, true)
						convey.So(pa, convey.ShouldNotBeNil)
						convey.So(pa.Approvals[1].Status, convey.ShouldEqual, string(ficomodel.JournalStatusApproved))
						convey.So(pa.Approvals[2].UserID, convey.ShouldEqual, "cxo")

						convey.Convey("approve lv3 - final", func() {
							prepareCtxData(ctx, "cxo", testCoID1)
							_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpApprove}})
							convey.So(err, convey.ShouldBeNil)
							submitted, _ = datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
							convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusReady)

							pa, _ := ficologic.GetPostingApprovalBySource(db, lj.CompanyID, string(ficomodel.SubledgerAccounting), lj.ID, false)
							convey.So(pa, convey.ShouldNotBeNil)
							convey.So(pa.Approvals[2].Status, convey.ShouldEqual, string(ficomodel.JournalStatusApproved))

							convey.Convey("post journal - using invalid user - shld fail", func() {
								prepareCtxData(ctx, "user01", testCoID1)
								_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpPost}})
								convey.So(err, convey.ShouldNotBeNil)

								convey.Convey("post journal - shld ok", func() {
									prepareCtxData(ctx, "finance-exec", testCoID1)
									_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpPost}})
									convey.So(err, convey.ShouldBeNil)

									posted, _ := datahub.GetByID(db, new(ficomodel.LedgerJournal), lj.ID)
									convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

									trxs, _ := datahub.FindByFilter(db, new(ficomodel.LedgerTransaction),
										dbflex.Eqs("SourceType", string(ficomodel.SubledgerAccounting), "SourceJournalID", lj.ID))
									debitSum := lo.SumBy(trxs, func(trx *ficomodel.LedgerTransaction) float64 {
										return lo.Ternary(trx.Amount > 0, trx.Amount, 0)
									})
									convey.So(debitSum, convey.ShouldEqual, 850000)

									convey.Convey("check balance", func() {
										balance := ficologic.NewLedgerBalanceCalc(db, lj.CompanyID)
										bal := balance.Get(nil, ficologic.LedgerBalanceGetOpts{
											LedgerAccounts: []string{"510000"},
											DimNames:       []string{"CC", "Site"},
										})
										convey.So(len(bal), convey.ShouldBeGreaterThan, 0)
										convey.So(bal[0].Balance, convey.ShouldEqual, 350000)
										convey.So(bal[0].Dimension.Get("Site"), convey.ShouldEqual, "KJA")
										// convey.Println()
										// convey.Println("balance:", codekit.JsonString(bal[0]))
									})
								})
							})
						})
					})
				})
			})
		})
	})
}

func TestLedgerJournalPostWithoutApproval(t *testing.T) {

}

func insertLedgerJournal(lj *ficomodel.LedgerJournal) (*ficomodel.LedgerJournal, error) {
	uriPath := "/v1/fico/ledgerjournal/insert"
	sr := svc.GetRoute(uriPath)
	if sr == nil {
		return nil, fmt.Errorf("missing: route: %s", uriPath)
	}
	ctx := kaos.NewContextFromService(svc, sr)
	prepareCtxData(ctx, "user01", testCoID1)
	nlj := new(ficomodel.LedgerJournal)
	e := svc.CallTo(uriPath, nlj, ctx, lj)
	return nlj, e
}
