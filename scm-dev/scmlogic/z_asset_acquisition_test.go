package scmlogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestAssetAcquisition(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "HOPO"

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("Insert journal and submit", t, func() {
		j := &scmmodel.AssetAcquisitionJournal{
			TransferName:  "Test purchase request",
			TrxDate:       trxDate,
			CompanyID:     testCoID1,
			JournalTypeID: "AssetAcquisition",
			TransferFrom: scmmodel.InventDimension{
				WarehouseID: whsID,
				SectionID:   "SCM",
			},
			ItemTranfers: []scmmodel.AssetItemTransfer{
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: 1000, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 2, ItemID: "Baut", UnitID: "Each", Qty: 500, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 3, ItemID: "Mur", UnitID: "Each", Qty: 1000, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
			},
		}
		j, err := callAPI(ctx, "/v1/scm/asset-acquisition/insert", j, j)

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("Asset Acquisition Journal id = *", j.ID, convey.ShouldNotBeEmpty)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		// posting profile
		_, err = callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
			JournalType: tenantcoremodel.TrxModule(scmmodel.AssetAcquisitionTrxType),
			JournalID:   j.ID,
			Op:          ficologic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(scmmodel.AssetAcquisitionJournal), j.ID)
		convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

		convey.Convey("approve", func() {
			_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
				JournalType: tenantcoremodel.TrxModule(scmmodel.AssetAcquisitionTrxType),
				JournalID:   j.ID,
				Op:          ficologic.PostOpApprove,
			}}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldNotBeNil)

			approved, _ := datahub.GetByID(db, new(scmmodel.AssetAcquisitionJournal), j.ID)
			convey.So(approved.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, j.CompanyID, tenantcoremodel.TrxModule(scmmodel.AssetAcquisitionTrxType).String(), j.ID, false)
			convey.Println("\napproval:", codekit.JsonString(pa.Approvers))

			convey.Convey("approve -> post", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
					JournalType: tenantcoremodel.TrxModule(scmmodel.AssetAcquisitionTrxType),
					JournalID:   j.ID,
					Op:          ficologic.PostOpApprove,
				}}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)

				posted, _ := datahub.GetByID(db, new(scmmodel.AssetAcquisitionJournal), j.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

				convey.Convey("validate qty and cost", func() {
					balCalc := scmlogic.NewInventBalanceCalc(db)
					bal, err := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
						CompanyID: testCoID1,
						ItemID:    []string{"Busi"},
						// BalanceDate: &fullQtyRequest.TrxDate,
					})
					convey.So(err, convey.ShouldBeNil)
					convey.So(len(bal), convey.ShouldEqual, 1)
					convey.So(bal[0].Qty, convey.ShouldEqual, -1000)
					convey.So(bal[0].QtyAvail, convey.ShouldEqual, -1000)
				})
			})
		})
	})
}
