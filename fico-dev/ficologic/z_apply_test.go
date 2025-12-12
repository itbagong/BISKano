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
	"github.com/smartystreets/goconvey/convey"
)

func TestApplyManual(t *testing.T) {
	dt08Jan := time.Date(2023, 1, 8, 0, 0, 0, 0, time.Local)
	dt10Jan := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "finance", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)

	convey.Convey("prepare expense and cash as well", t, func() {
		expenseJournal := &ficomodel.LedgerJournal{
			CompanyID:     testCoID1,
			JournalTypeID: "DEF", TrxDate: dt08Jan,
			Dimension:     tenantcoremodel.Dimension{}.Sets("PC", "MINING", "CC", "OPS"),
			DefaultOffset: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999"),
			Lines: []ficomodel.JournalLine{
				{Account: ficomodel.NewSubAccount(ficomodel.SubledgerExpense, "ATK"), Amount: 35000, Text: "Beli ATK di warung depan"},
			},
		}
		db.Save(expenseJournal)
		cjID := expenseJournal.ID
		_, err := new(ficologic.PostingProfileHandler).Post(ctx,
			[]ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting, JournalID: cjID, Op: ficologic.PostOpSubmit}})
		convey.So(err, convey.ShouldBeNil)

		payment := &ficomodel.CashJournal{
			CompanyID: testCoID1, CashBookID: "BRI01", TrxDate: dt10Jan, JournalTypeID: "DEF",
			Dimension: tenantcoremodel.Dimension{}.Sets("PC", "MINING", "CC", "OPS"),
			Lines:     []ficomodel.JournalLine{{Amount: -20000, Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999")}},
		}
		db.Save(payment)
		pjID := payment.ID
		_, err = new(ficologic.PostingProfileHandler).Post(ctx,
			[]ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: pjID, Op: ficologic.PostOpSubmit}})
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("validate cash schedule", func() {
			paymentScheds := ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerCashBank, payment.ID, "", 0)
			expenseScheds := ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerAccounting, expenseJournal.ID, "", 0)

			convey.SoMsg("expense schedule len = 1", len(expenseScheds), convey.ShouldEqual, 1)
			convey.SoMsg("payment schedule len = 1", len(paymentScheds), convey.ShouldEqual, 1)

			convey.Convey("apply manually without deduction", func() {
				err := ficologic.ApplyCashShedule(ficologic.CashApplyRequest{
					Db:          db,
					TrxDate:     dt10Jan,
					SourceRecID: paymentScheds[0].ID,
					Applies: []ficologic.CashApplyTo{
						{ApplyToRecID: expenseScheds[0].ID, Amount: paymentScheds[0].Amount},
					},
				})
				convey.So(err, convey.ShouldBeNil)

				convey.Convey("validate", func() {
					paymentScheds = ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerCashBank, payment.ID, "", 0)
					expenseScheds = ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerAccounting, expenseJournal.ID, "", 0)

					convey.SoMsg("payment settled in full", paymentScheds[0].Settled, convey.ShouldEqual, paymentScheds[0].Amount)
					convey.SoMsg("payment outstanding is 0", paymentScheds[0].Outstanding, convey.ShouldEqual, 0)
					convey.SoMsg("expense settled = payment amount", expenseScheds[0].Settled, convey.ShouldEqual, paymentScheds[0].Amount)
					convey.SoMsg("expense outstanding = -15000", expenseScheds[0].Outstanding, convey.ShouldEqual, -15000)

					convey.Convey("create new payment and apply with deduction", func() {
						payment := &ficomodel.CashJournal{
							CompanyID: testCoID1, CashBookID: "BRI01", TrxDate: dt10Jan, JournalTypeID: "DEF",
							Dimension: tenantcoremodel.Dimension{}.Sets("PC", "MINING", "CC", "OPS"),
							Lines:     []ficomodel.JournalLine{{Amount: -14500, Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999")}},
						}
						db.Save(payment)
						pjID := payment.ID
						_, err = new(ficologic.PostingProfileHandler).Post(ctx,
							[]ficologic.PostRequest{{JournalType: ficomodel.SubledgerCashBank, JournalID: pjID, Op: ficologic.PostOpSubmit}})
						convey.So(err, convey.ShouldBeNil)

						paymentScheds = ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerCashBank, payment.ID, "", 0)
						err = ficologic.ApplyCashShedule(ficologic.CashApplyRequest{
							Db:          db,
							SourceRecID: paymentScheds[0].ID,
							TrxDate:     dt10Jan,
							Applies: []ficologic.CashApplyTo{
								{ApplyToRecID: expenseScheds[0].ID, Amount: paymentScheds[0].Amount,
									Adjustment: []ficologic.CashApplyAdjustment{
										{Amount: -500,
											To: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "899999")}},
								}},
						})
						convey.So(err, convey.ShouldBeNil)
						expenseScheds = ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerAccounting, expenseJournal.ID, "", 0)
						convey.So(expenseScheds[0].Outstanding, convey.ShouldEqual, -500)

						settlementJournal, err := datahub.GetByFilter(db, new(ficomodel.LedgerJournal),
							dbflex.ElemMatch("References", dbflex.Eqs("Key", "SourceRecID", "Value", paymentScheds[0].ID)))
						convey.So(err, convey.ShouldBeNil)

						previews, err := new(ficologic.PostingProfileHandler).Post(ctx,
							[]ficologic.PostRequest{{JournalType: ficomodel.SubledgerAccounting,
								JournalID: settlementJournal.ID,
								Op:        ficologic.PostOpSubmit}})
						convey.SoMsg("posting settlement journal", err, convey.ShouldBeNil)
						convey.SoMsg("previews exist", len(previews), convey.ShouldEqual, 1)

						expenseScheds = ficologic.FindCashSchedule(db, payment.CompanyID, ficomodel.SubledgerAccounting, expenseJournal.ID, "", 0)
						convey.SoMsg("expense already settled fully", expenseScheds[0].Outstanding, convey.ShouldEqual, 0)

						voucherNo := ficologic.GetVoucherNo(previews[0])
						voucherLegderTrxs, err := datahub.FindByFilter(db, new(ficomodel.LedgerTransaction), dbflex.Eqs("CompanyID", testCoID1, "VoucherNo", voucherNo))
						convey.So(err, convey.ShouldBeNil)
						convey.So(len(voucherLegderTrxs), convey.ShouldBeGreaterThan, 0)
					})
				})
			})
		})
	})
}
