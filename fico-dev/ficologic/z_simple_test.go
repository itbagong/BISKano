package ficologic_test

import (
	"testing"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestSetIfEmpty(t *testing.T) {
	record := new(ficomodel.LedgerJournal)
	record.ID = "Sesuatu"
	tenantcorelogic.SetIfEmpty(&record.CompanyID, "ok")
	if record.CompanyID != "ok" {
		t.Fail()
	}
}

func TestMongo(t *testing.T) {
	convey.Convey("system: insert data with subitem", t, func() {
		ct := new(ficomodel.CashTransaction)
		ct.ID = "CT/Test/01"
		ct.CompanyID = testCoID2
		ct.Amount = 10000
		ct.CashBank = tenantcoremodel.CashBank{
			ID:         "TestBank",
			CurrencyID: "IDR",
		}

		ctx := kaos.NewContextFromService(svc, nil)
		prepareCtxData(ctx, "system-user", testCoID2)
		db := sebar.GetTenantDBFromContext(ctx)
		e := db.Save(ct)

		convey.Convey("validate", func() {
			convey.Convey("insert should has no error", func() {
				convey.So(e, convey.ShouldBeNil)

				convey.Convey("filter by account.Currency", func() {
					record, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction),
						dbflex.Eqs("CompanyID", testCoID2, "CashBank.CurrencyID", "IDR"))
					convey.So(record.Amount, convey.ShouldEqual, ct.Amount)
					convey.So(record.CashBank.ID, convey.ShouldEqual, ct.CashBank.ID)

					convey.Convey("filter by account._id", func() {
						record, _ := datahub.GetByFilter(db, new(ficomodel.CashTransaction),
							dbflex.Eqs("CompanyID", testCoID2, "CashBank._id", "TestBank"))
						convey.So(record.Amount, convey.ShouldEqual, ct.Amount)
						convey.So(record.CashBank.ID, convey.ShouldEqual, ct.CashBank.ID)
					})
				})
			})
		})
	})
}

func TestExtractPattern(t *testing.T) {
	convey.Convey("extract pattern", t, func() {
		p, i, err := ficologic.ExtractPattern("CBRI2023A1001")
		convey.So(err, convey.ShouldBeNil)
		convey.So(p, convey.ShouldEqual, "CBRI2023A%04d")
		convey.So(i, convey.ShouldEqual, 1001)
	})
}
