package scmlogic_test

import (
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestBalance(t *testing.T) {
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "balance-user", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)
	convey.Convey("test balance", t, func() {
		trxs := []*scmmodel.InventTrx{
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem1"}, Qty: 10, Status: scmmodel.ItemPlanned, InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem1"}, Qty: 10, Status: scmmodel.ItemConfirmed, InventDim: scmmodel.InventDimension{WarehouseID: "Whs02"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem2"}, Qty: -10, Status: scmmodel.ItemReserved, InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem3"}, Qty: 10, Status: scmmodel.ItemPlanned, InventDim: scmmodel.InventDimension{WarehouseID: "Whs03"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem4"}, Qty: 10, Status: scmmodel.ItemPlanned, InventDim: scmmodel.InventDimension{WarehouseID: "Whs02"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem5"}, Qty: 10, Status: scmmodel.ItemReserved, InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
		}
		err := datahub.Save(db, ficologic.ToDataModels(trxs)...)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("sync nil", func() {
			calc := scmlogic.NewInventBalanceCalc(db)
			err := calc.MakeSnapshot(testCoID1, nil)
			convey.So(err, convey.ShouldBeNil)

			origBals, err := calc.Get(&scmlogic.InventBalanceCalcOpts{CompanyID: testCoID1})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(origBals), convey.ShouldBeGreaterThan, 0)
			/*
				convey.Println("")
				convey.Print("balances:\n")
				convey.Println(strings.Join(lo.Map(origBals, func(b *scmmodel.ItemBalance, i int) string {
					return fmt.Sprintf("ItemID:%s, Whs:%s, DimID: %s, Qty:%.0f, QtyPlanned:%.0f QtyReserved:%.0f",
						b.ItemID, b.InventDim.WarehouseID, b.InventDim.InventDimID, b.Qty, b.QtyPlanned, b.QtyReserved)
				}), "\n"))
			*/

			convey.Convey("inject new trx and update nil", func() {
				trxs := []*scmmodel.InventTrx{
					{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem1"}, Qty: 5, Status: scmmodel.ItemPlanned, InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
					{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem1"}, Qty: 5, Status: scmmodel.ItemConfirmed, InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"}, TrxDate: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Now().Location())},
				}
				err := datahub.Save(db, ficologic.ToDataModels(trxs)...)
				convey.So(err, convey.ShouldBeNil)

				bals, err := calc.Sync(trxs)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(bals), convey.ShouldEqual, 1)
				convey.So(bals[0].QtyPlanned, convey.ShouldEqual, 15)
				convey.So(bals[0].Qty, convey.ShouldEqual, 5)
				//convey.Println("bals", codekit.JsonString(bals[0]))

				convey.Convey("add new trx after 5 jan, and run snapshot, it should not be included", func() {
					dt5Jan := time.Date(2023, 1, 5, 0, 0, 0, 0, time.Now().Location())
					dt6Jan := time.Date(2023, 1, 6, 0, 0, 0, 0, time.Now().Location())
					dt8Jan := time.Date(2023, 1, 8, 0, 0, 0, 0, time.Now().Location())
					dt15Jan := time.Date(2023, 1, 15, 0, 0, 0, 0, time.Now().Location())

					trxs := []*scmmodel.InventTrx{
						{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "BalItem1"}, Qty: 7,
							Status:    scmmodel.ItemConfirmed,
							InventDim: scmmodel.InventDimension{WarehouseID: "Whs01"},
							TrxDate:   dt8Jan},
					}
					err := datahub.Save(db, ficologic.ToDataModels(trxs)...)
					convey.So(err, convey.ShouldBeNil)

					err = calc.MakeSnapshot(testCoID1, &dt5Jan)
					convey.So(err, convey.ShouldBeNil)

					bals, err := calc.Get(&scmlogic.InventBalanceCalcOpts{
						CompanyID:   testCoID1,
						ItemID:      []string{"BalItem1"},
						InventDim:   scmmodel.InventDimension{WarehouseID: "Whs01"},
						BalanceDate: &dt6Jan,
					})
					convey.So(err, convey.ShouldBeNil)
					//convey.Println("")
					//convey.Println("bals", codekit.JsonString(bals))
					convey.So(len(bals), convey.ShouldEqual, 1)
					convey.So(bals[0].Qty, convey.ShouldEqual, 5)

					convey.Convey("check balance after snapshot", func() {
						err = calc.MakeSnapshot(testCoID1, nil)
						convey.So(err, convey.ShouldBeNil)

						bals, err = calc.Get(&scmlogic.InventBalanceCalcOpts{
							CompanyID:   testCoID1,
							ItemID:      []string{"BalItem1"},
							InventDim:   scmmodel.InventDimension{WarehouseID: "Whs01"},
							BalanceDate: &dt15Jan,
						})

						convey.So(err, convey.ShouldBeNil)
						//convey.Println("")
						//convey.Println("bals", codekit.JsonString(bals))
						convey.So(len(bals), convey.ShouldEqual, 1)
						convey.So(bals[0].Qty, convey.ShouldEqual, 12)
					})
				})
			})
		})
	})
}
