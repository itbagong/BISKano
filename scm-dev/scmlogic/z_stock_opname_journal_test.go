package scmlogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestStockOpname(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "WHOTestStockOpname"

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("insert stock opname journal and submit", t, func() {
		// insert item balance
		ibs := []*scmmodel.ItemBalance{
			{
				ItemID:      "THRUST_WASHER",
				SKU:         "A4003324562",
				CompanyID:   testCoID1,
				InventDim:   scmmodel.InventDimension{WarehouseID: whsID},
				Qty:         50,
				QtyPlanned:  10,
				QtyReserved: 5,
			},
		}
		for _, ib := range ibs {
			ib.Calc()
		}

		err := InsertModel(db, ibs)
		convey.So(err, convey.ShouldBeNil)

		// get lines and transform to
		glpay := &scmmodel.InventDimension{WarehouseID: whsID}
		lines := []scmmodel.StockOpnameDetail{}
		_, err = callAPI(ctx, "/v1/scm/stock-opname/get-lines", glpay, &lines)
		convey.Println("get-lines:", codekit.JsonString(lines))
		convey.So(err, convey.ShouldBeNil)

		soJournalLines := []scmmodel.StockOpnameJournalLine{}
		for _, l := range lines {
			soJournalLines = append(soJournalLines, scmmodel.StockOpnameJournalLine{
				InventJournalLine: scmmodel.InventJournalLine{LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: 1000, Text: "opening balance"},
				Description:       l.Description,
				QtyInSystem:       l.QtyInSystem,
				QtyActual:         "30",
				Gap:               30 - l.QtyInSystem,
			})
		}

		soj := &scmmodel.StockOpnameJournal{
			Name:          "Test Stock Opname",
			CompanyID:     testCoID1,
			Dimension:     tenantcoremodel.Dimension{}.Set("CC", "SCM"),
			InventDim:     scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"},
			JournalTypeID: "StockOpname",
			TrxType:       scmmodel.JournalOpname,
			TrxDate:       trxDate,
			Lines:         soJournalLines,
		}
		j, err := callAPI(ctx, "/v1/scm/stock-opname/insert", soj, soj)
		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg(fmt.Sprintf("journal id = %s", j.ID), j.ID, convey.ShouldNotBeEmpty)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		_, err = callAPI(ctx, "/v1/scm/postingprofile/post", []scmlogic.PostRequest{{
			JournalType: tenantcoremodel.TrxModule(scmmodel.JournalOpname),
			JournalID:   j.ID,
			Op:          scmlogic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(scmmodel.StockOpnameJournal), j.ID)
		convey.So(submitted.Status, convey.ShouldEqual, scmmodel.JournalSubmitted)

		convey.Convey("submit - positive test", func() {
			_, err := callAPI(ctx, "/v1/scm/postingprofile/post", []scmlogic.PostRequest{
				{
					JournalType: tenantcoremodel.TrxModule(scmmodel.JournalOpname),
					JournalID:   j.ID,
					Op:          scmlogic.PostOpApprove,
				},
			}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldNotBeNil)

			submitted, _ := datahub.GetByID(db, new(scmmodel.StockOpnameJournal), j.ID)
			convey.So(submitted.Status, convey.ShouldEqual, scmmodel.JournalSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, j.CompanyID, tenantcoremodel.TrxModule(scmmodel.JournalOpname).String(), j.ID, false)
			convey.Println()
			convey.Println("approval:", codekit.JsonString(pa.Approvers))

			convey.Convey("approve & post - positive test", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/postingprofile/post", []scmlogic.PostRequest{
					{
						JournalType: tenantcoremodel.TrxModule(scmmodel.JournalOpname),
						JournalID:   j.ID,
						Op:          scmlogic.PostOpApprove,
					},
				}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)

				posted, _ := datahub.GetByID(db, new(scmmodel.StockOpnameJournal), j.ID)
				convey.Println("StockOpnameJournal Posted:", codekit.JsonStringIndent(posted, "\t"))
				convey.So(posted.Status, convey.ShouldEqual, scmmodel.JournalPosted)

				balanceCalc := scmlogic.NewInventBalanceCalc(db)
				bals, _ := balanceCalc.Get(&scmlogic.InventBalanceCalcOpts{
					CompanyID: testCoID1,
					ItemID:    []string{"THRUST_WASHER"},
					InventDim: scmmodel.InventDimension{
						WarehouseID: whsID,
					},
				})
				convey.Println("Item Balance:", codekit.JsonStringIndent(bals, "\t"))
				convey.So(len(bals), convey.ShouldBeGreaterThan, 0)
				convey.So(bals[0].QtyPlanned, convey.ShouldEqual, 10) // TODO: apa yang sebaiknya di cek setelah stock opname???
			})
		})
	})
}
