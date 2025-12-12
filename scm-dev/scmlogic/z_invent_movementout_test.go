package scmlogic_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestMovementOut(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "HOTestMovementOut"
	test := scmmodel.InventDimension{
		WarehouseID: whsID,
		SectionID:   "SCM",
	}

	test.Calc()

	prepareCtxData(ctx, "random-user", testCoID1)
	injectItemMovementOut(db, whsID)

	convey.Convey("insert journal and submit", t, func() {
		j, err := insertJournal(ctx, &scmmodel.InventJournal{
			Dimension:     tenantcoremodel.Dimension{}.Set("CC", "SCM"),
			InventDim:     scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"},
			JournalTypeID: "MovementOut",
			TrxType:       scmmodel.JournalMovementOut,
			TrxDate:       trxDate,
			Lines: []scmmodel.InventJournalLine{
				{LineNo: 1, ItemID: "ItemOut1", UnitID: "Each", Qty: -10000, Text: "opening balance", InventDim: test},
			},
		})

		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("journal id = IVJ*", j.ID, convey.ShouldStartWith, "IVJ")
		convey.SoMsg(fmt.Sprintf("companyid = %s", testCoID1), j.CompanyID, convey.ShouldEqual, testCoID1)

		_, err = callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
			JournalType: scmmodel.ModuleInventory,
			JournalID:   j.ID,
			Op:          ficologic.PostOpSubmit,
		}}, &[]*tenantcoremodel.PreviewReport{})
		convey.So(err, convey.ShouldBeNil)

		submitted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
		convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

		convey.Convey("approve - negative test", func() {
			_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
				{JournalType: scmmodel.ModuleInventory, JournalID: j.ID, Op: ficologic.PostOpApprove},
			}, &[]*tenantcoremodel.PreviewReport{})
			convey.So(err, convey.ShouldNotBeNil)
			submitted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
			convey.So(submitted.Status, convey.ShouldEqual, ficomodel.JournalStatusSubmitted)

			pa, _ := ficologic.GetPostingApprovalBySource(db, j.CompanyID, scmmodel.ModuleInventory.String(), j.ID, false)
			convey.Println()
			convey.Println("approval:", codekit.JsonString(pa.Approvers))

			convey.Convey("approve & post - positive test", func() {
				prepareCtxData(ctx, "finance", testCoID1)
				_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{
					{JournalType: scmmodel.ModuleInventory, JournalID: j.ID, Op: ficologic.PostOpApprove},
				}, &[]*tenantcoremodel.PreviewReport{})
				convey.So(err, convey.ShouldBeNil)
				posted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
				convey.So(posted.Status, convey.ShouldEqual, ficomodel.JournalStatusPosted)

				balanceCalc := scmlogic.NewInventBalanceCalc(db)
				bals, _ := balanceCalc.Get(&scmlogic.InventBalanceCalcOpts{
					CompanyID: testCoID1,
					ItemID:    []string{"ItemOut1"},
					InventDim: test,
				})

				convey.So(len(bals), convey.ShouldBeGreaterThan, 0)
				convey.So(bals[0].QtyReserved, convey.ShouldEqual, -1000)
				convey.So(bals[0].Qty, convey.ShouldEqual, 1000)
				bals[0].Calc()
				convey.So(bals[0].QtyAvail, convey.ShouldEqual, 0)

				convey.Convey("update good issue partial", func() {
					issues, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx), dbflex.Eqs("CompanyID", testCoID1, "Status", scmmodel.ItemReserved, "InventDim.WarehouseID", whsID))
					issueLines := lo.MapToSlice(lo.GroupBy(issues, func(plan *scmmodel.InventTrx) string {
						groupIds := []string{plan.Item.ID}
						groupIds = append(groupIds, string(plan.SourceType), plan.SourceJournalID, fmt.Sprintf("%d", plan.SourceLineNo))
						return strings.Join(groupIds, "|")
					}), func(k string, plans []*scmmodel.InventTrx) scmmodel.InventReceiveIssueLine {
						line := scmmodel.InventReceiveIssueLine{}
						line.Qty = -100
						line.ItemID = plans[0].Item.ID
						line.InventDim = plans[0].InventDim
						line.Text = plans[0].Text
						line.SourceType = plans[0].SourceType
						line.SourceJournalID = plans[0].SourceJournalID
						line.SourceTrxType = plans[0].SourceTrxType
						line.SourceLineNo = plans[0].SourceLineNo

						return line
					})

					inventIssueRequest := &scmmodel.InventReceiveIssueJournal{
						WarehouseID:      whsID,
						TrxDate:          trxDate,
						TrxType:          scmmodel.InventIssuance,
						Dimension:        tenantcoremodel.Dimension{},
						Lines:            issueLines,
						PostingProfileID: "NADP",
					}

					inventIssueRequest, err = callAPI(ctx, "/v1/scm/inventreceiveissuejournal/insert", inventIssueRequest, inventIssueRequest)
					convey.So(err, convey.ShouldBeNil)
					convey.So(inventIssueRequest.CompanyID, convey.ShouldEqual, testCoID1)
					convey.So(inventIssueRequest.ID, convey.ShouldNotBeBlank)

					_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
						JournalType: tenantcoremodel.TrxModule(scmmodel.InventIssuance),
						JournalID:   inventIssueRequest.ID,
						Op:          ficologic.PostOpSubmit,
					}}, &[]*tenantcoremodel.PreviewReport{})
					convey.So(err, convey.ShouldBeNil)

					balCalc := scmlogic.NewInventBalanceCalc(db)
					issueItemOut1, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
						CompanyID: testCoID1,
						ItemID:    []string{"ItemOut1"},
						InventDim: test,
					})
					convey.So(len(issueItemOut1), convey.ShouldBeGreaterThan, 0)
					convey.So(issueItemOut1[0].QtyReserved, convey.ShouldEqual, -900)
					convey.So(issueItemOut1[0].Qty, convey.ShouldEqual, 900)

					convey.Convey("update good receive full - over qty - negative test", func() {
						overQtyLine := []scmmodel.InventReceiveIssueLine{
							issueLines[0],
						}

						overQtyLine[0].Qty = -1700
						overQtyRequest := &scmmodel.InventReceiveIssueJournal{
							WarehouseID:      whsID,
							TrxDate:          trxDate,
							TrxType:          scmmodel.InventReceive,
							Dimension:        tenantcoremodel.Dimension{},
							Lines:            overQtyLine,
							PostingProfileID: "NADP",
						}

						callAPI(ctx, "/v1/scm/inventreceiveissuejournal/insert", overQtyRequest, overQtyRequest)
						_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
							JournalType: tenantcoremodel.TrxModule(scmmodel.InventIssuance),
							JournalID:   overQtyRequest.ID,
							Op:          ficologic.PostOpSubmit,
						}}, &[]*tenantcoremodel.PreviewReport{})
						convey.So(err, convey.ShouldNotBeNil)

						convey.Convey("update good receive full", func() {
							fullQty := []scmmodel.InventReceiveIssueLine{
								issueLines[0],
							}

							fullQty[0].Qty = -900
							fullQtyRequest := &scmmodel.InventReceiveIssueJournal{
								WarehouseID:      whsID,
								TrxDate:          trxDate,
								TrxType:          scmmodel.InventReceive,
								Dimension:        tenantcoremodel.Dimension{},
								Lines:            fullQty,
								PostingProfileID: "NADP",
							}
							callAPI(ctx, "/v1/scm/inventreceiveissuejournal/insert", fullQtyRequest, fullQtyRequest)
							_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
								JournalType: tenantcoremodel.TrxModule(scmmodel.InventIssuance),
								JournalID:   fullQtyRequest.ID,
								Op:          ficologic.PostOpSubmit,
							}}, &[]*tenantcoremodel.PreviewReport{})
							convey.So(err, convey.ShouldBeNil)

							convey.Convey("validate qty and cost", func() {
								fullBal, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
									CompanyID:   testCoID1,
									ItemID:      []string{"ItemOut1"},
									BalanceDate: &fullQtyRequest.TrxDate,
								})
								convey.So(len(fullBal), convey.ShouldEqual, 1)
								convey.So(fullBal[0].Qty, convey.ShouldEqual, 0)
								convey.So(fullBal[0].QtyReserved, convey.ShouldEqual, 0)
							})
						})
					})
				})
			})
		})
	})
}

func injectItemMovementOut(db *datahub.Hub, whs string) {
	InsertModel(db, []*tenantcoremodel.Item{
		{ID: "ItemOut1", DefaultUnitID: "Each", CostUnit: 1},
	})

	dim := &scmmodel.InventDimension{
		WarehouseID: whs,
		SectionID:   "SCM",
	}
	dim.Calc()
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	balance := scmmodel.ItemBalance{
		ItemID:      "ItemOut1",
		CompanyID:   testCoID1,
		Qty:         1000,
		QtyReserved: 0,
		QtyPlanned:  0,
		InventDim:   *dim,
		BalanceDate: &trxDate,
	}

	db.Save(&balance)
}
