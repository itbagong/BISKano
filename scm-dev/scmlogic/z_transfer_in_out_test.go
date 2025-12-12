package scmlogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestTransferInOut(t *testing.T) {
	//prepare data
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("insert jurnal", t, func() {
		j, err := insertJournal(ctx, &scmmodel.InventJournal{
			Dimension: tenantcoremodel.Dimension{}.Set("CC", "SCM"),
			InventDim: scmmodel.InventDimension{
				WarehouseID: "HO",
				SectionID:   "SCM",
			},
			InventDimTo: scmmodel.InventDimension{
				WarehouseID: "HO",
				SectionID:   "KJA",
			},
			JournalTypeID: "Transfer",
			TrxType:       scmmodel.JournalTransfer,
			TrxDate:       trxDate,
			Lines: []scmmodel.InventJournalLine{
				{LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: 1000, Text: "opening balance"},
				{LineNo: 2, ItemID: "Baut", UnitID: "Each", Qty: 500, Text: "opening balance"},
				{LineNo: 3, ItemID: "Mur", UnitID: "Each", Qty: 1000, Text: "opening balance"},
			},
		})

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("journal id = IVJ*", j.ID, convey.ShouldStartWith, "IVJ")
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		convey.Convey("submit - positif test", func() {
			_, err = callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
				{
					JournalType: tenantcoremodel.TrxModule(scmmodel.JournalTransfer),
					JournalID:   j.ID,
					Op:          ficologic.PostOpSubmit,
				},
			}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldBeNil)

			submitted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			convey.So(submitted.TrxType, convey.ShouldEqual, scmmodel.JournalTransfer)

			convey.Convey("approve - positif test", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
					{JournalType: tenantcoremodel.TrxModule(scmmodel.JournalTransfer), JournalID: j.ID, Op: ficologic.PostOpApprove},
				}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)
				posted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

				// get item transactions
				inventTrxs := []scmmodel.InventTrx{}
				err = db.GetsByFilter(new(scmmodel.InventTrx), dbflex.Eq("SourceJournalID", posted.ID), &inventTrxs)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(inventTrxs), convey.ShouldEqual, 6)

				inventTrxIssue, err := datahub.GetByFilter(db, new(scmmodel.InventTrx), dbflex.And(
					dbflex.Eq("SourceJournalID", posted.ID),
					dbflex.Eq("SourceTrxType", scmmodel.InventIssuance),
				))
				convey.So(err, convey.ShouldBeNil)
				convey.SoMsg(fmt.Sprintf("inventTrxIssue.Qty: %v", inventTrxIssue.Qty), inventTrxIssue.Qty, convey.ShouldBeLessThan, 0)
				convey.So(inventTrxIssue.ID, convey.ShouldNotEqual, "")
				convey.So(inventTrxIssue.InventDim.WarehouseID, convey.ShouldEqual, "HO")
				convey.So(inventTrxIssue.InventDim.SectionID, convey.ShouldEqual, "SCM")

				inventTrxReceive, err := datahub.GetByFilter(db, new(scmmodel.InventTrx), dbflex.And(
					dbflex.Eq("SourceJournalID", posted.ID),
					dbflex.Eq("SourceTrxType", scmmodel.InventReceive),
				))
				convey.So(err, convey.ShouldBeNil)
				convey.So(inventTrxReceive.Qty, convey.ShouldBeGreaterThan, 0)
				convey.So(inventTrxReceive.ID, convey.ShouldNotEqual, "")
				convey.So(inventTrxReceive.InventDim.WarehouseID, convey.ShouldEqual, "HO")
				convey.So(inventTrxReceive.InventDim.SectionID, convey.ShouldEqual, "KJA")

				balanceCalc := scmlogic.NewInventBalanceCalc(db)
				bals, _ := balanceCalc.Get(&scmlogic.InventBalanceCalcOpts{
					CompanyID: testCoID1,
					ItemID:    []string{"Busi"},
				})
				convey.So(len(bals), convey.ShouldBeGreaterThan, 0)

				// check qty barang yang akan keluar
				balsOut := lo.Filter(bals, func(item *scmmodel.ItemBalance, i int) bool {
					return item.InventDim.WarehouseID == "HO" && item.InventDim.SectionID == "SCM"
				})
				convey.So(len(balsOut), convey.ShouldBeGreaterThan, 0)
				convey.So(balsOut[0].QtyReserved, convey.ShouldEqual, -1000)

				// check qty barang yang akan masuk
				balsIn := lo.Filter(bals, func(item *scmmodel.ItemBalance, i int) bool {
					return item.InventDim.WarehouseID == "HO" && item.InventDim.SectionID == "KJA"
				})
				convey.So(len(balsIn), convey.ShouldBeGreaterThan, 0)
				convey.So(balsIn[0].QtyPlanned, convey.ShouldEqual, 1000)
			})
		})
	})
	//create invent journal

}
