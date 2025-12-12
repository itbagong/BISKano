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

func TestMovementIn(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)
	trxDate := time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)
	whsID := "HOTestMovementIn"

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("insert journal and submit", t, func() {
		j, err := insertJournal(ctx, &scmmodel.InventJournal{
			Dimension:     tenantcoremodel.Dimension{}.Set("CC", "SCM"),
			InventDim:     scmmodel.InventDimension{WarehouseID: whsID, SectionID: "SCM"},
			JournalTypeID: "MovementIn",
			TrxType:       scmmodel.JournalMovementIn,
			TrxDate:       trxDate,
			Lines: []scmmodel.InventJournalLine{
				{LineNo: 1, ItemID: "Busi", UnitID: "Each", Qty: 1000, Text: "opening balance"},
				{LineNo: 2, ItemID: "Baut", UnitID: "Each", Qty: 500, Text: "opening balance"},
				{LineNo: 3, ItemID: "Mur", UnitID: "Each", Qty: 1000, Text: "opening balance"},
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
					ItemID:    []string{"Busi"},
					InventDim: scmmodel.InventDimension{
						WarehouseID: whsID,
					},
				})
				convey.So(len(bals), convey.ShouldBeGreaterThan, 0)
				convey.So(bals[0].QtyPlanned, convey.ShouldEqual, 1000)

				convey.Convey("update good receive partial", func() {
					// get planned transactions
					planneds, _ := datahub.FindByFilter(db, new(scmmodel.InventTrx),
						dbflex.Eqs("CompanyID", testCoID1, "Status", scmmodel.ItemPlanned, "InventDim.WarehouseID", whsID))
					receiveLines := lo.MapToSlice(lo.GroupBy(planneds, func(plan *scmmodel.InventTrx) string {
						groupIds := []string{plan.Item.ID}
						groupIds = append(groupIds, string(plan.SourceType), plan.SourceJournalID, fmt.Sprintf("%d", plan.SourceLineNo))
						return strings.Join(groupIds, "|")
					}), func(k string, plans []*scmmodel.InventTrx) scmmodel.InventReceiveIssueLine {
						line := scmmodel.InventReceiveIssueLine{}
						line.Qty = 100
						line.ItemID = plans[0].Item.ID
						line.InventDim = plans[0].InventDim
						line.Text = plans[0].Text
						line.SourceType = plans[0].SourceType
						line.SourceJournalID = plans[0].SourceJournalID
						line.SourceTrxType = plans[0].SourceTrxType
						line.SourceLineNo = plans[0].SourceLineNo
						return line
					})

					inventReceiveRequest := &scmmodel.InventReceiveIssueJournal{
						WarehouseID:      whsID,
						TrxDate:          trxDate,
						TrxType:          scmmodel.InventReceive,
						Dimension:        tenantcoremodel.Dimension{},
						Lines:            receiveLines,
						PostingProfileID: "NADP",
					}
					inventReceiveRequest, err = callAPI(ctx, "/v1/scm/inventreceiveissuejournal/insert", inventReceiveRequest, inventReceiveRequest)
					convey.So(err, convey.ShouldBeNil)
					convey.So(inventReceiveRequest.CompanyID, convey.ShouldEqual, testCoID1)
					convey.So(inventReceiveRequest.ID, convey.ShouldNotBeBlank)

					_, err := callAPI(ctx, "/v1/scm/new/postingprofile/post", []ficologic.PostRequest{{
						JournalType: tenantcoremodel.TrxModule(scmmodel.InventReceive),
						JournalID:   inventReceiveRequest.ID,
						Op:          ficologic.PostOpSubmit,
					}}, &[]*tenantcoremodel.PreviewReport{})
					convey.So(err, convey.ShouldBeNil)

					balCalc := scmlogic.NewInventBalanceCalc(db)
					plannedBusi, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
						CompanyID: testCoID1,
						ItemID:    []string{"Busi"},
						InventDim: scmmodel.InventDimension{WarehouseID: whsID},
					})
					convey.So(len(plannedBusi), convey.ShouldBeGreaterThan, 0)
					convey.So(plannedBusi[0].QtyPlanned, convey.ShouldEqual, 900)
					convey.So(plannedBusi[0].Qty, convey.ShouldEqual, 100)

					convey.Convey("update good receive full - over qty - negative test", func() {
						overQtyLine := []scmmodel.InventReceiveIssueLine{
							receiveLines[0],
						}
						overQtyLine[0].Qty = 1700
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
							JournalType: tenantcoremodel.TrxModule(scmmodel.InventReceive),
							JournalID:   overQtyRequest.ID,
							Op:          ficologic.PostOpSubmit,
						}}, &[]*tenantcoremodel.PreviewReport{})
						convey.So(err, convey.ShouldNotBeNil)

						convey.Convey("update good receive full", func() {
							fullQty := []scmmodel.InventReceiveIssueLine{
								receiveLines[0],
							}
							fullQty[0].Qty = 900
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
								JournalType: tenantcoremodel.TrxModule(scmmodel.InventReceive),
								JournalID:   fullQtyRequest.ID,
								Op:          ficologic.PostOpSubmit,
							}}, &[]*tenantcoremodel.PreviewReport{})
							convey.So(err, convey.ShouldBeNil)

							convey.Convey("validate qty and cost", func() {
								fullBal, _ := balCalc.Get(&scmlogic.InventBalanceCalcOpts{
									CompanyID:   testCoID1,
									ItemID:      []string{"Busi"},
									BalanceDate: &fullQtyRequest.TrxDate,
								})
								convey.So(len(fullBal), convey.ShouldEqual, 1)
								convey.So(fullBal[0].Qty, convey.ShouldEqual, 1000)
								convey.So(fullBal[0].QtyPlanned, convey.ShouldEqual, 0)
							})
						})
					})
				})
			})
		})
	})
}

func TestConvertUnit(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)

	prepareCtxData(ctx, "random-user", testCoID1)
	convey.Convey("unit conversion", t, func() {
		q, err := scmlogic.ConvertUnit(db, 20, "PCS", "KODI")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 1)

		q, err = scmlogic.ConvertUnit(db, 1, "KODI", "PCS")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 20)

		q, err = scmlogic.ConvertUnit(db, 40, "PCS", "KODI")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 2)

		q, err = scmlogic.ConvertUnit(db, 2, "KODI", "PCS")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 40)

		q, err = scmlogic.ConvertUnit(db, 100, "PCS", "KODI")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 5)

		q, err = scmlogic.ConvertUnit(db, 5, "KODI", "PCS")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, 100)

		q, err = scmlogic.ConvertUnit(db, -1000, "Each", "Each")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, -1000)

		q, err = scmlogic.ConvertUnit(db, -2, "BOX12", "PCS")
		convey.So(err, convey.ShouldBeNil)
		convey.So(q, convey.ShouldEqual, -24)

		q, err = scmlogic.ConvertUnit(db, 5, "BOX12", "EAC")
		convey.So(err, convey.ShouldBeNil)
		fmt.Printf("\n5 BOX12 = 60 EAC | 5 BOX12 = %v EAC\n", q)
		convey.So(q, convey.ShouldEqual, 60)

	})
}
