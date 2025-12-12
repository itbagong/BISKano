package main

import (
	"flag"
	"os"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/mfg/mfglogic"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoreconfig"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/samber/lo"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/mfg"
	logger      = sebar.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func checkAccessFn(ctx *kaos.Context, parm interface{}, permission string, accessLevel int) error {
	return nil
	/* use this for production: sebar.CheckAccess(ctx, sebar.CheckAccessRequest{
		Param: parm.(codekit.M), PermissionID: permission, AccessLevel: accessLevel
	})
	*/
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	// jwt
	scmconfig.Config.SetEventHub(ev)

	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")
	ev.SetTimeout(1 * time.Minute)

	if e := sebar.ConfigHasData(appConfig, "addr_auth_validation", "addr_access_validation"); e != nil {
		s.Log().Error(e.Error())
		os.Exit(1)
	}

	if err := serde.Serde(appConfig.Data, &tenantcoreconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}

	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.EventHub = ev

	if err := serde.Serde(appConfig.Data, &mfglogic.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	mfglogic.Config.EventHub = ev

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()

	func(models ...*kaos.ServiceModel) {
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"}
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}
		rbacEndPoints := []string{"get", "gets", "find"}
		rbacModelList := []string{
			// journals
			"work/request",
			"workorderplan",
		}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			// for NON-RBAC Models OR RBAC Model with "formconfig", "gridconfig", "listconfig", "new"
			switch modelPtr.Name {
			case "routine":
				publicModels[index].
					RegisterMWs(
						mfglogic.MWRoutine(),
					).
					RegisterPostMWs(
						mfglogic.MWPostRoutine(),
					)
			case "routine/detail":
				publicModels[index].RegisterPostMWs(
					mfglogic.MWPostRoutineDetail(),
				)
			// case "work/request":
			// 	publicModels[index].RegisterPostMWs(
			// 		mfglogic.MWPostWorkRequestGrid(),
			// 	)
			case "work/order", "workorder":
				publicModels[index].RegisterPostMWs(
					mfglogic.MWPostWorkOrderGrid(),
					// mfglogic.MWPostWorkOrder(),
				)
			case "routine/template":
				publicModels[index].RegisterPostMWs(
					mfglogic.MWPostRoutineTemplateGets(),
				)
			case "bom":
				publicModels[index].RegisterPostMWs(
					mfglogic.MWPostBomGets(),
				)
			}

			if lo.Contains(rbacModelList, modelPtr.Name) {
				// for public: "formconfig", "gridconfig", "listconfig", "new"
				newPublicEndPoints, _ := lo.Difference(publicEndPoints, rbacEndPoints)
				publicModels[index].AllowOnlyRoute(newPublicEndPoints...)

				// for limited access public: "get", "gets", "find"
				rbacModel := new(kaos.ServiceModel)
				*rbacModel = *modelPtr
				s.AddServiceModel(rbacModel)
				rbacModel.AllowOnlyRoute(rbacEndPoints...).RegisterMWs(
					// TODO: tambahkan filter companyID
					rbaclogic.MWRbacFilterDim("", "jwt"),
					scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
				)

				switch rbacModel.Name {
				case "work/request":
					rbacModel.RegisterPostMWs(
						mfglogic.MWPostWorkRequestGrid(),
					)
				case "workorderplan":
					rbacModel.RegisterPostMWs(
						mfglogic.MWPostWorkOrderPlanGets(),
						mfglogic.MWPostWorkOrderPlanGet(),
					)
				}

				publicModels = append(publicModels, rbacModel)
			} else {
				publicModels[index].AllowOnlyRoute(publicEndPoints...)
			}
		}

		//-- protected: "insert", "update", "save", "delete", "delete-many"
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		protectedModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			protectedModel := new(kaos.ServiceModel)
			*protectedModel = *modelPtr
			s.AddServiceModel(protectedModel)
			protectedModels[index] = protectedModel

			switch modelPtr.Name {
			case "work/request":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("WorkRequest", true, "_id"),
				)
			case "workorderplan":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("WorkOrder", true, "_id"),
					mfglogic.MWPostWorkOrderPlanSave(),
				)
			case "work/order", "workorder":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("WorkOrder", true, "_id"),
					mfglogic.MWPreAssignCompanyID(),
					mfglogic.MWPreWorkOrderSave(),
				)
			case "workorderplan/report/consumption":
				protectedModels[index].RegisterMWs(
					mfglogic.MWPreAssignCompanyID(),
				)
			case "workorderplan/report/output":
				protectedModels[index].RegisterMWs(
					mfglogic.MWPreAssignCompanyID(),
				)
			case "WorkOrderPlanReportResource":
				protectedModels[index].RegisterMWs(
					mfglogic.MWPreAssignCompanyID(),
				)
			case "bom":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("BOM", true, "_id"),
				)
			}
		}

		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
			// AllowOnlyRoute(publicEndPoints...). // move to public for loop above
			RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			})).
			Apply(publicModels...)

		//-- tenant admin
		s.Group().
			SetMod(modDB).
			SetDeployer(hd.DeployerName).
			RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
				Permission:          "TenantAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute(protectedEndPoints...).
			Apply(protectedModels...)

	}(
		s.RegisterModel(new(mfgmodel.RoutineTemplate), "routine/template"),
		s.RegisterModel(new(mfgmodel.Routine), "routine"),
		s.RegisterModel(new(mfgmodel.RoutineDetail), "routine/detail"),
		s.RegisterModel(new(mfgmodel.RoutineChecklist), "routine/checklist"),
		s.RegisterModel(new(mfgmodel.RoutineChecklistDetail), "routine/checklist/detail"),
		s.RegisterModel(new(mfgmodel.RoutineChecklistAttachment), "routine/checklist/attachment"),
		s.RegisterModel(new(mfgmodel.BoM), "bom"),
		s.RegisterModel(new(mfgmodel.BoMMaterial), "bom/material"),
		s.RegisterModel(new(mfgmodel.BoMManpower), "bom/manpower"),
		s.RegisterModel(new(mfgmodel.BoMMachinery), "bom/machinery"),
		s.RegisterModel(new(mfgmodel.WorkRequest), "work/request").DisableRoute("gridconfig"),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "work/order").DisableRoute("save", "get", "gridconfig").RegisterMWs(mfglogic.MWPreAssignCompanyID()),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "work/order").AllowOnlyRoute("save").RegisterMWs(mfglogic.MWPreWorkOrderSave(), mfglogic.MWPreAssignCompanyID()),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "work/order").AllowOnlyRoute("get").RegisterPostMWs(mfglogic.MWPostWorkOrder()),
		s.RegisterModel(new(mfgmodel.WorkOrder), "work/order").DisableRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderDetail), "work/order/detail"),
		s.RegisterModel(new(mfgmodel.WorkOrderDetailItem), "work/order/detail/item"),
		s.RegisterModel(new(mfgmodel.WorkOrderDetailManpower), "work/order/detail/manpower"),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "workorder").DisableRoute("save", "get").RegisterMWs(mfglogic.MWPreAssignCompanyID()),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "workorder").AllowOnlyRoute("save").RegisterMWs(mfglogic.MWPreWorkOrderSave(), mfglogic.MWPreAssignCompanyID()),
		// s.RegisterModel(new(mfgmodel.WorkOrder), "workorder").AllowOnlyRoute("get").RegisterPostMWs(mfglogic.MWPostWorkOrder()),
		s.RegisterModel(new(mfgmodel.WorkOrder), "workorder"),
		s.RegisterModel(new(mfgmodel.WorkOrderJournal), "workorder/journal").DisableRoute("save").RegisterMWs(mfglogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(mfgmodel.WorkOrderJournalType), "workorder/journal/type"),
		// s.RegisterModel(new(mfgmodel.WorkOrderDailyReport), "workorder/report/daily").DisableRoute("get"),
		s.RegisterModel(new(mfgmodel.WorkOrderDailyReport), "workorder/report/daily").AllowOnlyRoute("get").RegisterPostMWs(mfglogic.MWPostWorkOrderDailyReport()),
		s.RegisterModel(new(mfgmodel.PhysicalAvailability), "physical/availability"),
		s.RegisterModel(new(mfgmodel.WorkRequestorJournalType), "workrequestor/journal/type"),
		// s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").DisableRoute("save", "gets").RegisterPostMWs(mfglogic.MWPostWorkOrderPlanGets()),
		// s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").AllowOnlyRoute("gets").RegisterPostMWs(mfglogic.MWPostWorkOrderPlanGets()),
		s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").DisableRoute("save").RegisterPostMWs(mfglogic.MWPostWorkOrderPlanGets()),
		s.RegisterModel(new(mfgmodel.WorkOrderPlanReport), "workorderplan/report").DisableRoute("save", "submit", "get"),
		s.RegisterModel(new(mfgmodel.WorkOrderPlanReportConsumption), "workorderplan/report/consumption"),
		s.RegisterModel(new(mfgmodel.WorkOrderPlanReportResource), "workorderplan/report/resource"),
		s.RegisterModel(new(mfgmodel.WorkOrderPlanReportOutput), "workorderplan/report/output"),

		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryMaterial), "workorderplan/summary/material").DisableRoute("save", "gets"),
		s.RegisterModel(new(mfgmodel.WorkOrderSummaryMaterial), "workorderplan/summary/material").AllowOnlyRoute("gets").RegisterPostMWs(mfglogic.MWPostWorkOrderSummaryMaterialGets()),
		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryMaterial), "workorderplan/summary/material").AllowOnlyRoute("gets"),
		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryMaterial), "workorderplan/summary/material").AllowOnlyRoute("gets").RegisterPostMWs(mfglogic.MWPostWorkOrderSummaryMaterialGets()),

		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryResource), "workorderplan/summary/resource").DisableRoute("save", "gets"),
		s.RegisterModel(new(mfgmodel.WorkOrderSummaryResource), "workorderplan/summary/resource").AllowOnlyRoute("gets").RegisterPostMWs(mfglogic.MWPostWorkOrderSummaryResourceGets()),
		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryResource), "workorderplan/summary/resource").AllowOnlyRoute("gets"),

		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryOutput), "workorderplan/summary/output").DisableRoute("save", "gets"),
		s.RegisterModel(new(mfgmodel.WorkOrderSummaryOutput), "workorderplan/summary/output").AllowOnlyRoute("gets").RegisterPostMWs(mfglogic.MWPostWorkOrderSummaryOutputGets()),
		// s.RegisterModel(new(mfgmodel.WorkOrderSummaryOutput), "workorderplan/summary/output").AllowOnlyRoute("gets"),
	)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(mfglogic.RoutineEngine), "routine").DisableRoute("add-new"),
		s.RegisterModel(new(mfglogic.RoutineEngine), "routine").AllowOnlyRoute("add-new").RegisterMWs(
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			tenantcorelogic.MWPreAssignSequenceNo("Routine", true, "_id"),
		),
		s.RegisterModel(new(mfglogic.RoutineDetailEngine), "routine/detail"),
		s.RegisterModel(new(mfglogic.RoutineChecklistEngine), "routine/checklist"),
		s.RegisterModel(new(mfglogic.BoMMaterialEngine), "bom/material"),
		s.RegisterModel(new(mfglogic.BoMManpowerEngine), "bom/manpower"),
		s.RegisterModel(new(mfglogic.BoMMachineryEngine), "bom/machinery"),
		s.RegisterModel(new(mfglogic.PostingProfileHandler), "postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
		s.RegisterModel(new(mfglogic.WorkOrderJournalEngine), "workorder/journal").AllowOnlyRoute("save"),
		s.RegisterModel(new(mfglogic.WODailyReportEngine), "workorder/report/daily"),
		s.RegisterModel(new(mfglogic.PhysicalAvailabilityEngine), "physical/availability"),
		s.RegisterModel(new(mfglogic.WorkOrderEngine), "workorder").DisableRoute("update-status"),
		s.RegisterModel(new(mfglogic.WorkRequestEngine), "work/request"),
		s.RegisterModel(new(mfglogic.WorkOrderPlanEngine), "workorderplan").DisableRoute("save", "gets-available-stock").RegisterMWs(mfglogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(mfglogic.WorkOrderPlanEngine), "workorderplan").AllowOnlyRoute("gets-available-stock").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		}), mfglogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(mfglogic.WorkOrderPlanEngine), "workorderplan").AllowOnlyRoute("save").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		}), mfglogic.MWPreAssignCompanyID(),
			tenantcorelogic.MWPreAssignSequenceNo("WorkOrder", true, "_id"),
		),
		s.RegisterModel(new(mfglogic.WorkOrderPlanReportEngine), "workorderplan/report").AllowOnlyRoute("get"),
		s.RegisterModel(new(mfglogic.WorkOrderPlanReportEngine), "workorderplan/report").DisableRoute("get").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
		s.RegisterModel(new(mfglogic.ApprovalAggregatorHandler), "approvalaggregator"),
	)

	//-- custom api for ui
	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(scmmodel.InventReceiveIssueLine), "workorder/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(scmmodel.InventReceiveIssueLine), "workorder/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.RoutineTemplateItems), "routine/template/items").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkDescription), "workorder/work").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(mfgmodel.WorkDescription), "workorder/work").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.ManPower), "manpower").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(mfgmodel.Machinery), "machinery").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(mfgmodel.ItemUsage), "workorder/work/itemusage").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(mfgmodel.ManPowerUsage), "workorder/work/manpowerusage").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(mfgmodel.Output), "workorder/work/output").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderDailyResume), "workorder/report/resume").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderGrid), "work/order").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkRequestGrid), "work/request").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderMaterialItem), "workorderplan/tab/material").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderResourceItem), "workorderplan/tab/resource").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(mfgmodel.WorkOrderOutputItem), "workorderplan/tab/output").AllowOnlyRoute("gridconfig"),
	)

	// nats api
	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(mfglogic.WorkOrderEngine), "workorder").AllowOnlyRoute("update-status").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
		s.RegisterModel(new(mfglogic.PostingProfileHandler), "postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
	)

	// nats api with model
	s.Group().SetMod(modDB).SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(mfgmodel.WorkOrder), "workorder").AllowOnlyRoute("insert").RegisterMWs(
			tenantcorelogic.MWPreAssignSequenceNo("WorkOrder", true, "_id"),
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			mfglogic.MWPreAssignCompanyID(),
		),
		s.RegisterModel(new(mfgmodel.WorkRequest), "work/request").AllowOnlyRoute("save").
			RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("WorkRequest", true, "_id"),
				kamis.JWT(kamis.JWTSetupOptions{
					Secret:           appConfig.Data.GetString("jwt_secret"),
					GetSessionMethod: "NATS",
					GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
				}),
			),
		s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").AllowOnlyRoute("save", "insert").RegisterMWs(
			tenantcorelogic.MWPreAssignSequenceNo("WorkOrder", true, "_id"),
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			mfglogic.MWPreAssignCompanyID(),
		),
	)
	s.RegisterModel(new(mfglogic.PreviewLogic), "preview")

	return nil
}
