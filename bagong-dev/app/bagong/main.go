package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongconfig"
	"git.kanosolution.net/sebar/bagong/bagonglogic"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/bagong"
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
	if err := serde.Serde(appConfig.Data, &bagongconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	bagongconfig.Config.EventHub = ev
	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.EventHub = ev

	ev.SetTimeout(60 * time.Second)
	if err := serde.Serde(appConfig.Data, &bagonglogic.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()

	ficologic.RegisterLab(s)

	func(models ...*kaos.ServiceModel) {
		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModels[index] = s.RegisterModel(modelPtr.Model, modelPtr.Name).DisableRoute(modelPtr.DisableRoutes()...)

			switch modelPtr.Name {
			case "expense":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostLedgerAccount("DefaultLedgerAccountID"),
				)
			case "trayek":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostDimension("SiteID"),
					bagonglogic.MWPostTerminal("Terminals"),
				)
			case "siteentry":
				publicModels[index].RegisterMWs(func(ctx *kaos.Context, i interface{}) (bool, error) {
					jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
					dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})

					if len(dimIface) > 0 {
						dim := tenantcoremodel.Dimension{}
						if err := serde.Serde(dimIface, &dim); err != nil {
							return false, err
						}

						if len(dim) > 0 {
							vSiteID := []string{}
							for _, item := range dim {
								if item.Key == "Site" && item.Value != "" {
									vSiteID = append(vSiteID, item.Value)
								}
							}
							if len(vSiteID) > 0 {
								ctx.Data().Set("DBModFilter", []*dbflex.Filter{
									dbflex.In("SiteID", vSiteID...),
								})
							}
						}
					}

					return true, nil
				}).RegisterPostMWs(
					bagonglogic.MWPostSiteEntry(),
				)
			case "siteentry_asset":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostSiteEntryAsset(),
				)
			case "accident_fund":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostAccidentFund(),
				)
			case "asset":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostAsset(),
				)
			case "siteentry_nonasset":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostNonAsset(),
				)
			case "siteentry_trayekritase":
				publicModels[index].RegisterPostMWs(
					bagonglogic.MWPostSiteEntryTrayekRitase(),
				)
			}

		}
		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
			AllowOnlyRoute("formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new").
			Apply(publicModels...)

		//-- tenant admin
		s.Group().
			SetMod(modDB).
			SetDeployer(hd.DeployerName).
			RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
				Permission:          "BagongAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute("insert", "update", "delete", "save").
			Apply(models...)

	}(
		// s.RegisterModel(new(bagongmodel.Contact), "contact"),
		s.RegisterModel(new(bagongmodel.Asset), "asset").DisableRoute("insert", "update", "save", "delete", "get"),
		s.RegisterModel(new(bagongmodel.Terminal), "terminal"),
		s.RegisterModel(new(bagongmodel.Trayek), "trayek"),
		s.RegisterModel(new(bagongmodel.Claim), "claim"),
		s.RegisterModel(new(bagongmodel.BGVendor), "vendor").DisableRoute("save", "get"),
		s.RegisterModel(new(bagongmodel.EmployeeDetail), "employeedetail"),
		s.RegisterModel(new(bagongmodel.CustomerDetail), "customerdetail"),
		s.RegisterModel(new(bagongmodel.CustomerConfiguration), "customerconfiguration"),
		s.RegisterModel(new(bagongmodel.Expense), "expense"),
		s.RegisterModel(new(bagongmodel.VendorSubmission), "vendorsubmissiontype"),
		s.RegisterModel(new(bagongmodel.ExpenseType), "expensetype"),
		s.RegisterModel(new(bagongmodel.RentalContracts), "rentalcontract"),
		s.RegisterModel(new(bagongmodel.VendorJournalTypeConfiguration), "vendorjournaltypeconfiguration"),
		s.RegisterModel(new(bagongmodel.CustomerJournalTypeConfiguration), "customerjournaltypeconfiguration"),
		s.RegisterModel(new(bagongmodel.SiteEntryJournalTypeConfiguration), "siteentryjournaltypeconfiguration"),
		s.RegisterModel(new(bagongmodel.InventoryTransactionJournalTypeConfiguration), "inventorytransactionjournaltypeconfiguration"),
		s.RegisterModel(new(bagongmodel.MCUCondition), "mcucondition"),
		s.RegisterModel(new(bagongmodel.TaxInvoice), "taxinvoice"),
		// site entry trx
		s.RegisterModel(new(bagongmodel.SiteEntry), "siteentry"),
		s.RegisterModel(new(bagongmodel.SiteEntryAsset), "siteentry_asset"),
		s.RegisterModel(new(bagongmodel.SiteEntryNonAsset), "siteentry_nonasset").DisableRoute("save", "update", "insert"),
		s.RegisterModel(new(bagongmodel.SiteEntryBTSDetail), "siteentry_btsdetail").DisableRoute("save", "update", "insert"),
		s.RegisterModel(new(bagongmodel.SiteEntryMiningDetail), "siteentry_miningdetail").DisableRoute("save", "update", "insert"),
		s.RegisterModel(new(bagongmodel.SiteEntryMiningUsage), "siteentry_miningusage").DisableRoute("save", "update", "insert"),
		s.RegisterModel(new(bagongmodel.SiteEntryTrayekDetail), "siteentry_trayekdetail").DisableRoute("save", "update", "insert"),
		s.RegisterModel(new(bagongmodel.SiteEntryTrayekRitase), "siteentry_trayekritase").DisableRoute("save", "update", "insert", "delete"),
		s.RegisterModel(new(bagongmodel.SiteEntryTourismDetail), "siteentry_tourismdetail").DisableRoute("save", "update", "insert"),
		// consequence risk matrix
		s.RegisterModel(new(bagongmodel.Severity), "severity"),
		s.RegisterModel(new(bagongmodel.Likelihood), "likelihood"),
		s.RegisterModel(new(bagongmodel.RiskMatrix), "riskmatrix"),
		s.RegisterModel(new(bagongmodel.AccidentFund), "accident_fund"),
		s.RegisterModel(new(bagongmodel.AccidentFundDetail), "accident_funddetail"),
		s.RegisterModel(new(bagongmodel.AssetBooking), "assetbooking"),
		s.RegisterModel(new(bagongmodel.AssetBookingAllocation), "assetbooking-allocation"),
		s.RegisterModel(new(bagongmodel.AssetMovement), "asset-movement"),
	)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(bagonglogic.EmployeeHandler), "employee").DisableRoute("get-journal", "get-employee-by-id"),
		s.RegisterModel(new(bagonglogic.CustomerHandler), "customer"),
		s.RegisterModel(new(bagonglogic.PayrollHandler), "payroll"),
		s.RegisterModel(new(bagonglogic.VendorEngine), "vendor").AllowOnlyRoute("get-vendor-active", "get-vendor-by-site", "get-vendor-bank"),
		// site entry
		s.RegisterModel(new(bagonglogic.SiteEntryEngine), "siteentry"),
		s.RegisterModel(new(bagonglogic.TrayekEngine), "trayek"),
		s.RegisterModel(new(bagonglogic.AccidentFundEngine), "accident_fund"),
		s.RegisterModel(new(bagonglogic.InvoiceHandler), "invoice"),
		s.RegisterModel(new(bagonglogic.AssetEngine), "asset").AllowOnlyRoute("gets-filter", "generate-depreciation", "get-depreciation", "get-asset-detail"),
		s.RegisterModel(new(bagonglogic.AssetEngine), "asset").AllowOnlyRoute("get-assets").RegisterPostMWs(bagonglogic.MWPostAsset()),

		s.RegisterModel(new(bagonglogic.PostingProfileHandler), "postingprofile"),
		s.RegisterModel(new(bagonglogic.TaxEngine), "taxtransaction"),
		s.RegisterModel(new(bagonglogic.EmployeeHandler), "employee").AllowOnlyRoute("get-journal"),
		s.RegisterModel(new(bagonglogic.MasterUploadEngine), "masterupload"),
		s.RegisterModel(new(bagonglogic.AssetMovementEngine), "asset-movement"),
		s.RegisterModel(new(bagonglogic.CashHandler), "cash"),
	)
	s.Group().SetMod(modDB).AllowOnlyRoute("delete").Apply(
		s.RegisterModel(new(bagongmodel.BGPayrollSubmission), "payroll"),
	)

	s.Group().
		SetDeployer(hd.DeployerName).
		RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
			Permission:          "BagongAdmin",
			RequiredAccessLevel: 7,
			CheckFunction:       checkAccessFn,
		})).
		AllowOnlyRoute("save", "get").
		Apply(
			s.RegisterModel(new(bagonglogic.VendorEngine), "vendor"),
			s.RegisterModel(new(bagonglogic.PayrollBenefitEngine), "payrollbenefit"),
			s.RegisterModel(new(bagonglogic.PayrollDeductionEngine), "payrolldeduction"),
		)

	//-- custom api with auth
	// s.Group().
	// 	RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
	// 		Secret:           appConfig.Data.GetString("jwt_secret"),
	// 		GetSessionMethod: "NATS",
	// 		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	// 	})).
	// 	SetDeployer(hd.DeployerName).
	// 	Apply(
	// 		s.RegisterModel(new(bagonglogic.PostingProfileHandler), "postingprofile"),
	// 		s.RegisterModel(new(bagonglogic.EmployeeHandler), "employee").AllowOnlyRoute("get-journal"),
	// 	)

	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		// s.RegisterModel(new(bagongmodel.Contact), "customer/contact").AllowOnlyRoute("gridconfig"),
		// s.RegisterModel(new(bagongmodel.Contact), "customer/contact").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.VendorTerm), "vendor/term").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.BGPayrollBenefitDetail), "payrollbenefit/detail").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.BGPayrollDeductionDetail), "payrolldeduction/detail").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.AssetUnit), "asset/detail/unit").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.AssetProperty), "asset/detail/property").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.AssetRegisterInfo), "asset/detail/registerinfo").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.AssetElectronic), "asset/detail/electronic").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.Depreciation), "asset/detail/depreciation").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.DepreciationActivity), "asset/detail/depreciationactivity").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.AssetOtherInfo), "asset/detail/assetotherinfo").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.VendorSubmissionGrid), "vendorsubmission").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.VendorSubmissionForm), "vendorsubmission").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.VendorSubmissionLineGrid), "vendorsubmission/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.VendorSubmissionLineForm), "vendorsubmission/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.VendorContact), "vendor/contact").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.VendorBank), "vendor/bank").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.TirePosition), "tire_position").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteExpense), "site_expense").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteIncome), "site_income").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteAttachment), "site_attachment").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.Corridor), "corridor").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.Corridor), "corridor").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.UserInfo), "user_info").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.UserInfo), "user_info").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.RitaseDetail), "siteentry_btsdetail/detail/ritasedetail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.KMDetail), "siteentry_miningdetail/detail/kmdetail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.Shift), "shift").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.Shift), "shift").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.Configuration), "siteconfiguration").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.Overtime), "overtime").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.Overtime), "overtime").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.ClaimLine), "claim/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.ClaimSummaryAmount), "claim/summary_amount").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.PettyCashSubmissionDetail), "pettycashsubmission_detail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.EmployeeExpenseSubmissionDetail), "employeeexpensesubmission_detail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.EmployeeAttachment), "employee_attachment").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.EmployeeAttachment), "employee_attachment").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.AssetBookingLine), "assetbooking/lines").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.AssetBookingLine), "assetbooking/lines").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.FamilyMembers), "employee/familymembers").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.FamilyMembers), "employee/familymembers").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.RitaseFuelUsage), "siteentry_btsdetail/fuelusage").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.RitaseFuelUsage), "siteentry_btsdetail/fuelusage").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.PayrollSubmissionCustom), "payroll_submission_custom").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(bagongmodel.PayrollSubmissionCustom), "payroll_submission_custom").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SubmissionLines), "payroll_submission_custom/submission_lines").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(bagongmodel.SiteExpenseTrayekGrid), "siteexpense-trayek/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteExpenseTrayekReadGrid), "siteexpense-trayek-read/grid").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(bagongmodel.SiteExpenseGrid), "siteexpense/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteExpenseReadGrid), "siteexpense-read/grid").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(bagongmodel.SiteIncomeGrid), "siteincome/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.SiteIncomeReadGrid), "siteincome-read/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(bagongmodel.AssetMovementLine), "asset-movement/line").AllowOnlyRoute("gridconfig"),
	)
	s.Group().SetMod(modDB).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(bagongmodel.VendorSubmission), "vendorsubmission").
			AllowOnlyRoute("gets").RegisterPostMWs(tenantcorelogic.MWPostVendorName()),
		s.RegisterModel(new(bagongmodel.VendorSubmission), "vendorsubmission").DisableRoute("gets"),
	)
	s.RegisterModel(new(bagongmodel.PettyCashSubmission), "pettycash_submission").SetMod(modUI, modDB).SetDeployer(hd.DeployerName).AllowOnlyRoute("save", "find", "formconfig", "gridconfig", "listconfig", "gets", "get", "delete")
	s.RegisterModel(new(bagongmodel.EmployeeExpenseSubmission), "employeeexpense_submission").SetMod(modUI, modDB).SetDeployer(hd.DeployerName).AllowOnlyRoute("save").RegisterMW(
		func(ctx *kaos.Context, payload interface{}) (bool, error) {
			//fmt.Println("DDDDLDLD", payload.(*bagongmodel.EmployeeExpenseSubmission))
			pp, ok := payload.(*bagongmodel.EmployeeExpenseSubmission)
			if !ok {
				return true, errors.New("Tipe data tidak sama")
			}

			if pp.EmployeeName == "" {
				// hub := sebar.GetTenantDBFromContext(ctx)
				// ctx.DefaultEvent()
				evHub, _ := ctx.DefaultEvent()
				employee := []tenantcoremodel.Employee{}
				param := dbflex.NewQueryParam()
				filter := dbflex.NewFilter("_id", dbflex.OpEq, pp.EmployeeID, nil)
				param.Where = filter
				err := evHub.Publish("/v1/tenant/employee/find", param, &employee, nil)
				if err == nil {
					pp.EmployeeName = employee[0].Name
					//fmt.Println(pp.EmployeeName)

				} else {
					return true, err
				}
				return true, nil
			} else {
				return true, nil
			}
		}, "preSaveEmployeeExpenseSubmission")
	s.RegisterModel(new(bagongmodel.EmployeeExpenseSubmission), "employeeexpense_submission").SetMod(modUI, modDB).SetDeployer(hd.DeployerName).AllowOnlyRoute("find", "formconfig", "gridconfig", "listconfig", "gets", "get", "delete")

	// register generate summary
	s.RegisterModel(new(bagongmodel.SiteEntryTrayekRitase), "siteentry_trayekritase").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save", "delete").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryTrayekRitase)
				bagonglogic.GenerateSummary(ctx, res.SiteEntryAssetID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryTrayekDetail), "siteentry_trayekdetail").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryTrayekDetail)
				bagonglogic.GenerateSummary(ctx, res.ID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryTourismDetail), "siteentry_tourismdetail").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryTourismDetail)
				bagonglogic.GenerateSummary(ctx, res.ID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryMiningDetail), "siteentry_miningdetail").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryMiningDetail)
				bagonglogic.GenerateSummary(ctx, res.ID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryMiningUsage), "siteentry_miningusage").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryMiningUsage)
				bagonglogic.GenerateSummary(ctx, res.ID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryBTSDetail), "siteentry_btsdetail").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryBTSDetail)
				bagonglogic.GenerateSummary(ctx, res.ID)
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.SiteEntryNonAsset), "siteentry_nonasset").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := payload.(*bagongmodel.SiteEntryNonAsset)
				h := sebar.GetTenantDBFromContext(ctx)
				// gets SiteEntryAsset
				siteEntryAssets := []bagongmodel.SiteEntryAsset{}
				if e := h.GetsByFilter(new(bagongmodel.SiteEntryAsset), dbflex.Eq("SiteEntryID", res.ID), &siteEntryAssets); e != nil {
					ctx.Log().Errorf("SiteEntryAsset not found: %s", res.ID)
					return true, nil
				}
				siteIncome := 0.0
				siteExpense := 0.0
				for _, val := range siteEntryAssets {
					siteIncome += val.Income
					siteExpense += val.Expense
				}

				siteEntry := new(bagongmodel.SiteEntry)
				if e := h.GetByID(siteEntry, res.ID); e != nil {
					ctx.Log().Errorf("Failed populate data SiteEntry: %s", e.Error())
					return true, nil
				}
				siteEntry.Income = res.Income + siteIncome
				siteEntry.Expense = res.Expense + siteExpense
				if siteEntry.Purpose == "Trayek" {
					siteEntry.Revenue = (siteEntry.Income - siteEntry.Expense)
				} else {
					siteEntry.Revenue = 0
				}
				if e := h.Save(siteEntry); e != nil {
					ctx.Log().Errorf("Save data SiteEntry error: %s", e.Error())
					return true, nil
				}
				return true, nil
			})

	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("get").
		RegisterPostMWs(
			bagonglogic.MWPostGetAsset(),
		)

	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := ctx.Data().Get("FnResult", new(bagongmodel.Asset)).(*bagongmodel.Asset)
				if res.ID != "" {
					h := sebar.GetTenantDBFromContext(ctx)
					tenantAsset := new(tenantcoremodel.Asset)
					e := h.GetByID(tenantAsset, res.ID)
					if e != nil {
						ctx.Log().Warningf("error: fail to get tenant asset")
						return true, nil
					}
					if tenantAsset.GroupID == "ELC" || tenantAsset.GroupID == "PRT" {
						return true, nil
					}
					ev, _ := ctx.DefaultEvent()
					if ev == nil {
						ctx.Log().Warningf("missing: event")
						return true, nil
					}
					if err := ev.Publish("/v1/tenant/dimension/save", &tenantcoremodel.DimensionMaster{
						ID:            res.ID,
						Label:         res.DetailUnit.PoliceNum,
						DimensionType: "Asset",
						IsActive:      res.IsActive,
					}, nil, nil); err != nil {
						ctx.Log().Warningf("sync dimension asset: %s", err.Error())
					}
				}
				return true, nil
			})

	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(knats.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := ctx.Data().Get("FnResult", new(bagongmodel.Asset)).(*bagongmodel.Asset)
				if res.ID != "" {
					h := sebar.GetTenantDBFromContext(ctx)
					tenantAsset := new(tenantcoremodel.Asset)
					e := h.GetByID(tenantAsset, res.ID)
					if e != nil {
						ctx.Log().Warningf("error: fail to get tenant asset")
						return true, nil
					}
					if tenantAsset.GroupID == "ELC" || tenantAsset.GroupID == "PRT" {
						return true, nil
					}
					ev, _ := ctx.DefaultEvent()
					if ev == nil {
						ctx.Log().Warningf("missing: event")
						return true, nil
					}
					if err := ev.Publish("/v1/tenant/dimension/save", &tenantcoremodel.DimensionMaster{
						ID:            res.ID,
						Label:         res.DetailUnit.PoliceNum,
						DimensionType: "Asset",
						IsActive:      res.IsActive,
					}, nil, nil); err != nil {
						ctx.Log().Warningf("sync dimension asset: %s", err.Error())
					}
				}
				return true, nil
			})

	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("delete").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				asset, ok := payload.(*bagongmodel.Asset)
				if !ok {
					return true, nil
				}
				ev, _ := ctx.DefaultEvent()
				if ev == nil {
					ctx.Log().Warningf("missing: event")
					return true, nil
				}

				if err := ev.Publish("/v1/tenant/dimension/delete", asset, nil, nil); err != nil {
					ctx.Log().Warningf("delete dimension asset: %s", err.Error())
				}
				return true, nil
			})

	s.RegisterModel(new(bagongmodel.Site), "sitesetup").SetMod(modUI, modDB).SetDeployer(hd.DeployerName).RegisterMWs(rbaclogic.MWRbacFilterDim("", "jwt")).DisableRoute("insert", "update", "save", "delete")
	s.RegisterModel(new(bagongmodel.Site), "sitesetup").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "update", "save").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				res := ctx.Data().Get("FnResult", new(bagongmodel.Site)).(*bagongmodel.Site)
				if res.ID != "" {
					ev, _ := ctx.DefaultEvent()
					if ev == nil {
						ctx.Log().Warningf("missing: event")
						return true, nil
					}

					if err := ev.Publish("/v1/tenant/dimension/save", &tenantcoremodel.DimensionMaster{
						ID:            res.ID,
						Label:         res.Name,
						DimensionType: "Site",
					}, nil, nil); err != nil {
						ctx.Log().Warningf("sync dimension site: %s", err.Error())
					}
				}
				return true, nil
			})
	s.RegisterModel(new(bagongmodel.Site), "sitesetup").SetMod(modDB).SetDeployer(hd.DeployerName).
		AllowOnlyRoute("delete").
		RegisterPostMWs(
			func(ctx *kaos.Context, payload interface{}) (bool, error) {
				site, ok := payload.(*bagongmodel.Site)
				if !ok {
					return true, nil
				}
				ev, _ := ctx.DefaultEvent()
				if ev == nil {
					ctx.Log().Warningf("missing: event")
					return true, nil
				}

				if err := ev.Publish("/v1/tenant/dimension/delete", site, nil, nil); err != nil {
					ctx.Log().Warningf("delete dimension site: %s", err.Error())
				}
				return true, nil
			})

	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(knats.DeployerName).AllowOnlyRoute("get")
	s.RegisterModel(new(bagongmodel.Asset), "asset").SetMod(modDB).SetDeployer(knats.DeployerName).AllowOnlyRoute("gets").RegisterPostMWs(
		bagonglogic.MWPostAsset(),
	)

	s.RegisterModel(new(karamodel.AttendanceTrx), "admin/trx").
		SetDeployer(hd.DeployerName).SetMod(modDB).
		AllowOnlyRoute("gets").
		RegisterMW(func(ctx *kaos.Context, i interface{}) (bool, error) {
			param, ok := i.(*dbflex.QueryParam)
			if !ok {
				fmt.Println("error convert sam")
			}

			ids := []interface{}{}
			filters := []*dbflex.Filter{}
			if param.Where != nil {
				for _, item := range param.Where.Items {
					if item.Field == "DirectSupervisor" {
						ids = item.Value.([]interface{})
						continue
					}

					// for save another filters
					// delete DirectSupervisor filter, will change it with UserID
					filters = append(filters, item)
				}
			}

			// find & set employee filter, when direct supervisor is filled
			if len(ids) > 0 {
				h := sebar.GetTenantDBFromContext(ctx)
				type EmployeeID struct {
					EmployeeID string
				}
				employees := []EmployeeID{}
				h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
					dbflex.In("DirectSupervisor", ids...),
				).SetSelect("EmployeeID"), &employees)

				empIds := make([]string, len(employees))
				for i, e := range employees {
					empIds[i] = e.EmployeeID
				}

				filters = append(filters, dbflex.In("UserID", empIds...))
				param.Where = dbflex.And(filters...)
			}

			return true, nil
		}, "get-employee").
		RegisterPostMW(func(ctx *kaos.Context, i interface{}) (bool, error) {
			res := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
			db := sebar.GetTenantDBFromContext(ctx)
			wls := sebar.NewMapRecordWithORM(db, new(karamodel.WorkLocation))
			asset := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Asset))
			masterData := sebar.NewMapRecordWithORM(db, new(tenantcoremodel.MasterData))
			trxs := *(res.Get("data", &([]karamodel.AttendanceTrx{})).(*[]karamodel.AttendanceTrx))
			for idx, trx := range trxs {
				wl, err := wls.Get(trx.WorkLocationID)
				if err == nil {
					trx.WorkLocationID = wl.Name
				}
				ast, err := asset.Get(trx.Ref1)
				if err == nil {
					trx.Ref2 = ast.Name
				}
				master, err := masterData.Get(trx.Message)
				if err == nil {
					trx.Message = master.Name
				}
				trxs[idx] = trx
			}
			res.Set("data", &trxs)
			ctx.Data().Set("FnResult", res)
			return true, nil
		}, "trxGets")

	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(bagonglogic.AssetEngine), "asset").AllowOnlyRoute("acquire", "add-asset"),
		s.RegisterModel(new(bagonglogic.VendorEngine), "vendor").AllowOnlyRoute("get"),
		s.RegisterModel(new(bagonglogic.SiteEngine), "sitesetup").AllowOnlyRoute("get-site-ids", "get-sites", "get-site-by-id"),
		s.RegisterModel(new(bagonglogic.AssetEngine), "asset").AllowOnlyRoute("gets-asset-active-per-site"),
		s.RegisterModel(new(bagonglogic.CustomerHandler), "customer").AllowOnlyRoute("get"),
		s.RegisterModel(new(bagonglogic.EmployeeHandler), "employee").AllowOnlyRoute("get-employee-by-id"),
	)

	s.RegisterRoute(func(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
		evHub, _ := ctx.DefaultEvent()

		menu := ""
		if hr, ok := ctx.Data().Get("http_request", nil).(*http.Request); ok {
			queryValues := hr.URL.Query()
			for _, vs := range queryValues {
				if len(vs) > 0 {
					menu = vs[0]
				}
			}
		} else {
			return nil, errors.New("menu doesn't specified")
		}

		payload = payload.MergeWhere(false, dbflex.And(
			dbflex.Eq("AppID", "bagong"),
			dbflex.Eq("Section", "Transaction"),
			dbflex.Contains("Uri", fmt.Sprintf("/%s/", menu)),
		))

		reply := []rbacmodel.AppMenu{}
		err := evHub.Publish("/v1/admin/menu/find", payload, &reply, nil)
		if err != nil {
			return nil, fmt.Errorf("error when get menu: %s", err.Error())
		}

		return reply, nil
	}, "menu")

	return nil
}
