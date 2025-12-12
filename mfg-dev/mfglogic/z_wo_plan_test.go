package mfglogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfglogic"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestWOPlan(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	Bom1 := "BOM-001"
	// Bom2 := "BOM-004"

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("Insert journal and submit", t, func() {
		sumMats, sumMans, sumOuts := getWorkOrderPlanSummariesData(db, Bom1) // UI get bom data and append when save

		// save Work Order Plan
		saveReq := &mfglogic.WorkOrderPlanSaveReq{
			WorkOrderPlan: mfgmodel.WorkOrderPlan{
				TrxDate:         trxDate,
				WOName:          "Test Work Order",
				WorkDescription: "Test Work Order",
				BOM:             Bom1,
				JournalTypeID:   "WorkOrder",
			},
			WorkOrderSummaryMaterial: sumMats,
			WorkOrderSummaryResource: sumMans,
			WorkOrderSummaryOutput:   sumOuts,
		}

		saveReq, err := callAPI(ctx, "/v1/mfg/workorderplan/save", saveReq, saveReq)
		convey.So(err, convey.ShouldBeNil)

		wop := saveReq.WorkOrderPlan
		convey.SoMsg("Work Order Plan id = *", wop.ID, convey.ShouldNotBeEmpty)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), wop.CompanyID, convey.ShouldEqual, testCoID1)

		// TODO: test Bom changes, old material lines will be removed

		convey.Convey("Submit", func() {
			_, err = callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
				JournalType: tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderPlan),
				JournalID:   wop.ID,
				Op:          ficologic.PostOpSubmit,
			}}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldBeNil)

			submitted, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderPlan), wop.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			convey.Convey("Approve -> Direct Post", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
					JournalType: tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderPlan),
					JournalID:   wop.ID,
					Op:          ficologic.PostOpApprove,
				}}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)

				posted, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderPlan), wop.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

				convey.Convey("Report Save & Submit", func() {
					prepareCtxData(ctx, "random-user", testCoID1)
					saveReq := &mfglogic.WorkOrderPlanReportSaveReq{
						WorkOrderPlanReport: mfgmodel.WorkOrderPlanReport{
							WorkOrderPlanID: wop.ID,
						},
						WorkOrderPlanReportConsumptionLines: []mfgmodel.WorkOrderMaterialItem{
							{
								ItemID:      "SHOCKABSORBER",
								SKU:         "6553097dbae8381cd665b8d4",
								Description: "",
								UnitID:      "EAC",
								Qty:         5,
								UnitCost:    400000,
								Total:       5 * 400000,
							},
						},
						WorkOrderPlanReportResourceLines: []mfgmodel.WorkOrderResourceItem{
							{
								ExpenseType:  "EXPENSE_XXX",
								ActivityName: "Nyuci Bis",
								Employee:     "EMPLOYEEID_001",
								WorkingHour:  5,
							},
						},
						WorkOrderPlanReportOutputLines: []mfgmodel.WorkOrderOutputItem{
							{
								Type:        mfgmodel.WorkOrderOutputTypeWasteItem,
								SKU:         "6552f890bae8381cd665b48b",
								Description: "",
								GroupID:     "CAT_GROUP_ID",
								UnitID:      "Each",
								Qty:         5,
							},
						},
					}

					saveReq, err = callAPI(ctx, "/v1/mfg/workorderplan/report/save", saveReq, saveReq)
					convey.So(err, convey.ShouldBeNil)
					convey.So(saveReq.ID, convey.ShouldNotBeEmpty)
					convey.So(saveReq.WorkOrderPlanReportConsumptionID, convey.ShouldNotBeEmpty)
					convey.So(saveReq.WorkOrderPlanReportResourceID, convey.ShouldNotBeEmpty)
					convey.So(saveReq.WorkOrderPlanReportOutputID, convey.ShouldNotBeEmpty)

					var empty interface{}
					_, err = callAPI(ctx, "/v1/mfg/workorderplan/report/submit", &mfglogic.WorkOrderPlanReportSubmitReq{ID: saveReq.ID}, &empty)
					convey.So(err, convey.ShouldBeNil)

					submittedCons, err := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportConsumption), saveReq.WorkOrderPlanReportConsumptionID)
					convey.So(err, convey.ShouldBeNil)
					convey.So(submittedCons.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)
					submittedRes, err := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportResource), saveReq.WorkOrderPlanReportResourceID)
					convey.So(err, convey.ShouldBeNil)
					convey.So(submittedRes.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)
					submittedOut, err := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportOutput), saveReq.WorkOrderPlanReportOutputID)
					convey.So(err, convey.ShouldBeNil)
					convey.So(submittedOut.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

					convey.Convey("Report - Posting Profile - All Reports Approve", func() {
						prepareCtxData(ctx, "finance", testCoID1)
						_, err = callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
							JournalType: tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportConsumption),
							JournalID:   saveReq.WorkOrderPlanReportConsumptionID,
							Op:          ficologic.PostOpApprove,
						}}, &[]*tenantcoremodel.PreviewReport{})
						convey.Println("finance | /v1/mfg/postingprofile/post |", mfgmodel.JournalWorkOrderReportConsumption, "| err:", err)
						convey.So(err, convey.ShouldBeNil)

						_, err = callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
							JournalType: tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportResource),
							JournalID:   saveReq.WorkOrderPlanReportResourceID,
							Op:          ficologic.PostOpApprove,
						}}, &[]*tenantcoremodel.PreviewReport{})
						convey.So(err, convey.ShouldBeNil)

						_, err = callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
							JournalType: tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderReportOutput),
							JournalID:   saveReq.WorkOrderPlanReportOutputID,
							Op:          ficologic.PostOpApprove,
						}}, &[]*tenantcoremodel.PreviewReport{})
						convey.So(err, convey.ShouldBeNil)

						postedCons, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportConsumption), saveReq.WorkOrderPlanReportConsumptionID)
						convey.So(postedCons.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)
						postedRes, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportResource), saveReq.WorkOrderPlanReportResourceID)
						convey.So(postedRes.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)
						postedOut, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderPlanReportOutput), saveReq.WorkOrderPlanReportOutputID)
						convey.So(postedOut.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

						// TODO: test after each Report is posted
					})
				})
			})
		})
	})
}

func getWorkOrderPlanSummariesData(db *datahub.Hub, bomID string) ([]mfgmodel.WorkOrderSummaryMaterial, []mfgmodel.WorkOrderSummaryResource, []mfgmodel.WorkOrderSummaryOutput) {
	bom, err := datahub.GetByID(db, new(mfgmodel.BoM), bomID)
	convey.So(err, convey.ShouldBeNil)

	bomMats := []mfgmodel.BoMMaterial{}
	err = db.GetsByFilter(new(mfgmodel.BoMMaterial), dbflex.Eq("BoMID", bomID), &bomMats)
	convey.So(err, convey.ShouldBeNil)

	bomMans := []mfgmodel.BoMManpower{}
	err = db.GetsByFilter(new(mfgmodel.BoMManpower), dbflex.Eq("BoMID", bomID), &bomMans)
	convey.So(err, convey.ShouldBeNil)

	sumMats := lo.Map(bomMats, func(d mfgmodel.BoMMaterial, i int) mfgmodel.WorkOrderSummaryMaterial {
		return mfgmodel.WorkOrderSummaryMaterial{
			ItemID:   d.ItemID,
			SKU:      d.SKU,
			UnitID:   d.UoM,
			Required: float64(d.Qty),
		}
	})

	sumMans := lo.Map(bomMans, func(d mfgmodel.BoMManpower, i int) mfgmodel.WorkOrderSummaryResource {
		return mfgmodel.WorkOrderSummaryResource{
			ExpenseType: "", // TODO: ga tau dari mana, di Bom Manpower ga ada Expense Type nya
			TargetHour:  float64(d.StandartHour),
			RatePerHour: d.RatePerHour,
		}
	})

	unit, _ := datahub.GetByID(db, new(tenantcoremodel.Item), bom.ItemID)

	output := mfgmodel.WorkOrderSummaryOutput{
		Type:                 lo.Ternary(bom.OutputType == mfgmodel.BomOutputTypeItem, mfgmodel.WorkOrderOutputTypeWOOutput, mfgmodel.WorkOrderOutputTypeWasteLedger),
		InventoryLedgerAccID: lo.Ternary(bom.OutputType == mfgmodel.BomOutputTypeItem, bom.ItemID, bom.LedgerID),
		SKU:                  bom.SKU,
		Description:          bom.Description,
		Group:                bom.BoMGroup,
		QtyAmount:            1,
		UnitID:               unit.ID,
	}

	return sumMats, sumMans, []mfgmodel.WorkOrderSummaryOutput{output}
}

func getSummaryMaterials(db *datahub.Hub, workOrderPlanID string) []mfgmodel.WorkOrderSummaryMaterial {
	sumMats := []mfgmodel.WorkOrderSummaryMaterial{}
	err := db.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", workOrderPlanID), &sumMats)
	convey.So(err, convey.ShouldBeNil)
	return sumMats
}

func getBoMMaterials(db *datahub.Hub, bomID string) []mfgmodel.BoMMaterial {
	bomMats := []mfgmodel.BoMMaterial{}
	err := db.GetsByFilter(new(mfgmodel.BoMMaterial), dbflex.Eq("BoMID", bomID), &bomMats)
	convey.So(err, convey.ShouldBeNil)
	return bomMats
}
