package scmlogic_test

import (
	"fmt"
	"strings"
	"testing"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestSplitInventTrx(t *testing.T) {
	convey.Convey("create invent trx", t, func() {
		ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "randomUser", testCoID1)
		db := sebar.GetTenantDBFromContext(ctx)

		inventTrxSources := []*scmmodel.InventTrx{
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "RandomItem"}, Qty: 10, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J1", SourceLineNo: 1, Text: "Sesuatu 1"},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "RandomItem"}, Qty: 10, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J1", SourceLineNo: 2, Text: "Sesuatu 2"},
			{CompanyID: testCoID1, Item: tenantcoremodel.Item{ID: "RandomItem"}, Qty: 10, Status: scmmodel.ItemPlanned, SourceType: scmmodel.ModuleInventory, SourceJournalID: "J1", SourceLineNo: 3, Text: "Sesuatu 3"},
		}

		err := InsertModel(db, inventTrxSources)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("split by item", func() {
			splitter := scmlogic.NewInventSplit(db)
			splitSources, splitTargets, err := splitter.SetOpts(&scmlogic.InventSplitOpts{
				SplitType:    scmlogic.SplitByItem,
				CompanyID:    testCoID1,
				ItemID:       "RandomItem",
				SourceStatus: string(scmmodel.ItemPlanned),
			}).Split(12, string(scmmodel.ItemConfirmed))
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(splitSources), convey.ShouldEqual, 1)
			convey.SoMsg("qty splitted shld be 12", lo.SumBy(splitTargets, func(s *scmmodel.InventTrx) float64 {
				return s.Qty
			}), convey.ShouldEqual, 12)
			convey.SoMsg("split trx len is 2", len(splitTargets), convey.ShouldEqual, 2)

			origRemains, _ := datahub.Find(db, new(scmmodel.InventTrx),
				dbflex.NewQueryParam().SetWhere(dbflex.Eqs("Item._id", "RandomItem", "Status", scmmodel.ItemPlanned)).SetSort("_id"))
			convey.SoMsg("remaining source len = 2", len(origRemains), convey.ShouldEqual, 2)
			convey.SoMsg("remaining qty = 8", origRemains[0].Qty, convey.ShouldEqual, 8)
			convey.SoMsg("remaining status = planned", origRemains[0].Status, convey.ShouldEqual, scmmodel.ItemPlanned)

			convey.Println("")
			convey.Println("remains:", fmt.Sprintf("%s|%s|%.2f", origRemains[0].ID, origRemains[0].Status, origRemains[0].Qty))
			convey.Println("splitted:", strings.Join(lo.Map(splitTargets, func(t *scmmodel.InventTrx, index int) string {
				return fmt.Sprintf("%s|%s|%.2f", t.ID, t.Status, t.Qty)
			}), ", "))

			convey.Convey("split by source", func() {
				splitSources, splitTargets, err = splitter.SetOpts(&scmlogic.InventSplitOpts{
					SplitType:       scmlogic.SplitBySource,
					CompanyID:       testCoID1,
					SourceType:      scmmodel.ModuleInventory.String(),
					SourceJournalID: "J1",
					SourceLineNo:    origRemains[0].SourceLineNo,
					SourceStatus:    string(scmmodel.ItemPlanned),
				}).Split(5, string(scmmodel.ItemConfirmed))
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(splitSources), convey.ShouldEqual, 1)

				origRemains, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.Eqs(
					"SourceType", scmmodel.ModuleInventory,
					"SourceJournalID", "J1",
					"SourceLineNo", origRemains[0].SourceLineNo,
					"Status", scmmodel.ItemPlanned))
				convey.SoMsg("remaining source len = 1", len(origRemains), convey.ShouldEqual, 1)
				convey.SoMsg("remaining qty = 3", origRemains[0].Qty, convey.ShouldEqual, 3)
				convey.SoMsg("remaining status = planned", origRemains[0].Status, convey.ShouldEqual, scmmodel.ItemPlanned)

				convey.Println("")
				convey.Println("remains:", fmt.Sprintf("%s|%s|%.2f|%s",
					origRemains[0].ID, origRemains[0].Status, origRemains[0].Qty, origRemains[0].InventDim.InventDimID))
				convey.Println("splitted:", strings.Join(lo.Map(splitTargets, func(t *scmmodel.InventTrx, index int) string {
					return fmt.Sprintf("%s|%s|%.2f|%s", t.ID, t.Status, t.Qty, t.InventDim.InventDimID)
				}), ", "))

				convey.Convey("clear", func() {
					err := db.DeleteByFilter(new(scmmodel.InventTrx), dbflex.Eq("Item._id", "RandomItem"))
					convey.So(err, convey.ShouldBeNil)
				})
			})
		})
	})
}
