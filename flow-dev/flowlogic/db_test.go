package flowlogic_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/flow/flowlogic"
	"git.kanosolution.net/sebar/flow/flowmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/strikememongo"
	"github.com/ariefdarmawan/strikememongo/strikememongolog"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

var (
	err             error
	dbSvr           *strikememongo.Server
	tenantDBConnStr = "%s/db_%s"
	companyID       = "flowco"
	svc             *kaos.Service
)

func TestMain(m *testing.M) {
	dbSvr, err = strikememongo.StartWithOptions(&strikememongo.Options{
		MongodBin:      "/Applications/mongodb/bin/mongod",
		LogLevel:       strikememongolog.LogLevelSilent,
		StartupTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer dbSvr.Stop()

	hm := kaos.NewHubManager(nil)
	hm.SetHubBuilder(func(key, group string) (*datahub.Hub, error) {
		vTenantConnStr := fmt.Sprintf(tenantDBConnStr, dbSvr.URI(), key)
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(vTenantConnStr, false), true, 100)
		hconn.SetAutoCloseDuration(2 * time.Second)
		return hconn, nil
	})
	defer hm.Close()

	svc = kaos.NewService().SetBasePoint("/v1/flow")
	svc.SetHubManager(hm)
	//svc.Log().SetLevelStdOut(logger.ErrorLevel, false)

	dbm := sebar.NewDBModFromContext()
	svc.RegisterModel(new(flowmodel.FlowTemplate), "template").SetMod(dbm)
	svc.RegisterModel(new(flowmodel.Request), "request").SetMod(dbm).AllowOnlyRoute("get")
	svc.PrepareRoutes("")
	os.Exit(m.Run())
}

func TestInsertFlow(t *testing.T) {
	convey.Convey("insert flow template", t, func() {
		flowSave, err := insertTemplate()

		convey.Convey("insert data", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(time.Since(flowSave.Created), convey.ShouldBeLessThan, 2*time.Second)

			convey.Convey("get data", func() {
				flowGet, err := getTemplate(flowSave.ID)
				convey.So(err, convey.ShouldBeNil)
				convey.So(flowGet.Created.UTC().Unix(), convey.ShouldAlmostEqual, flowSave.Created.UTC().Unix())
			})
		})

		convey.Convey("re-insert", func() {
			_, err := insertTemplate()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestRejectRequest(t *testing.T) {
	convey.Convey("working with request", t, func() {
		var (
			req *flowmodel.Request
			err error
		)

		convey.Convey("create request", func() {
			_, e := getTemplate("T01")
			if e != nil {
				insertTemplate()
			}

			ctx := kaos.NewContextFromService(svc, nil)
			ctx.Data().Set("jwt_reference_id", "user01")
			ctx.Data().Set("CompanyID", companyID)
			reqHandler := new(flowlogic.Request)
			req, err = reqHandler.Create(ctx, &flowlogic.RequestPayload{
				TemplateID: "T01",
				Title:      "Template request 01",
				Start:      true,
				Payload:    codekit.M{"Site": "KJA"},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

			convey.Convey("re-start a request (negative test)", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				reqHandler := new(flowlogic.Request)
				_, err = reqHandler.Start(ctx, req)
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "invalid status")
			})

			convey.Convey("approve using different user, should raise error invalid: user", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				ctx.Data().Set("jwt_reference_id", "fakeuser01")
				ctx.Data().Set("CompanyID", companyID)

				h := sebar.GetTenantDBFromContext(ctx)
				convey.So(h, convey.ShouldNotBeNil)

				task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RE"))
				convey.So(err, convey.ShouldBeNil)

				reqHandler := new(flowlogic.Request)
				_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
					TaskID:   task.ID,
					Approval: false,
					Reason:   "rejection test ae",
				})
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, "invalid: user")
				//convey.Println("\n", "msg", err.Error())
			})

			convey.Convey("reject first step, should close the request as fail", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				ctx.Data().Set("jwt_reference_id", "reviewer01")
				ctx.Data().Set("CompanyID", companyID)

				h := sebar.GetTenantDBFromContext(ctx)
				convey.So(h, convey.ShouldNotBeNil)

				task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RE"))
				convey.So(err, convey.ShouldBeNil)

				reqHandler := new(flowlogic.Request)
				_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
					TaskID:   task.ID,
					Approval: false,
					Reason:   "rejection test ae",
				})
				convey.So(err, convey.ShouldBeNil)

				req, _ := datahub.GetByID(h, new(flowmodel.Request), req.ID)
				convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestFail)
			})
		})
	})
}

func TestApproveRequest(t *testing.T) {
	convey.Convey("working with request", t, func() {
		var (
			req *flowmodel.Request
			err error
		)

		convey.Convey("create request", func() {
			_, e := getTemplate("T01")
			if e != nil {
				insertTemplate()
			}

			ctx := kaos.NewContextFromService(svc, nil)
			ctx.Data().Set("jwt_reference_id", "user01")
			ctx.Data().Set("CompanyID", companyID)
			reqHandler := new(flowlogic.Request)
			req, err = reqHandler.Create(ctx, &flowlogic.RequestPayload{
				TemplateID: "T01",
				Title:      "Template request 01",
				Start:      true,
				Payload:    codekit.M{"Site": "KJA"},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

			convey.Convey("approve first step", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				ctx.Data().Set("jwt_reference_id", "reviewer01")
				ctx.Data().Set("CompanyID", companyID)

				h := sebar.GetTenantDBFromContext(ctx)
				convey.So(h, convey.ShouldNotBeNil)

				task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RE"))
				convey.So(err, convey.ShouldBeNil)

				reqHandler := new(flowlogic.Request)
				_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
					TaskID:   task.ID,
					Approval: true,
				})
				convey.So(err, convey.ShouldBeNil)

				req, _ := datahub.GetByID(h, new(flowmodel.Request), req.ID)
				convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

				convey.Convey("approve second step", func() {
					ctx := kaos.NewContextFromService(svc, nil)
					ctx.Data().Set("jwt_reference_id", "manager01")
					ctx.Data().Set("CompanyID", companyID)

					h := sebar.GetTenantDBFromContext(ctx)
					convey.So(h, convey.ShouldNotBeNil)

					task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RM"))
					convey.So(err, convey.ShouldBeNil)

					reqHandler := new(flowlogic.Request)
					_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
						TaskID:   task.ID,
						Approval: true,
					})
					convey.So(err, convey.ShouldBeNil)

					req, _ := datahub.GetByID(h, new(flowmodel.Request), req.ID)
					convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestSuccess)
				})
			})
		})
	})
}

func TestApproveRequest2(t *testing.T) {
	convey.Convey("working with request with 2 branches", t, func() {
		var (
			req *flowmodel.Request
			err error
		)

		convey.Convey("create request", func() {
			_, e := getTemplate("T02")
			if e != nil {
				insertTemplate2()
			}

			ctx := kaos.NewContextFromService(svc, nil)
			ctx.Data().Set("jwt_reference_id", "user01")
			ctx.Data().Set("CompanyID", companyID)
			reqHandler := new(flowlogic.Request)
			req, err = reqHandler.Create(ctx, &flowlogic.RequestPayload{
				TemplateID: "T02",
				Title:      "Template request 02",
				Start:      true,
				Payload:    codekit.M{"Site": "KJA"},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

			convey.Convey("approve first step", func() {
				ctx := kaos.NewContextFromService(svc, nil)
				ctx.Data().Set("jwt_reference_id", "reviewer01")
				ctx.Data().Set("CompanyID", companyID)

				h := sebar.GetTenantDBFromContext(ctx)
				convey.So(h, convey.ShouldNotBeNil)

				task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RE"))
				convey.So(err, convey.ShouldBeNil)

				reqHandler := new(flowlogic.Request)
				_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
					TaskID:   task.ID,
					Approval: true,
				})
				convey.So(err, convey.ShouldBeNil)

				req, _ := datahub.GetByID(h, new(flowmodel.Request), req.ID)
				convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

				convey.Convey("approve second step", func() {
					ctx := kaos.NewContextFromService(svc, nil)
					ctx.Data().Set("jwt_reference_id", "manager02")
					ctx.Data().Set("CompanyID", companyID)

					h := sebar.GetTenantDBFromContext(ctx)
					convey.So(h, convey.ShouldNotBeNil)

					task, err := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Setup.ID", "RM2"))
					convey.So(err, convey.ShouldBeNil)

					reqHandler := new(flowlogic.Request)
					_, err = reqHandler.Review(ctx, &flowlogic.ReviewPayload{
						TaskID:   task.ID,
						Approval: true,
					})
					convey.So(err, convey.ShouldBeNil)

					req, _ := datahub.GetByID(h, new(flowmodel.Request), req.ID)
					convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestSuccess)

					convey.Convey("check all task shld be success or cancelled", func() {
						reqTasks, err := datahub.FindByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Version", req.Version))
						convey.So(err, convey.ShouldBeNil)
						closedTask := lo.Filter(reqTasks, func(t *flowmodel.RequestTask, i int) bool {
							return t.Status == flowmodel.TaskCancel || t.Status == flowmodel.TaskSuccess
						})
						convey.So(len(closedTask), convey.ShouldEqual, len(req.Template.Tasks))
					})
				})
			})
		})
	})
}

func TestCancelRequest(t *testing.T) {
	convey.Convey("working with request", t, func() {
		var (
			req *flowmodel.Request
			err error
		)

		convey.Convey("create request", func() {
			_, e := getTemplate("T01")
			if e != nil {
				insertTemplate()
			}

			ctx := kaos.NewContextFromService(svc, nil)
			ctx.Data().Set("jwt_reference_id", "user01")
			ctx.Data().Set("CompanyID", companyID)
			reqHandler := new(flowlogic.Request)
			req, err = reqHandler.Create(ctx, &flowlogic.RequestPayload{
				TemplateID: "T01",
				Title:      "Template request 01",
				Start:      true,
				Payload:    codekit.M{"Site": "KJA"},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(req.Status, convey.ShouldEqual, flowmodel.RequestRunning)

			convey.Convey("cancel request", func() {
				_, err := reqHandler.Cancel(ctx, &flowlogic.CancelPayload{
					ID:     req.ID,
					Reason: "test cancel",
				})
				convey.So(err, convey.ShouldBeNil)

				db := sebar.GetTenantDBFromContext(ctx)
				updatedReq, _ := datahub.GetByID(db, new(flowmodel.Request), req.ID)
				convey.So(updatedReq, convey.ShouldNotBeNil)
				convey.So(updatedReq.Status, convey.ShouldEqual, flowmodel.RequestCancel)

				convey.Convey("check all task shld be cancelled", func() {
					reqTasks, err := datahub.FindByFilter(db, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Version", req.Version))
					convey.So(err, convey.ShouldBeNil)
					notCancelledTask := lo.Filter(reqTasks, func(t *flowmodel.RequestTask, i int) bool {
						return t.Status != flowmodel.TaskCancel
					})
					convey.So(len(notCancelledTask), convey.ShouldEqual, 0)
				})
			})
		})
	})
}

func insertTemplate() (*flowmodel.FlowTemplate, error) {
	flowInsertPath := "/v1/flow/template/insert"
	srInsert := svc.GetRoute(flowInsertPath)
	ctx := kaos.NewContextFromService(svc, srInsert)

	flow := &flowmodel.FlowTemplate{
		ID:   "T01",
		Name: "Flow test 01",
		Tasks: []flowmodel.Task{
			{ID: "RE", Name: "Review Entry",
				TaskType:          flowmodel.TaskReview,
				StopRequestIfFail: true,
				Users:             []flowmodel.TaskUser{{UserID: "reviewer01"}}},
			{ID: "RM", Name: "Review by Manager", TaskType: flowmodel.TaskReview,
				StopRequestIfSuccess: true, StopRequestIfFail: true,
				Users: []flowmodel.TaskUser{{UserID: "manager01"}}},
		},
		Routes: []flowmodel.Route{{FromID: "RE", ToID: "RM"}},
	}

	fs, err := svc.Call(flowInsertPath, ctx, flow)
	if err != nil {
		return nil, err
	}
	return fs.(*flowmodel.FlowTemplate), nil
}

func insertTemplate2() (*flowmodel.FlowTemplate, error) {
	flowInsertPath := "/v1/flow/template/insert"
	srInsert := svc.GetRoute(flowInsertPath)
	ctx := kaos.NewContextFromService(svc, srInsert)

	flow := &flowmodel.FlowTemplate{
		ID:   "T02",
		Name: "Flow test 02",
		Tasks: []flowmodel.Task{
			{ID: "RE", Name: "Review Entry",
				TaskType:          flowmodel.TaskReview,
				StopRequestIfFail: true,
				Users:             []flowmodel.TaskUser{{UserID: "reviewer01"}}},
			{ID: "RM1", Name: "Review by Manager 1", TaskType: flowmodel.TaskReview,
				StopRequestIfSuccess: true, StopRequestIfFail: true,
				Users: []flowmodel.TaskUser{{UserID: "manager01"}}},
			{ID: "RM2", Name: "Review by Manager 2", TaskType: flowmodel.TaskReview,
				StopRequestIfSuccess: true, StopRequestIfFail: true,
				Users: []flowmodel.TaskUser{{UserID: "manager02"}}},
		},
		Routes: []flowmodel.Route{{FromID: "RE", ToID: "RM"}, {FromID: "RE", ToID: "RM2"}},
	}

	fs, err := svc.Call(flowInsertPath, ctx, flow)
	if err != nil {
		return nil, err
	}
	return fs.(*flowmodel.FlowTemplate), nil
}

func getTemplate(id string) (*flowmodel.FlowTemplate, error) {
	flowGetData := "/v1/flow/template/get"
	srGet := svc.GetRoute(flowGetData)
	ctx := kaos.NewContextFromService(svc, srGet)

	d, e := svc.Call(flowGetData, ctx, []interface{}{id})
	if e != nil {
		return nil, e
	}
	return d.(*flowmodel.FlowTemplate), nil
}
