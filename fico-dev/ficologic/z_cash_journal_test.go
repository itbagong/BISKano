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
	"github.com/smartystreets/goconvey/convey"
)

func TestCashJournalApprovePosting(t *testing.T) {
	convey.Convey("prepare journal", t, func() {
		ctx := kaos.NewContextFromService(svc, nil)
		prepareCtxData(ctx, "user01", testCoID1)
		db := sebar.GetTenantDBFromContext(ctx)

		amt := float64(20000000)
		cj, err := insertCashJournal(&ficomodel.CashJournal{
			JournalTypeID:    "OPENING",
			TrxDate:          time.Date(2023, 1, 10, 0, 0, 0, 0, time.Now().Local().Location()),
			CurrencyID:       "IDR",
			Text:             "Test inject cash balance",
			PostingProfileID: "DEF",
			CashBookID:       "BRI01",
			Dimension:        tenantcoremodel.Dimension{}.Set("PC", "PC-NONE").Set("CC", "OPS").Set("Site", "HO"),
			Lines: []ficomodel.JournalLine{
				{LineNo: 1, OffsetAccount: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "399999"), Amount: amt},
			}})
		convey.So(err, convey.ShouldBeNil)
		convey.So(cj.CompanyID, convey.ShouldEqual, testCoID1)
		convey.So(cj.ID, convey.ShouldStartWith, "CBJ01")

		convey.Convey("submit", func() {
			pph := new(ficologic.PostingProfileHandler)
			_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpSubmit}})
			convey.So(err, convey.ShouldBeNil)
			submitted, _ := datahub.GetByID(db, new(ficomodel.CashJournal), cj.ID)
			pa, _ := ficologic.GetPostingApprovalBySource(db, cj.CompanyID, string(ficomodel.SubledgerCashBank), cj.ID, true)
			convey.So(submitted.Status, convey.ShouldEqual, "SUBMITTED")
			convey.So(len(pa.Approvers), convey.ShouldBeGreaterThan, 0)
			convey.So(pa.Approvers[0].UserIDs[0], convey.ShouldEqual, "ho-treasury")

			convey.Convey("approve", func() {
				prepareCtxData(ctx, "ho-treasury", testCoID1)
				_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpApprove}})
				convey.So(err, convey.ShouldBeNil)

				prepareCtxData(ctx, "finance", testCoID1)
				_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpApprove}})
				convey.So(err, convey.ShouldBeNil)

				prepareCtxData(ctx, "cxo", testCoID1)
				_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpApprove}})
				convey.So(err, convey.ShouldBeNil)

				convey.Convey("status = READY", func() {
					submitted, _ := datahub.GetByID(db, new(ficomodel.CashJournal), cj.ID)
					convey.So(submitted.Status, convey.ShouldEqual, "READY")

					convey.Convey("cash balance = OK", func() {
						balanceCalc := ficologic.NewCashBalanceCalc(db, cj.CompanyID)
						balances := balanceCalc.Get(nil, &ficologic.CashBalanceGetOpts{CashBookIDs: []string{"BRI01"}})
						convey.So(len(balances), convey.ShouldBeGreaterThan, 0)
						convey.So(balances[0].Planned, convey.ShouldEqual, amt)

						convey.Convey("posting", func() {
							prepareCtxData(ctx, "finance-exec", testCoID1)
							_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpPost}})
							posted, _ := datahub.GetByID(db, new(ficomodel.CashJournal), cj.ID)

							convey.SoMsg("posting err is nil", err, convey.ShouldBeNil)
							convey.SoMsg("status is READY", posted.Status, convey.ShouldEqual, "POSTED")

							bankTrx, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction),
								dbflex.Eqs("CompanyID", testCoID1,
									"CashBank._id", "BRI01",
									"SourceType", ficomodel.SubledgerCashBank,
									"SourceJournalID", cj.ID,
								))
							convey.SoMsg("bank trx err is not nil", bankTrx, convey.ShouldNotBeNil)
							convey.SoMsg("bank trx amount OK", bankTrx.Amount, convey.ShouldEqual, amt)

							ledgerTrx, _ := datahub.GetByFilter(db, new(ficomodel.LedgerTransaction),
								dbflex.Eqs("CompanyID", testCoID1,
									"SourceType", ficomodel.SubledgerCashBank,
									"Account._id", "399999",
									"SourceJournalID", cj.ID))
							convey.SoMsg("ledger trx err is not nil", ledgerTrx, convey.ShouldNotBeNil)
							convey.SoMsg("ledger trx amount OK", ledgerTrx.Amount, convey.ShouldEqual, -amt)
						})
					})
				})
			})
		})
	})
}

func insertCashJournal(j *ficomodel.CashJournal) (*ficomodel.CashJournal, error) {
	uriPath := "/v1/fico/cashjournal/insert"
	sr := svc.GetRoute(uriPath)
	if sr == nil {
		return nil, fmt.Errorf("missing: route: %s", uriPath)
	}
	ctx := prepareCtxData(kaos.NewContextFromService(svc, sr), "user01", testCoID1)
	nj := new(ficomodel.CashJournal)
	e := svc.CallTo(uriPath, nj, ctx, j)
	return nj, e
}
