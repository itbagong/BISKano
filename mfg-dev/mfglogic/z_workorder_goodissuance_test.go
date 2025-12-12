package mfglogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestWo(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)

	prepareCtxData(ctx, "random-user", testCoID1)
	injectItemBalance(db, WHSID)
	convey.Convey("Insert WO Journal and submit", t, func() {
		wo, err := insertWorkOrder(ctx, &mfgmodel.WorkOrderJournal{
			Name:          "Test work order",
			TrxDate:       &trxDate,
			CompanyID:     testCoID1,
			JournalTypeID: "WorkOrder",
			InventDim: scmmodel.InventDimension{
				WarehouseID: WHSID,
				SectionID:   "MFG",
			},
			ItemUsage: []scmmodel.InventReceiveIssueLine{
				scmmodel.InventReceiveIssueLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: -1000, InventDim: scmmodel.InventDimension{
							WarehouseID: WHSID, SectionID: "SCM",
						},
					},
					Item: tenantcoremodel.Item{
						ID: "Busi",
					},
				},
				scmmodel.InventReceiveIssueLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo: 2, ItemID: "Baut", UnitID: "Each", Qty: -500, InventDim: scmmodel.InventDimension{
							WarehouseID: WHSID, SectionID: "SCM",
						},
					},
					Item: tenantcoremodel.Item{
						ID: "Baut",
					},
				},
				scmmodel.InventReceiveIssueLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo: 3, ItemID: "Mur", UnitID: "Each", Qty: -1000, InventDim: scmmodel.InventDimension{
							WarehouseID: WHSID, SectionID: "SCM",
						},
					},
					Item: tenantcoremodel.Item{
						ID: "Mur",
					},
				},
			},
		})

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("Work order id = *", wo.ID, convey.ShouldNotBeEmpty)
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), wo.CompanyID, convey.ShouldEqual, testCoID1)

		_, err = callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
			JournalType: scmmodel.ModuleWorkorder,
			JournalID:   wo.ID,
			Op:          ficologic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderJournal), wo.ID)
		convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

		convey.Convey("approve -negative test", func() {
			_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
				JournalType: scmmodel.ModuleWorkorder,
				JournalID:   wo.ID,
				Op:          ficologic.PostOpApprove,
			}}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldNotBeNil)
			submitted, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderJournal), wo.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, wo.CompanyID, scmmodel.ModuleWorkorder.String(), wo.ID, false)
			convey.So(pa, convey.ShouldNotBeNil)
			convey.Println("approval:", codekit.JsonString(pa.Approvers))
			convey.Convey("approve & post - positive test", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{
					{JournalType: scmmodel.ModuleWorkorder, JournalID: wo.ID, Op: ficologic.PostOpApprove},
				}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)
				posted, _ := datahub.GetByID(db, new(mfgmodel.WorkOrderJournal), wo.ID)

				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

				balanceCalc := scmlogic.NewInventBalanceCalc(db)
				bals, _ := balanceCalc.Get(&scmlogic.InventBalanceCalcOpts{
					CompanyID: testCoID1,
					ItemID:    []string{"Busi"},
					InventDim: scmmodel.InventDimension{
						WarehouseID: WHSID, SectionID: "SCM",
					},
				})
				convey.So(len(bals), convey.ShouldEqual, 1)
				convey.So(bals[0].QtyPlanned, convey.ShouldEqual, 0)
				convey.So(bals[0].QtyReserved, convey.ShouldEqual, -1000)
				convey.So(bals[0].Qty, convey.ShouldEqual, 1000)
				bals[0].Calc()
				convey.So(bals[0].QtyAvail, convey.ShouldEqual, 0)
				// convey.Convey("update good receive partial", func() {
				// 	//get planned transactions
				// 	reserveds, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.Eqs("CompanyID", testCoID1, "Status", scmmodel.ItemReserved, "InventDim.WarehouseID", WHSID))

				// 	reservedLines := lo.MapToSlice(lo.GroupBy(reserveds, func(plan *scmmodel.InventTrx) string {
				// 		groupIds := []string{plan.Item.ID}
				// 		groupIds = append(groupIds, string(plan.SourceType), plan.SourceJournalID, fmt.Sprintf("%d", plan.SourceLineNo))
				// 		return strings.Join(groupIds, "|")
				// 	}), func(k string, plans []*scmmodel.InventTrx) scmmodel.InventReceiveIssueLine {
				// 		line := scmmodel.InventReceiveIssueLine{}
				// 		line.Qty = -100
				// 		line.ItemID = plans[0].Item.ID
				// 		line.Item.ID = plans[0].Item.ID
				// 		line.InventDim = plans[0].InventDim
				// 		line.Text = plans[0].Text
				// 		line.SourceType = plans[0].SourceType
				// 		line.SourceJournalID = plans[0].SourceJournalID
				// 		line.SourceTrxType = plans[0].SourceTrxType
				// 		line.SourceLineNo = plans[0].SourceLineNo

				// 		return line
				// 	})

				// 	inventReceiveRequest := &scmmodel.InventReceiveIssueJournal{
				// 		WarehouseID:      WHSID,
				// 		TrxDate:          trxDate,
				// 		TrxType:          scmmodel.InventReceive,
				// 		Dimension:        tenantcoremodel.Dimension{},
				// 		Lines:            reservedLines,
				// 		PostingProfileID: "NADP",
				// 	}
				// 	inventReceiveRequest, err = callAPI(ctx, "/v1/mfg/inventreceiveissuejournal/insert", inventReceiveRequest, inventReceiveRequest)
				// 	convey.So(err, convey.ShouldBeNil)
				// 	convey.So(inventReceiveRequest.CompanyID, convey.ShouldEqual, testCoID1)
				// 	convey.So(inventReceiveRequest.ID, convey.ShouldNotBeBlank)

				// 	_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
				// 		JournalType: tenantcoremodel.TrxModule(scmmodel.InventIssuance),
				// 		JournalID:   inventReceiveRequest.ID,
				// 		Op:          ficologic.PostOpSubmit,
				// 	}}, &[]*tenantcoremodel.PreviewReport{})
				// 	convey.So(err, convey.ShouldBeNil)

				// 	inventDim := &scmmodel.InventDimension{
				// 		WarehouseID: WHSID,
				// 		SectionID:   "SCM",
				// 	}
				// 	inventDim.Calc()

				// 	balCalc := scmlogic.NewInventBalanceCalc(db)
				// 	issueBusi, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
				// 		CompanyID: testCoID1,
				// 		ItemID:    []string{"Busi"},
				// 		InventDim: *inventDim,
				// 	})

				// 	convey.So(len(issueBusi), convey.ShouldBeGreaterThan, 0)
				// 	convey.So(issueBusi[0].QtyReserved, convey.ShouldEqual, -900)
				// 	convey.So(issueBusi[0].Qty, convey.ShouldEqual, 900)

				// 	convey.Convey("update good receive full - over qty - negative test", func() {
				// 		overQtyLine := []scmmodel.InventReceiveIssueLine{
				// 			reservedLines[0],
				// 		}
				// 		overQtyLine[0].Qty = -1700
				// 		overQtyRequest := &scmmodel.InventReceiveIssueJournal{
				// 			WarehouseID:      WHSID,
				// 			TrxDate:          trxDate,
				// 			TrxType:          scmmodel.InventIssuance,
				// 			Dimension:        tenantcoremodel.Dimension{},
				// 			Lines:            overQtyLine,
				// 			PostingProfileID: "NADP",
				// 		}
				// 		callAPI(ctx, "/v1/mfg/inventreceiveissuejournal/insert", overQtyRequest, overQtyRequest)
				// 		_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []ficologic.PostRequest{{
				// 			JournalType: tenantcoremodel.TrxModule(scmmodel.InventReceive),
				// 			JournalID:   overQtyRequest.ID,
				// 			Op:          ficologic.PostOpSubmit,
				// 		}}, &[]*tenantcoremodel.PreviewReport{})
				// 		convey.So(err, convey.ShouldNotBeNil)

				// 		convey.Convey("update good receive full", func() {
				// 			fullQty := []scmmodel.InventReceiveIssueLine{
				// 				reservedLines[0],
				// 			}
				// 			fullQty[0].Qty = -900
				// 			fullQtyRequest := &scmmodel.InventReceiveIssueJournal{
				// 				WarehouseID:      WHSID,
				// 				TrxDate:          trxDate,
				// 				TrxType:          scmmodel.InventIssuance,
				// 				Dimension:        tenantcoremodel.Dimension{},
				// 				Lines:            fullQty,
				// 				PostingProfileID: "NADP",
				// 			}
				// 			callAPI(ctx, "/v1/mfg/inventreceiveissuejournal/insert", fullQtyRequest, fullQtyRequest)
				// 			_, err := callAPI(ctx, "/v1/mfg/postingprofile/post", []mfglogic.PostRequest{{
				// 				JournalType: tenantcoremodel.TrxModule(scmmodel.InventIssuance),
				// 				JournalID:   fullQtyRequest.ID,
				// 				Op:          mfglogic.PostOpSubmit,
				// 			}}, &[]*tenantcoremodel.PreviewReport{})
				// 			convey.So(err, convey.ShouldBeNil)

				// 			convey.Convey("validate qty and cost", func() {
				// 				fullBal, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
				// 					CompanyID:   testCoID1,
				// 					ItemID:      []string{"Busi"},
				// 					BalanceDate: &fullQtyRequest.TrxDate,
				// 				})
				// 				convey.So(len(fullBal), convey.ShouldEqual, 1)
				// 				convey.So(fullBal[0].Qty, convey.ShouldEqual, 0)
				// 				convey.So(fullBal[0].QtyReserved, convey.ShouldEqual, 0)
				// 			})
				// 		})
				// 	})
				// })
			})
		})
	})
}

func injectItemBalance(db *datahub.Hub, whs string) {
	InsertModel(db, []*tenantcoremodel.Item{
		{ID: "Busi", DefaultUnitID: "Each", CostUnit: 1},
	})

	dim := &scmmodel.InventDimension{
		WarehouseID: whs,
		SectionID:   "SCM",
	}
	dim.Calc()
	trxDate := time.Date(2023, 1, 9, 0, 0, 0, 0, time.Local)
	balance := scmmodel.ItemBalance{
		ItemID:      "Busi",
		CompanyID:   testCoID1,
		Qty:         1000,
		QtyReserved: 0,
		QtyPlanned:  0,
		InventDim:   *dim,
		BalanceDate: &trxDate,
	}

	db.Save(&balance)
}
