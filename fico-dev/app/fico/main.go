package main

import (
	"flag"
	"os"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/leekchan/accounting"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/fico"
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
	ev.SetTimeout(60 * time.Second)
	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.FinancialPeriodModules = []string{"Finance", "Inventory"}
	ficoconfig.Config.EventHub = ev
	ficoconfig.Config.Log = s.Log()

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")
	// s.RegisterMW(kamis.NeedJWT(), "nedjwt")

	if ficoconfig.Config.AddrAccessValidation == "" {
		s.Log().Error("missinng: AddrAccessValidation")
		os.Exit(-1)
	}

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()

	func(models ...*kaos.ServiceModel) {
		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModels[index] = s.RegisterModel(modelPtr.Model, modelPtr.Name)

			switch modelPtr.Name {
			case "customerjournaltype", "vendorjournaltype", "cashjournaltype":
				publicModels[index].RegisterPostMWs(
					ficologic.MWPostPostingProfileID(),
					ficologic.MWPostTaxCodeIDs(),
					ficologic.MWPostChargeCodeIDs(),
				)
			case "loansetup":
				publicModels[index].RegisterPostMWs(ficologic.MWPostGetsLoanSetup())
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
				Permission:          "FicoAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute("insert", "update", "delete", "save").
			Apply(models...)
	}(
		s.RegisterModel(new(ficomodel.PostingProfile), "postingprofile").DisableRoute("update"),
		s.RegisterModel(new(ficomodel.PostingProfilePIC), "postingprofile/pic").DisableRoute("update"),
		s.RegisterModel(new(ficomodel.PaymentTerm), "paymentterm"),
		s.RegisterModel(new(ficomodel.TaxSetup), "taxcode"),
		s.RegisterModel(new(ficomodel.TaxTransaction), "taxtransaction"),
		s.RegisterModel(new(ficomodel.ChargeSetup), "chargecode"),
		s.RegisterModel(new(ficomodel.LedgerJournalType), "ledgerjournaltype"),
		s.RegisterModel(new(ficomodel.VendorJournalType), "vendorjournaltype"),
		s.RegisterModel(new(ficomodel.CustomerJournalType), "customerjournaltype"),
		s.RegisterModel(new(ficomodel.SiteEntryJournalType), "siteentryjournaltype"),
		s.RegisterModel(new(ficomodel.TaxSetup), "taxsetup"),
		s.RegisterModel(new(ficomodel.ChargeSetup), "chargesetup"),
		s.RegisterModel(new(ficomodel.PayrollBenefit), "payrollbenefit"),
		s.RegisterModel(new(ficomodel.PayrollDeduction), "payrolldeduction"),
		s.RegisterModel(new(ficomodel.LoanSetup), "loansetup"),
		s.RegisterModel(new(ficomodel.CashJournalType), "cashjournaltype"),
		s.RegisterModel(new(ficomodel.AssetJournalType), "assetjournaltype"),
		s.RegisterModel(new(ficomodel.SheJournalType), "shejournaltype"),
		s.RegisterModel(new(ficomodel.FixedAssetNumber), "fixedassetnumber").DisableRoute("insert", "save", "delete"),
		s.RegisterModel(new(ficomodel.FixedAssetNumberList), "fixedassetnumberlist").AllowOnlyRoute("get", "gets", "find"),
		s.RegisterModel(new(ficomodel.TaxGroup), "taxgroup"),
	)

	s.Group().
		SetMod(modDB).
		SetDeployer(hd.DeployerName).
		AllowOnlyRoute("update").
		Apply(
			s.RegisterModel(new(ficomodel.PostingProfilePIC), "postingprofile/pic").
				RegisterPostMWs(ficologic.MWPostSavePostingProfilePIC()),
			s.RegisterModel(new(ficomodel.PostingProfile), "postingprofile").
				RegisterPostMWs(ficologic.MWPostSavePostingProfile()),
		)

	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(ficomodel.PostingUsers), "postingprofile/approver"),
		s.RegisterModel(new(ficomodel.JournalTypeContext), "journaltypecontext").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.AddressAndTax), "vendorjournal/address").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(ficomodel.VendorJournalLineGrid), "vendorjournal/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.VendorJournalLineForm), "vendorjournal/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(ficomodel.CustomerJournalLineGrid), "customerjournal/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.CustomerJournalLineForm), "customerjournal/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(ficomodel.JournalLine), "journal/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.JournalLine), "journal/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(ficomodel.LoanLine), "loansetup/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.CashInJournalLineGrid), "cashin/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.CashOutJournalLineGrid), "cashout/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.SubmissionemployeeJournalLineGrid), "submissionemployee/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.ApplyAdjustmentJournalLineForm), "applyadjustment/line").AllowOnlyRoute("formconfig"),

		s.RegisterModel(new(ficomodel.SubmissionEmployeeLedgerJournal), "submissionemployee/journal").AllowOnlyRoute("gridconfig", "formconfig"),

		s.RegisterModel(new(ficomodel.AddressAndTax), "customerjournal/address").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(ficomodel.FixedAssetNumberDetail), "fixedassetnumber/detail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.CashReconGrid), "cashrecongrid").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(ficomodel.CashRecon), "cashrecon").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(ficomodel.ConfirmApplyGrid), "apply/confirm").AllowOnlyRoute("gridconfig"),

		s.RegisterModel(new(ficomodel.CashTransaction), "cashtransaction").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(ficomodel.TransactionHistory), "transaction_history").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.TransactionHistoryCashBank), "transaction_history/cashbank").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(ficomodel.TransactionHistoryCOA), "transaction_history/coa").AllowOnlyRoute("gridconfig"),
	)

	ficologic.RegisterLedgerJournal(s, modDB, modUI)
	ficologic.RegisterCustomerJournal(s, modDB, modUI)
	ficologic.RegisterVendorJournal(s, modDB, modUI)
	ficologic.RegisterCashJournal(s, modDB, modUI)
	ficologic.RegisterFiscalYear(s, modDB, modUI)
	ficologic.RegisterCGBook(s)
	ficologic.RegisterLab(s)

	s.RegisterModel(new(ficologic.ApplyLogic), "apply").SetDeployer(hd.DeployerName)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(ficologic.PostingProfileHandler), "postingprofile").DisableRoute("get-approval-by-source-user"),
		s.RegisterModel(new(ficologic.RecalcHandler), "recalc").DisableRoute("fix-by-type-ev"),
		s.RegisterModel(new(ficologic.LoanSetupHandler), "loansetup"),
		s.RegisterModel(new(ficologic.FixedAssetNumberHandler), "fixedassetnumber"),
		s.RegisterModel(new(ficologic.FixedAssetNumberListEngine), "fixedassetnumberlist").AllowOnlyRoute("gets-filter"),
		s.RegisterModel(new(ficologic.ApplyHandler), "apply"),
		s.RegisterModel(new(ficologic.LedgerAccountBalanceHandler), "ledgeraccountbalance"),
		s.RegisterModel(new(ficologic.CashBankBalanceHandler), "cashbankbalance"),
		s.RegisterModel(new(ficologic.CustomerBalanceHandler), "customerbalance"),
		s.RegisterModel(new(ficologic.VendorBalanceHandler), "vendorbalance"),
		s.RegisterModel(new(ficologic.CashReconLogic), "cashrecon"),
		s.RegisterModel(new(ficologic.CashBalanceHandler), "cashbalance"),
		s.RegisterModel(new(ficologic.PettyCashHandler), "pettycash"),
		s.RegisterModel(new(ficologic.EmployeeExpenseHandler), "employeeexpense"),
		s.RegisterModel(new(ficologic.ApprovalAggregatorHandler), "approvalaggregator"),
		s.RegisterModel(new(ficologic.ApprovalLogHandler), "approvallog"),
		s.RegisterModel(new(ficologic.PreviewLogic), "preview"),
		s.RegisterModel(new(ficologic.NotificationHandler), "notification"),
		s.RegisterModel(new(ficologic.PostingProfilePICHandler), "postingprofile/pic"),

		/*
			// NOTE & TODO: ini class-e belum di push, merah soale
			s.RegisterModel(new(ficologic.VendorJournalTypeHandler), "vendorjournaltypehandler"),
			s.RegisterModel(new(ficologic.CustomerJournalTypeHandler), "customerjournaltypehandler"),
			s.RegisterModel(new(ficologic.JournalHandler), "journal"),
		*/
	)

	//-- custom api with auth
	s.Group().
		RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})).
		SetDeployer(hd.DeployerName).
		Apply(
			s.RegisterModel(new(ficologic.PostingProfileHandler), "postingprofile").AllowOnlyRoute("get-approval-by-source-user"),
		)

	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(ficomodel.PayrollBenefit), "payrollbenefit").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(ficomodel.PayrollDeduction), "payrolldeduction").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(ficomodel.VendorJournalType), "vendorjournaltype").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(ficomodel.CustomerJournalType), "customerjournaltype").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(ficologic.EventPostingProfile), "postingprofile").AllowOnlyRoute("post"),
		s.RegisterModel(new(ficologic.FixedAssetNumberListEngine), "fixedassetnumberlist").AllowOnlyRoute("use"),
		s.RegisterModel(new(ficomodel.TaxSetup), "taxsetup").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(ficologic.PostingProfileHandler), "postingprofile").AllowOnlyRoute("map-source-data-to-url"),
		s.RegisterModel(new(ficologic.RecalcHandler), "recalc").AllowOnlyRoute("fix-by-type-ev"),
	)

	//-- model api with nats
	s.Group().SetMod(modDB).SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(ficomodel.VendorJournal), "vendorjournal").AllowOnlyRoute("insert").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
	)

	accounting.DefaultAccounting("", 2)
	return nil
}
