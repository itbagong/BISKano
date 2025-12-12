package ficologic_test

import (
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestBankRecon(t *testing.T) {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "user-01", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	dt31Dec := time.Date(2022, 12, 31, 0, 0, 0, 0, time.Local)
	dt05Jan := time.Date(2023, 1, 5, 0, 0, 0, 0, time.Local)
	dt15Jan := time.Date(2023, 1, 15, 0, 0, 0, 0, time.Local)
	dt31Jan := time.Date(2023, 1, 31, 0, 0, 0, 0, time.Local)

	convey.Convey("prepare bank and trx to be reconciled", t, func() {
		bank := tenantcoremodel.CashBank{
			ID:                 "BRECON01",
			Name:               "Bank Recon 01",
			CashBankGroupID:    "DEF",
			CurrencyID:         "IDR",
			MainBalanceAccount: "2200001",
		}
		db.Save(&bank)

		bal := ficomodel.CashBalance{
			CompanyID:   testCoID1,
			CashBookID:  "BRECON01",
			BalanceDate: &dt31Dec,
			Balance:     1000000,
		}
		db.Save(&bal)

		trxs := []*ficomodel.CashTransaction{
			{CompanyID: testCoID1, CashBank: bank, Status: ficomodel.AmountConfirmed, TrxDate: dt05Jan, VoucherNo: "V01", Amount: -100000},
			{CompanyID: testCoID1, CashBank: bank, Status: ficomodel.AmountConfirmed, TrxDate: dt05Jan, VoucherNo: "V02", Amount: -100000},
			{CompanyID: testCoID1, CashBank: bank, Status: ficomodel.AmountConfirmed, TrxDate: dt15Jan, VoucherNo: "V03", Amount: 50000},
		}
		for _, trx := range trxs {
			db.Save(trx)
		}
		balEng := ficologic.NewCashBalanceHub(db)
		balEng.Sync(nil, ficologic.CashBalanceOpt{
			CompanyID:  testCoID1,
			AccountIDs: []string{"BRECON01"},
		})
		bals, _ := balEng.Get(nil, ficologic.CashBalanceOpt{CompanyID: testCoID1, AccountIDs: []string{"BRECON01"}})
		convey.So(len(bals), convey.ShouldEqual, 1)
		convey.So(bals[0].Balance, convey.ShouldEqual, 850000)

		convey.Convey("initiate bank recon", func() {
			br := &ficomodel.CashRecon{
				CompanyID:         testCoID1,
				CashBankID:        "BRECON01",
				ReconDate:         dt31Jan,
				PreviousReconDate: &dt31Dec,
				PreviousBalance:   1000000,
				ReconBalance:      880000,
			}
			db.Save(br)

			eng := new(ficologic.CashReconLogic)
			br, err := eng.StartRecon(ctx, br)
			convey.So(err, convey.ShouldBeNil)
			convey.SoMsg("Recon diff = 120000", br.Diff, convey.ShouldEqual, br.ReconBalance-br.PreviousBalance)
			trxs, _ := eng.GetTransactions(ctx, br)
			convey.SoMsg("len trxs = 3", len(trxs), convey.ShouldEqual, 3)

			convey.Convey("reconcile trxs", func() {
				trxIDs := lo.Map(trxs, func(t *ficomodel.CashTransaction, index int) string {
					return t.ID
				})
				req := &ficologic.CashReconcileRequest{
					CashReconID:        br.ID,
					CashTransactionIDs: trxIDs,
				}
				br, err := eng.Reconcile(ctx, req)
				convey.So(err, convey.ShouldBeNil)
				convey.SoMsg("Recon diff = 30000", br.Diff, convey.ShouldEqual, 30000)

				// TODO: auto journal for gaps
				/*
					convey.Convey("reconcile adjustment", func() {
						br.Lines = []ficomodel.JournalLine{
							{LineNo: 1,
								Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "220010"),
								Amount:  30000},
						}

						req := &ficologic.CashReconcileRequest{
							CashReconID:        br.ID,
							CashTransactionIDs: trxIDs,
						}
						br, err := eng.Reconcile(ctx, req)
						convey.So(err, convey.ShouldBeNil)
						convey.SoMsg("Recon diff = 0", br.Diff, convey.ShouldEqual, 0)
					})
				*/
			})
		})
	})
}
