package ficologic_test

import (
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestCheque(t *testing.T) {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "user-01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	dt05Jan := time.Date(2023, 1, 5, 0, 0, 0, 0, time.Local)
	dt15Jan := time.Date(2023, 1, 15, 0, 0, 0, 0, time.Local)

	convey.Convey("prepare cheque", t, func() {
		ledger := &tenantcoremodel.LedgerAccount{
			ID:          "20CB00",
			Name:        "Cash & Bank Cheque",
			AccountType: tenantcoremodel.BalanceSheetAccount,
		}

		cbg := &tenantcoremodel.CashBankGroup{
			ID:                 "CBGCG",
			Name:               "Cash book check group",
			IsActive:           true,
			MainBalanceAccount: "20CB00",
		}

		cb := &tenantcoremodel.CashBank{
			ID:              "CBG01",
			Name:            "Cash Bank 01",
			CashBankGroupID: "CBGCG",
			IsActive:        true,
		}

		db.Save(ledger)
		db.Save(cbg)
		db.Save(cb)

		cgb := ficomodel.ChequeGiroBook{
			CashBookID: "CBG01",
			From:       "C2023X2100",
			Kind:       ficomodel.Cheque,
			Qty:        100,
		}

		addr := "/v1/fico/cg/create-update-book"
		err := svc.CallTo(addr, &cgb, ctx, &cgb)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("validate cheque", func() {
			cgs, _ := datahub.FindByFilter(db, new(ficomodel.ChequeGiro), dbflex.Eq("CheckBookID", cgb.ID))
			convey.SoMsg("cheques len = 100", len(cgs), convey.ShouldEqual, 100)
			convey.SoMsg("cheque 1 No = C2023X2100", cgs[0].ID, convey.ShouldEqual, "C2023X2100")
			convey.SoMsg("cheque 1 No = C2023X2199", cgs[99].ID, convey.ShouldEqual, "C2023X2199")

			convey.Convey("reserve cheque to trx", func() {
				// create journal
				cj := ficomodel.CashJournal{
					CashJournalType: "DEF",
					CashBookID:      "CBG01",
					TrxDate:         dt05Jan,
					Lines: []ficomodel.JournalLine{
						{LineNo: 1,
							Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "599991"),
							Amount:  -100000,
						},
					},
				}

				addr = "/v1/fico/cashjournal/insert"
				err := svc.CallTo(addr, &cj, ctx, &cj)
				convey.So(err, convey.ShouldBeNil)

				/* reserve cheque
				Proses ini perlu dipastikan bahwa cash journal lines sudah terbentuk dgn line no yang tepat
				*/
				chequeNo := "C2023X2105"
				cq := ficomodel.ChequeGiro{
					ID:        chequeNo,
					IssueDate: &dt05Jan,
					ClearDate: &dt15Jan,
					Amount:    cj.Lines[0].Amount,
					Memo:      "Bayar broooo",
					BfcName:   "Si Tukang Tagih",
					// dua data dibawah ini perlu sdh ada di database
					CashJournalID: cj.ID,
					LineNo:        1,
				}
				addr = "/v1/fico/cg/reserve"
				err = svc.CallTo(addr, &cq, ctx, &cq)
				convey.So(err, convey.ShouldBeNil)

				db.GetByID(&cj, cj.ID)
				convey.So(cj.Lines[0].ChequeGiroID, convey.ShouldEqual, chequeNo)

				/*
				 belum di test

				 release sebelum clear date - negative test
				 release >= clear date - positive test
				 unreserve
				*/
			})
		})
	})
}
