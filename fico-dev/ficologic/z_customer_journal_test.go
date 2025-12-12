package ficologic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestCustomerJournal(t *testing.T) {
	convey.Convey("system: create customer journal", t, func() {
		cj, err := insertCustomerJournal(&ficomodel.CustomerJournal{
			CustomerID:    "C01",
			JournalTypeID: "DEF",
			TrxDate:       time.Date(2023, 1, 15, 0, 0, 0, 0, time.Now().Location()),
			Dimension:     tenantcoremodel.Dimension{}.Set("PC", "OPS").Set("Site", "KJA"),
			TaxCodes:      []string{"PPN"},
			Text:          "Test 01",
			Lines: []ficomodel.JournalLine{
				{LineNo: 1, Amount: 250000, Text: "Development service Jan 2023",
					Taxable: true,
					Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "400001")},
				{LineNo: 2, Amount: 500000, Text: "Claimable Jan 2023",
					Taxable: true,
					Account: ficomodel.NewSubAccount(ficomodel.SubledgerAccounting, "400001")},
			},
		})

		convey.Convey("validate: check error and journal no", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(cj.ID, convey.ShouldStartWith, "CSJ01")

			convey.Convey("random-user: submit", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				db := sebar.GetTenantDBFromContext(ctx)
				prepareCtxData(ctx, "finance", testCoID1)
				pph := new(ficologic.PostingProfileHandler)
				_, err := pph.Post(ctx, []ficologic.PostRequest{
					{JournalType: ficomodel.SubledgerCustomer, JournalID: cj.ID, Op: ficologic.PostOpSubmit},
				})
				convey.So(err, convey.ShouldBeNil)

				convey.Convey("validate submission", func() {
					submitted, _ := datahub.GetByID(db, new(ficomodel.CustomerJournal), cj.ID)
					convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

					convey.Convey("validate balance", func() {
						balCust := ficologic.NewCustomerBalanceHub(db)
						bals, err := balCust.Get(nil, ficologic.CustomerBalanceOpt{
							CompanyID:  cj.CompanyID,
							AccountIDs: []string{cj.CustomerID}})
						convey.So(err, convey.ShouldBeNil)
						convey.So(len(bals), convey.ShouldBeGreaterThan, 0)
						convey.So(bals[0].Balance, convey.ShouldEqual, 750000)
					})
				})
			})
		})
	})
}

func insertCustomerJournal(j *ficomodel.CustomerJournal) (*ficomodel.CustomerJournal, error) {
	uriPath := "/v1/fico/customerjournal/insert"
	sr := svc.GetRoute(uriPath)
	if sr == nil {
		return nil, fmt.Errorf("missing: route: %s", uriPath)
	}
	ctx := prepareCtxData(kaos.NewContextFromService(svc, sr), "user01", testCoID1)
	nj := new(ficomodel.CustomerJournal)
	e := svc.CallTo(uriPath, nj, ctx, j)
	return nj, e
}
