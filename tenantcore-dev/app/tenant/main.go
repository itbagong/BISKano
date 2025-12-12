package main

import (
	"flag"
	"os"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoreconfig"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/kasset"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/tenant"
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
	if e := sebar.ConfigHasData(appConfig, "addr_auth_validation", "addr_access_validation"); e != nil {
		s.Log().Error(e.Error())
		os.Exit(1)
	}

	if err := serde.Serde(appConfig.Data, &tenantcoreconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	tenantcoreconfig.Config.EventHub = ev

	s3, e := SetS3Config(s, appConfig)
	if e != nil {
		s.Log().Error(e.Error())
	}

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")
	// s.RegisterMW(kamis.NeedJWT(), "nedjwt")

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()
	sequenceLogic := tenantcorelogic.DefaultSequence()

	func(models ...*kaos.ServiceModel) {
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"}
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			switch modelPtr.Name {
			case "cashbank":
				publicModels[index].
					RegisterMWs(rbaclogic.MWRbacFilterDim("", "jwt")).
					RegisterPostMWs(
						tenantcorelogic.MWPostCashbankGroup(),
						tenantcorelogic.MWPostCurrency(),
						tenantcorelogic.MWPostLedgerAccount("MainBalanceAccount"),
					)
			case "cashbankgroup":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostLedgerAccount("MainBalanceAccount"),
				)
			case "customer":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostCustomerGroup(),
				)
			case "employee":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostEmployeeGroup(),
				)
			case "vendor":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostVendorGroup(),
				)
			case "vendorgroup":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostLedgerAccount("MainBalanceAccount", "DepositAccount"),
				)
			case "asset":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostAssetGroup(),
				)
			case "company":
				publicModels[index].RegisterPostMWs(
					tenantcorelogic.MWPostCurrency("ReportingCurrency"),
				)
				// case "item":
				// publicModels[index].RegisterPostMWs(
				// 	tenantcorelogic.MWPostItemModel(),
				// )
			}
		}

		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
			AllowOnlyRoute(publicEndPoints...).
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
			Apply(models...)

	}(
		s.RegisterModel(new(tenantcoremodel.ExpenseType), "expensetype"),
		s.RegisterModel(new(tenantcoremodel.ExpenseTypeGroup), "expensetypegroup"),
		s.RegisterModel(new(tenantcoremodel.Contact), "contact"),
		s.RegisterModel(new(tenantcoremodel.Asset), "asset"),
		s.RegisterModel(new(tenantcoremodel.AssetGroup), "assetgroup"),
		s.RegisterModel(new(tenantcoremodel.CashBank), "cashbank").DisableRoute("insert"),
		s.RegisterModel(new(tenantcoremodel.CashBank), "cashbank").AllowOnlyRoute("insert").
			RegisterMWs(tenantcorelogic.MWPreAssignCustomSequenceNo("MasterCashBank")),
		s.RegisterModel(new(tenantcoremodel.CashBankGroup), "cashbankgroup"),
		s.RegisterModel(new(tenantcoremodel.Employee), "employee"),
		s.RegisterModel(new(tenantcoremodel.EmployeeGroup), "employeegroup"),
		s.RegisterModel(new(tenantcoremodel.Currency), "currency"),
		s.RegisterModel(new(tenantcoremodel.Company), "company"),
		s.RegisterModel(new(tenantcoremodel.SiteEntryJournalType), "siteentryjournaltype"),
		s.RegisterModel(new(tenantcoremodel.DimensionMaster), "dimension"),
		s.RegisterModel(new(tenantcoremodel.LedgerAccount), "ledgeraccount"),
		s.RegisterModel(new(tenantcoremodel.LedgerAccount), "ledgeraccount/coa").
			AllowOnlyRoute("find").
			RegisterMWs(func(ctx *kaos.Context, i interface{}) (bool, error) {
				ctx.Data().Set("DBModFilter", []*dbflex.Filter{
					dbflex.Ne("Status", "SCOA01"),
				})
				return true, nil
			}),
		s.RegisterModel(new(tenantcoremodel.CustomerGroup), "customergroup"),
		s.RegisterModel(new(tenantcoremodel.Customer), "customer").DisableRoute("insert"),
		s.RegisterModel(new(tenantcoremodel.Customer), "customer").AllowOnlyRoute("insert").
			RegisterMWs(tenantcorelogic.MWPreAssignSequenceNo("Customer", false, "_id")),
		s.RegisterModel(new(tenantcoremodel.VendorGroup), "vendorgroup"),
		s.RegisterModel(new(tenantcoremodel.Vendor), "vendor").DisableRoute("insert", "save"),
		s.RegisterModel(new(tenantcoremodel.Vendor), "vendor").AllowOnlyRoute("insert", "save").
			RegisterMWs(tenantcorelogic.MWPreAssignSequenceNo("Vendor", false, "_id")),
		s.RegisterModel(new(tenantcoremodel.ChecklistTemplate), "checklisttemplate"),
		s.RegisterModel(new(tenantcoremodel.ReferenceTemplate), "referencetemplate"),
		s.RegisterModel(new(tenantcoremodel.ItemTemplate), "itemtemplate"),
		s.RegisterModel(new(tenantcoremodel.UoM), "unit"),
		s.RegisterModel(new(tenantcoremodel.Item), "item").DisableRoute("gridconfig", "insert", "save", "delete"),
		s.RegisterModel(new(tenantcoremodel.Item), "item").AllowOnlyRoute("delete").RegisterMWs(tenantcorelogic.MWPreItemDelete(), tenantcorelogic.MWPreLogDelete()).RegisterPostMWs(tenantcorelogic.MWPostLogDelete(tenantcoremodel.LogMenuItem)),
		s.RegisterModel(new(tenantcoremodel.Item), "item").AllowOnlyRoute("insert", "save").RegisterMWs(tenantcorelogic.MWPreLogSave(), tenantcorelogic.MWPreItem()).RegisterPostMWs(tenantcorelogic.MWPostLogSave(tenantcoremodel.LogMenuItem)),
		s.RegisterModel(new(tenantcoremodel.ItemGroup), "itemgroup"),
		s.RegisterModel(new(tenantcoremodel.ItemSerial), "itemserial"),
		s.RegisterModel(new(tenantcoremodel.ItemBatch), "itembatch"),
		s.RegisterModel(new(tenantcoremodel.ItemSpec), "itemspec"),
		s.RegisterModel(new(tenantcoremodel.NumberSequence), "numseq"),
		s.RegisterModel(new(tenantcoremodel.NumberSequenceSetup), "numseqsetup"),
		s.RegisterModel(new(tenantcoremodel.LocationWarehouseGroup), "warehouse/group"),
		s.RegisterModel(new(tenantcoremodel.LocationWarehouse), "warehouse"),
		s.RegisterModel(new(tenantcoremodel.LocationSection), "section"),
		s.RegisterModel(new(tenantcoremodel.LocationAisle), "aisle"),
		s.RegisterModel(new(tenantcoremodel.LocationBox), "box"),
		s.RegisterModel(new(tenantcoremodel.UnitConversion), "unit/conversion"),
		s.RegisterModel(new(tenantcoremodel.SpecVariant), "specvariant"),
		s.RegisterModel(new(tenantcoremodel.SpecSize), "specsize"),
		s.RegisterModel(new(tenantcoremodel.SpecGrade), "specgrade"),
		// dynamic master
		s.RegisterModel(new(tenantcoremodel.MasterDataType), "masterdatatype"),
		s.RegisterModel(new(tenantcoremodel.MasterData), "masterdata"),
		// s.RegisterModel(new(tenantcoremodel.CustomItemDownload), "custom-item-download").DisableRoute("gets"), // moved to bellow, use tenantcorelogic.ItemSpecSearchResult
	)

	s.RegisterModel(new(tenantcoremodel.ItemGrid), "item").SetMod(modUI).AllowOnlyRoute("gridconfig")

	// http://localhost:37000/v1/tenant/warehouse
	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(sequenceLogic, "numseq"),
		s.RegisterModel(new(tenantcoremodel.Employee), "employee").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(new(tenantcoremodel.DimensionMaster), "dimension").SetMod(modDB).AllowOnlyRoute("find", "save", "delete"),
		s.RegisterModel(new(tenantcorelogic.ItemSpecEngine), "itemspec").AllowOnlyRoute("gets-info-by-id"),
		s.RegisterModel(new(tenantcoremodel.Asset), "asset").SetMod(modDB).RegisterMWs(tenantcorelogic.MWPostAssetGroup()),
		s.RegisterModel(new(tenantcoremodel.Item), "item").SetMod(modDB),
		s.RegisterModel(new(tenantcoremodel.LocationWarehouse), "warehouse").SetMod(modDB).AllowOnlyRoute("find"),
		s.RegisterModel(tenantcorelogic.NewPDFEngine(s3), "pdf"),
		s.RegisterModel(new(tenantcorelogic.ItemEngine), "custom-item-download").AllowOnlyRoute("download"),
	)

	s.RegisterModel(new(tenantcoremodel.ChecklistItem), "checklistitem").SetMod(modUI).AllowOnlyRoute("gridconfig", "formconfig")
	s.RegisterModel(new(tenantcoremodel.Preview), "preview").SetMod(modDB, modUI).AllowOnlyRoute("formconfig", "gridconfig", "get", "gets", "find")
	s.RegisterModel(new(tenantcorelogic.ItemSpecSearchResult), "itemspec/search").SetMod(modUI).AllowOnlyRoute("formconfig", "gridconfig")
	s.RegisterModel(new(tenantcorelogic.ItemSpecSearchResult), "custom-item-download").SetMod(modUI).AllowOnlyRoute("formconfig", "gridconfig")

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(tenantcorelogic.EmployeeEngine), "employee"),
		s.RegisterModel(new(tenantcorelogic.ItemSpecEngine), "itemspec").DisableRoute("gets"),
		s.RegisterModel(new(tenantcorelogic.UnitConversionEngine), "unit/conversion"),
		s.RegisterModel(new(tenantcorelogic.UnitEngine), "unit"),
		s.RegisterModel(new(tenantcorelogic.ItemEngine), "item").DisableRoute("gets"),
		// s.RegisterModel(new(tenantcorelogic.ItemEngine), "custom-item-download").AllowOnlyRoute("gets"),
		s.RegisterModel(new(tenantcorelogic.ItemSpecEngine), "custom-item-download").AllowOnlyRoute("gets"),
		//.DisableRoute("get-item-for-download"),
	)

	return nil
}

func SetS3Config(s *kaos.Service, appConfig *sebar.AppConfig) (*kasset.S3Asset, error) {
	var (
		e  error
		s3 *kasset.S3Asset
	)

	storage_ep := appConfig.Data.GetString("storage_end_point")
	storage_region := appConfig.Data.GetString("storage_region")
	storage_key := appConfig.Data.GetString("storage_key")
	storage_secret := appConfig.Data.GetString("storage_secret")
	storage_bucket := appConfig.Data.GetString("storage_bucket")

	//-- get config
	cfg := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(storage_key, storage_secret, ""))
	cfg.WithRegion(storage_region)

	//-- end-point for S3-alike, ie:minio
	if storage_ep != "" {
		cfg.WithEndpoint(storage_ep)
	}
	cfg.DisableSSL = aws.Bool(true)
	cfg.S3ForcePathStyle = aws.Bool(true)
	s3, e = kasset.NewS3WithConfig(storage_bucket, cfg)
	if e != nil {
		s.Log().Errorf("fail to connect asset storage service. %s", e.Error())
	}

	return s3, nil
}
