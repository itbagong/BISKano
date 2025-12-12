package scmlogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestPr(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "HOPO"

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("Insert journal and submit", t, func() {
		j := &scmmodel.PurchaseRequestJournal{
			Name:          "Test purchase request",
			TrxDate:       trxDate,
			DocumentDate:  &trxDate,
			CompanyID:     testCoID1,
			JournalTypeID: "PurchaseRequest",
			Location: scmmodel.InventDimension{
				WarehouseID: whsID,
				SectionID:   "SCM",
			},
			Lines: []scmmodel.PurchaseJournalLine{
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: 1000, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 2, ItemID: "Baut", UnitID: "Each", Qty: 500, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
				{InventJournalLine: scmmodel.InventJournalLine{LineNo: 3, ItemID: "Mur", UnitID: "Each", Qty: 1000, InventDim: scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"}}},
			},
		}
		j, err := callAPI(ctx, "/v1/scm/purchase/request/insert", j, j)

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("Purchase request id = *", j.ID, convey.ShouldNotBeEmpty)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		//posting purchase request
		_, err = callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
			JournalType: tenantcoremodel.TrxModule(scmmodel.PurchRequest),
			JournalID:   j.ID,
			Op:          ficologic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(scmmodel.PurchaseRequestJournal), j.ID)
		convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

		convey.Convey("approve", func() {
			_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
				JournalType: tenantcoremodel.TrxModule(scmmodel.PurchRequest),
				JournalID:   j.ID,
				Op:          ficologic.PostOpApprove,
			}}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldNotBeNil)

			approved, _ := datahub.GetByID(db, new(scmmodel.PurchaseRequestJournal), j.ID)
			convey.So(approved.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, j.CompanyID, tenantcoremodel.TrxModule(scmmodel.PurchRequest).String(), j.ID, false)
			convey.Println("\napproval:", codekit.JsonString(pa.Approvers))

			convey.Convey("post", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
					JournalType: tenantcoremodel.TrxModule(scmmodel.PurchRequest),
					JournalID:   j.ID,
					Op:          ficologic.PostOpApprove,
				}}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)

				posted, _ := datahub.GetByID(db, new(scmmodel.PurchaseRequestJournal), j.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)
			})
		})
	})
}
