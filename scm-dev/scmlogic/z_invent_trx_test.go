package scmlogic_test

import (
	"testing"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestInventTrxGetsFilter(t *testing.T) {
	convey.Convey("create invent trx", t, func() {
		ctx := kaos.NewContextFromService(svc, nil)
		db := sebar.GetTenantDBFromContext(ctx)

		inventTrxSources := []*scmmodel.InventTrx{
			{CompanyID: testCoID1, InventDim: scmmodel.InventDimension{WarehouseID: "WH1"}, Item: tenantcoremodel.Item{ID: "JerukGradeA"}, Qty: 1000, TrxQty: 1000, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J1", SourceLineNo: 1, Text: "Sesuatu"},
			{CompanyID: testCoID1, InventDim: scmmodel.InventDimension{WarehouseID: "WH1"}, Item: tenantcoremodel.Item{ID: "JerukGradeA"}, Qty: 2000, TrxQty: 2000, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J1", SourceLineNo: 1, Text: "Sesuatu"},
			{CompanyID: testCoID1, InventDim: scmmodel.InventDimension{WarehouseID: "WH1"}, Item: tenantcoremodel.Item{ID: "AppleGradeA"}, Qty: 5000, TrxQty: 5000, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J2", SourceLineNo: 1, Text: "Sesuatu"},
			{CompanyID: testCoID1, InventDim: scmmodel.InventDimension{WarehouseID: "WH1"}, Item: tenantcoremodel.Item{ID: "JerukGradeA"}, Qty: 1200, TrxQty: 1200, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J2", SourceLineNo: 2, Text: "Sesuatu"},
			{CompanyID: testCoID1, InventDim: scmmodel.InventDimension{WarehouseID: "WH1"}, Item: tenantcoremodel.Item{ID: "JerukGradeA"}, Qty: 800, TrxQty: 800, Status: scmmodel.ItemConfirmed, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J2", SourceLineNo: 2, Text: "Sesuatu"},
		}

		err := InsertModel(db, inventTrxSources)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("get filter", func() {
			payload := &scmlogic.InventTrxGetsFilterRequest{
				CompanyID:   testCoID1,
				WarehouseID: "WH1",
				SourceType:  scmmodel.ModuleInventory,
			}
			res, err := new(scmlogic.InventTrxEngine).GetsFilter(ctx, payload)
			datas, ok := res.Get("data").([]scmmodel.InventTrxReceipt)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(len(datas), convey.ShouldEqual, 3)

			resSettle, exist := lo.Find(datas, func(d scmmodel.InventTrxReceipt) bool {
				return d.SourceType == scmmodel.ModuleInventory && d.SourceJournalID == "J2" && d.SourceLineNo == 2
			})
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(resSettle.SettledQty, convey.ShouldEqual, 800)
			convey.So(resSettle.OriginalQty, convey.ShouldEqual, 2000)
		})
	})
}
