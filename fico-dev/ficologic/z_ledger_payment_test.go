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

func TestPayExpense(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	prepareCtxData(ctx, "user01", testCoID1)

	convey.Convey("create a ledger expense", t, func() {
		lj := &ficomodel.LedgerJournal{
			CompanyID:        testCoID1,
			TrxDate:          time.Date(2023, 1, 10, 0, 0, 0, 0, time.Now().Location()),
			JournalTypeID:    "GEN",
			PostingProfileID: "NADP",
			Text:             "Test pay expense",
			DefaultOffset:    ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerAccounting, AccountID: "899999"},
			Dimension: tenantcoremodel.Dimension{
				{Key: "PC", Value: "PC-NONE"},
				{Key: "CC", Value: "OPS"},
				{Key: "Site", Value: "KJA"},
			},
			Lines: []ficomodel.JournalLine{
				{LineNo: 1, Amount: -350000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: "Las"}},
				{LineNo: 2, Amount: -500000, Account: ficomodel.SubledgerAccount{AccountType: ficomodel.SubledgerExpense, AccountID: "General"}}},
		}
		lj, e := insertLedgerJournal(lj)
		convey.So(e, convey.ShouldBeNil)
		convey.So(lj.ID, convey.ShouldNotEqual, "")
		convey.Printf("ledger expense %s is created", lj.ID)
		convey.Convey("post the expense", func() {
			prepareCtxData(ctx, "finance", testCoID1)
			pph := new(ficologic.PostingProfileHandler)
			_, err := pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: lj.ID, Op: ficologic.PostOpSubmit}})

			// validate
			convey.SoMsg("post error should be nil", err, convey.ShouldBeNil)
			vouchers := ficologic.FindLegderTrxBySource(db, ficomodel.SubledgerAccounting, lj.ID, 0, true)
			convey.SoMsg("voucher should be created", len(vouchers), convey.ShouldBeGreaterThan, 0)
			convey.Printf("voucher: %s", vouchers[0].VoucherNo)

			convey.Convey("validate cash schedule", func() {

				// validate
				schedule, err := datahub.GetByFilter(db, new(ficomodel.CashSchedule), dbflex.Eqs("SourceType", ficomodel.SubledgerAccounting, "SourceJournalID", lj.ID, "Account.AccountType", ficomodel.SubledgerExpense, "Account.AccountID", "Las"))
				convey.So(err, convey.ShouldBeNil)
				convey.SoMsg("schedule amount should be OK", schedule.Amount, convey.ShouldEqual, -350000)

				convey.Convey("post a cash out and apply to schedule", func() {
					cj, err := insertCashJournal(&ficomodel.CashJournal{
						CompanyID:        testCoID1,
						Dimension:        tenantcoremodel.Dimension{}.Set("PC", "NONE").Set("CC", "OPS").Set("Site", "HQ").Set("AssetID", "NONE"),
						JournalTypeID:    "CO-OPS",
						PostingProfileID: "NADP",
						CashBookID:       "BRI01",
						TrxDate:          lj.TrxDate.AddDate(0, 0, 2),
						Lines: []ficomodel.JournalLine{
							{LineNo: 1, Amount: -350000, OffsetAccount: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999")},
						},
					})
					convey.So(err, convey.ShouldBeNil)

					prepareCtxData(ctx, "finance", testCoID1)
					pph := new(ficologic.PostingProfileHandler)
					_, err = pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpSubmit}})
					convey.SoMsg("cash posting error should be nil", err, convey.ShouldBeNil)

					cashTrx, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction), dbflex.Eqs("CompanyID", cj.CompanyID, "SourceType", ficomodel.SubledgerCashBank, "SourceJournalID", cj.ID))
					convey.SoMsg("cash transaction is created", cashTrx.Amount, convey.ShouldEqual, -350000)

					scheduleCashOut, _ := datahub.GetByFilter(db, new(ficomodel.CashSchedule), dbflex.Eqs("CompanyID", cj.CompanyID, "SourceType", ficomodel.SubledgerCashBank, "SourceJournalID", cj.ID, "Account.AccountType", ficomodel.SubledgerCashBank, "Account.AccountID", cj.CashBookID))
					convey.SoMsg("schedule cashout amount should be OK", scheduleCashOut.Amount, convey.ShouldEqual, -350000)
					convey.SoMsg("schedule cashout type should be Payment", scheduleCashOut.Direction, convey.ShouldEqual, ficomodel.CashExpense)

					_, err = ficologic.SettleCashSchedule(db, scheduleCashOut, schedule)
					convey.SoMsg("apply schedule error should be nil", err, convey.ShouldBeNil)

					// validate
					schedules, _ := datahub.FindByFilter(db, new(ficomodel.CashSchedule), dbflex.Eq("_id", schedule.ID))
					for _, schedule := range schedules {
						if schedule.Amount == schedule.Settled && schedule.Outstanding == 0 && schedule.Status == ficomodel.CashSettled {
							continue
						}

						convey.SoMsg(fmt.Sprintf("schedule for %s/%s is not settled properly", schedule.Account.AccountType, schedule.Account.AccountID), false, convey.ShouldBeTrue)
					}

					convey.Convey("post 2nd cash out with same apply, should fail", func() {
						cj, err := insertCashJournal(&ficomodel.CashJournal{
							CompanyID:        testCoID1,
							Dimension:        tenantcoremodel.Dimension{}.Set("PC", "NONE").Set("CC", "OPS").Set("Site", "HQ").Set("AssetID", "NONE"),
							JournalTypeID:    "CO-OPS",
							PostingProfileID: "NADP",
							CashBookID:       "BRI01",
							TrxDate:          lj.TrxDate.AddDate(0, 0, 2),
							Lines: []ficomodel.JournalLine{
								{LineNo: 1, Amount: 350000, OffsetAccount: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999")},
							},
						})
						convey.So(err, convey.ShouldBeNil)

						prepareCtxData(ctx, "finance", testCoID1)
						pph := new(ficologic.PostingProfileHandler)
						pph.Post(ctx, []ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: cj.ID, Op: ficologic.PostOpSubmit}})

						scheduleCashOut, _ := datahub.GetByFilter(db, new(ficomodel.CashSchedule), dbflex.Eqs("CompanyID", cj.CompanyID, "SourceType", ficomodel.SubledgerCashBank, "SourceJournalID", cj.ID, "Account.AccountType", ficomodel.SubledgerCashBank, "Account.AccountID", cj.CashBookID))

						cas, _ := ficologic.SettleCashSchedule(db, scheduleCashOut, schedule)
						convey.SoMsg("the should be no apply", len(cas), convey.ShouldEqual, 0)
					})
				})
			})
		})
	})
}
