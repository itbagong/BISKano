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

func TestVendorJournal(t *testing.T) {
	var (
		err error
	)

	dt08Jan := time.Date(2023, 1, 8, 0, 0, 0, 0, time.Local)
	//dt10Jan := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "finance", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)

	convey.Convey("create vendor journal", t, func() {
		vj := &ficomodel.VendorJournal{
			CompanyID: testCoID1, TrxDate: dt08Jan, VendorID: "V03", JournalTypeID: "DEF",
			Dimension: tenantcoremodel.Dimension{}.Sets("PC", "MINING", "CC", "OPS", "Site", "KJA"),
			Lines: []ficomodel.JournalLine{
				{Account: ficomodel.NewSubAccount(ficomodel.SubledgerExpense, "ATK"), Amount: 75000},
			},
		}
		vj, err = insertJournal(vj, "/v1/fico/vendorjournal/insert", "site-admin-kja", testCoID1)
		convey.SoMsg("vendor journal creation = OK", err, convey.ShouldBeNil)
		convey.SoMsg("vendor journal id is VJ23*", vj.ID, convey.ShouldStartWith, "VJ23")

		convey.Convey("post vendor journal", func() {
			previews, err := new(ficologic.PostingProfileHandler).Post(ctx,
				[]ficologic.PostRequest{{JournalType: ficomodel.SubledgerVendor, JournalID: vj.ID, Op: ficologic.PostOpSubmit}})
			convey.SoMsg("vendor journal submission error = nil", err, convey.ShouldBeNil)
			convey.SoMsg("previews length = 1", len(previews), convey.ShouldEqual, 1)

			voucherNo := previews[0].Header.GetString("VoucherNo")
			vtrx, _ := datahub.GetByFilter(db, new(ficomodel.VendorTransaction), dbflex.Eqs("CompanyID", vj.CompanyID, "VoucherNo", voucherNo))
			vsched, _ := datahub.GetByFilter(db, new(ficomodel.CashSchedule), dbflex.Eqs("CompanyID", vj.CompanyID, "VoucherNo", voucherNo))

			convey.SoMsg("vendor transaction created", vtrx.ID, convey.ShouldNotBeBlank)
			convey.SoMsg("vendor payment schedule is created", vsched.ID, convey.ShouldNotBeBlank)

			convey.Convey("make payment journal with auto apply", func() {

				convey.Convey("validate payment", func() {

				})
			})
		})
	})
}
