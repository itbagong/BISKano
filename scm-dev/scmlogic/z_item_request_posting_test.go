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

func TestItemRequest(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	// db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "ItemRequestWHS"

	prepareCtxData(ctx, "random-user", testCoID1)

	convey.Convey("insert item request and submit", t, func() {
		j, err := insertItemRequest(ctx, &scmmodel.ItemRequest{
			CompanyID:     testCoID1,
			Name:          "item request test",
			TrxDate:       trxDate,
			TrxType:       scmmodel.ItemRequestType,
			Dimension:     tenantcoremodel.Dimension{}.Set("CC", "SCM"),
			InventDimTo:   scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"},
			JournalTypeID: "ItemRequest",
		})

		err = InsertModel(db, []*scmmodel.ItemRequestDetail{
			{ItemRequestID: j.ID, ItemID: "Busi"},
			{ItemRequestID: j.ID, ItemID: "Baut"},
			{ItemRequestID: j.ID, ItemID: "Mur"},
		})

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		_, err = callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
			JournalType: tenantcoremodel.TrxModule(scmmodel.ItemRequestType),
			JournalID:   j.ID,
			Op:          ficologic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(scmmodel.ItemRequest), j.ID)
		convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

		convey.Convey("approve - negative test", func() {
			_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
				{JournalType: tenantcoremodel.TrxModule(scmmodel.ItemRequestType), JournalID: j.ID, Op: ficologic.PostOpApprove},
			}, &[]*tenantcoremodel.PreviewReport{})

			convey.So(err, convey.ShouldNotBeNil)
			submitted, _ := datahub.GetByID(db, new(scmmodel.ItemRequest), j.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, j.CompanyID, string(scmmodel.ItemRequestType), j.ID, false)
			convey.Println()
			convey.Println("approval:", codekit.JsonString(pa.Approvers))

			convey.Convey("approve & post - positive test", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
					{JournalType: tenantcoremodel.TrxModule(scmmodel.ItemRequestType), JournalID: j.ID, Op: ficologic.PostOpApprove},
				}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)
				posted, _ := datahub.GetByID(db, new(scmmodel.ItemRequest), j.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)
			})
		})
	})
}
