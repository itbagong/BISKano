package main

import (
	"flag"
	"os"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
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
	serviceName = "v1/scm"
	logger      = sebar.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	if e := sebar.ConfigHasData(appConfig, "addr_auth_validation", "addr_access_validation"); e != nil {
		s.Log().Error(e.Error())
		os.Exit(1)
	}
	ev.SetTimeout(1 * time.Minute)

	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.EventHub = ev

	if err := serde.Serde(appConfig.Data, &scmconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	scmconfig.Config.SetEventHub(ev)

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")

	func(models ...*kaos.ServiceModel) {
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"}
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}
		rbacEndPoints := []string{"get", "gets", "find"}
		rbacModelList := []string{
			// journals
			"purchase/order",
			"purchase/request",
			"inventory/journal",
			"inventory/receive",
			"item/request",
			"asset-acquisition",

			// other than journals
			"good-receipt",
			"good-issue",
			// "inventoryadjustment",
			// "stock-opname",
			// "item/balance",
		}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		// publicModels := []*kaos.ServiceModel{}
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			// for NON-RBAC Models OR RBAC Model with "formconfig", "gridconfig", "listconfig", "new"
			switch modelPtr.Name {
			case "movementintransaction/detail":
				publicModels[index].RegisterPostMWs(
					scmlogic.MWPostMovementInItemDetail(),
				)
			case "movementouttransaction/detail":
				publicModels[index].RegisterPostMWs(
					scmlogic.MWPostMovementOutItemDetail(),
				)
			case "transfer/in", "transfer/out":
				publicModels[index].RegisterPostMWs(
					scmlogic.MWPostItem(),
				)
			case "item/min-max":
				publicModels[index].RegisterPostMWs(
					scmlogic.MWPostItemMinMaxGets(),
				)
			case "inventoryadjustment":
				publicModels[index].RegisterPostMWs(
					scmlogic.MWPostInventoryAdjustmentGets(),
				).RegisterMWs(
				// rbaclogic.MWRbacFilterDim("", "jwt"),
				// scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
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
				case "item/request":
					rbacModel.RegisterPostMWs(
						scmlogic.MWPostItemRequest(),
					).RegisterMWs(
						scmlogic.MWPreInjectFindByKeywordApproval(),
					)
				case "purchase/request":
					rbacModel.RegisterPostMWs(
						scmlogic.MWPostPurchaseRequestGets(),
						scmlogic.MWPostPurchaseRequestGet(),
					)
				case "purchase/order":
					rbacModel.RegisterPostMWs(
						scmlogic.MWPostPurchaseOrderGets(),
						scmlogic.MWPostPurchaseOrderGet(),
					)
				case "inventory/journal":
					rbacModel.RegisterPostMWs(
						scmlogic.MWPostInventoryJournal(),
					)
				case "inventory/receive":
					rbacModel.RegisterPostMWs(
						scmlogic.MWPostInventoryReceive(),
					)
					// case "item/balance":
					// 	rbacModel.RegisterPostMWs(
					// 		scmlogic.MWPostItemBalance(),
					// 	)
				}

				publicModels = append(publicModels, rbacModel)
			} else {
				publicModels[index].AllowOnlyRoute(publicEndPoints...)
			}
		}

		//-- protected
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		protectedModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			protectedModel := new(kaos.ServiceModel)
			*protectedModel = *modelPtr
			s.AddServiceModel(protectedModel)
			protectedModels[index] = protectedModel

			switch modelPtr.Name {
			case "inventory/journal":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventJournal", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					// scmlogic.MWPreAssignSite(),
				)
			case "inventory/receive":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventReceiveIssueJournal", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					scmlogic.MWPreInventoryReceiveSave(),
				)
			case "purchase/request":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("PurchaseRequest", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					scmlogic.MWPrePRCalcTax(),
				)
			case "purchase/order":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("PurchaseOrder", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					scmlogic.MWPrePurchaseOrderSave(),
				)
			case "item/request":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("ItemRequest", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
				)
			case "asset-acquisition":
				protectedModels[index].RegisterMWs(
					scmlogic.MWPreAssignCompanyID(),
				)
			case "inventory/transaction":
				protectedModels[index].RegisterMWs(
					scmlogic.MWPreAssignCompanyID(),
				)
			case "stock-opname":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("StockOpname", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					rbaclogic.MWRbacFilterDim("", "jwt"),
					scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
				)
			case "inventoryadjustment":
				protectedModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventoryAdjustment", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					rbaclogic.MWRbacFilterDim("", "jwt"),
					scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
				)
			case "item/min-max":
				protectedModels[index].RegisterMWs(
					scmlogic.MWPreItemMinMaxSave(),
					tenantcorelogic.MWPreAssignSequenceNo("ItemMinMax", true, "_id"),
					scmlogic.MWPreAssignCompanyID(),
					rbaclogic.MWRbacFilterDim("", "jwt"),
					scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
				)
			}
		}

		// public
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

		// protected
		s.Group().
			SetMod(modDB).
			SetDeployer(hd.DeployerName).
			RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			})).
			AllowOnlyRoute(protectedEndPoints...).
			Apply(protectedModels...)

	}(
		s.RegisterModel(new(scmmodel.InventoryTransactionJournalType), "inventorytransactionjournaltype"), // TODO: remove this change to InventJournalType
		s.RegisterModel(new(scmmodel.ItemBalance), "item/balance").DisableRoute("gridconfig").RegisterPostMWs(
			scmlogic.MWPostItemBalance(),
		).RegisterMWs(
			scmlogic.MWPreItemBalance(),
		),
		s.RegisterModel(new(scmmodel.MovementIn), "movementintransaction"),
		s.RegisterModel(new(scmmodel.MovementInDetail), "movementintransaction/detail"),
		s.RegisterModel(new(scmmodel.MovementInItemBatch), "movementintransaction/batch"),
		s.RegisterModel(new(scmmodel.MovementInItemSerial), "movementintransaction/serial"),
		s.RegisterModel(new(scmmodel.MovementOut), "movementouttransaction"),
		s.RegisterModel(new(scmmodel.MovementOutDetail), "movementouttransaction/detail"),
		s.RegisterModel(new(scmmodel.MovementOutItemBatch), "movementouttransaction/batch"),
		s.RegisterModel(new(scmmodel.MovementOutItemSerial), "movementouttransaction/serial"),
		s.RegisterModel(new(scmmodel.VendorPriceList), "vendor/pricelist"),
		s.RegisterModel(new(scmmodel.VendorPriceListItem), "vendor/pricelistitem"),
		s.RegisterModel(new(scmmodel.Transfer), "transfer"),
		s.RegisterModel(new(scmmodel.TransferIn), "transfer/in"),
		s.RegisterModel(new(scmmodel.TransferOut), "transfer/out"),
		s.RegisterModel(new(scmmodel.TransferDetail), "transfer/detail"),
		s.RegisterModel(new(scmmodel.TransferInBatch), "transfer/in/batch"),
		s.RegisterModel(new(scmmodel.TransferInSerial), "transfer/in/serial"),
		s.RegisterModel(new(scmmodel.TransferOutBatch), "transfer/out/batch"),
		s.RegisterModel(new(scmmodel.TransferOutSerial), "transfer/out/serial"),
		// s.RegisterModel(new(scmmodel.StockOpname), "stockopname"),
		// s.RegisterModel(new(scmmodel.StockOpnameDetail), "stockopname/detail"),
		s.RegisterModel(new(scmmodel.InventoryAdjustment), "inventoryadjustment"),
		s.RegisterModel(new(scmmodel.InventoryAdjustmentDetail), "inventoryadjustment/detail"),
		s.RegisterModel(new(scmmodel.PurchaseRequestJournal), "purchase/request").DisableRoute("gridconfig"),
		s.RegisterModel(new(scmmodel.PurchaseOrderJournal), "purchase/order").DisableRoute("gridconfig"),
		s.RegisterModel(new(scmmodel.ItemRequestJournalType), "item/request/journal/type"),
		s.RegisterModel(new(scmmodel.ItemRequest), "item/request").DisableRoute("gridconfig"),
		s.RegisterModel(new(scmmodel.ItemRequestDetail), "item/request/detail"),
		s.RegisterModel(new(scmmodel.GoodReceipt), "good-receipt"),
		s.RegisterModel(new(scmmodel.GoodReceiptDetail), "good-receipt/detail"),
		s.RegisterModel(new(scmmodel.GoodReceiptItemBatch), "good-receipt/batch"),
		s.RegisterModel(new(scmmodel.GoodReceiptItemSerial), "good-receipt/serial"),
		s.RegisterModel(new(scmmodel.GoodIssue), "good-issue"),
		s.RegisterModel(new(scmmodel.GoodIssueDetail), "good-issue/detail"),
		s.RegisterModel(new(scmmodel.GoodIssueItemBatch), "good-issue/batch"),
		s.RegisterModel(new(scmmodel.GoodIssueItemSerial), "good-issue/serial"),
		s.RegisterModel(new(scmmodel.ItemMinMax), "item/min-max"),
		s.RegisterModel(new(scmmodel.InventJournal), "inventory/journal").DisableRoute("gridconfig"),
		s.RegisterModel(new(scmmodel.InventJournalType), "inventory/journal/type"),
		s.RegisterModel(new(scmmodel.InventTrx), "inventory/transaction"),
		s.RegisterModel(new(scmmodel.InventReceiveIssueJournal), "inventory/receive").DisableRoute("gridconfig"),
		s.RegisterModel(new(scmmodel.PurchaseRequestJournalType), "purchase/request/journal/type"),
		s.RegisterModel(new(scmmodel.PurchaseOrderJournalType), "purchase/order/journal/type"),
		s.RegisterModel(new(scmmodel.AssetAcquisitionJournal), "asset-acquisition"),
		s.RegisterModel(new(scmmodel.AssetAcquisitionJournalType), "asset-acquisition/journal/type"),
		s.RegisterModel(new(scmmodel.StockOpnameJournal), "stock-opname"),
	)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(scmlogic.MovementInEngine), "movementintransaction"),
		s.RegisterModel(new(scmlogic.MovementInDetailEngine), "movementintransaction/detail"),
		s.RegisterModel(new(scmlogic.MovementInBatchEngine), "movementintransaction/batch"),
		s.RegisterModel(new(scmlogic.MovementInSerialEngine), "movementintransaction/serial"),
		s.RegisterModel(new(scmlogic.MovementOutEngine), "movementouttransaction"),
		s.RegisterModel(new(scmlogic.MovementOutDetailEngine), "movementouttransaction/detail"),
		s.RegisterModel(new(scmlogic.MovementOutBatchEngine), "movementouttransaction/batch"),
		s.RegisterModel(new(scmlogic.MovementOutSerialEngine), "movementouttransaction/serial"),
		s.RegisterModel(new(scmlogic.TransferEngine), "transfer"),
		s.RegisterModel(new(scmlogic.TransferDetailEngine), "transfer/detail"),
		s.RegisterModel(new(scmlogic.TransferInBatchEngine), "transfer/in/batch"),
		s.RegisterModel(new(scmlogic.TransferInSerialEngine), "transfer/in/serial"),
		s.RegisterModel(new(scmlogic.TransferOutBatchEngine), "transfer/out/batch"),
		s.RegisterModel(new(scmlogic.TransferOutSerialEngine), "transfer/out/serial"),
		s.RegisterModel(new(scmlogic.StockOpnameEngine), "stock-opname"),
		s.RegisterModel(new(scmlogic.InventoryAdjustmentEngine), "inventoryadjustment"),
		s.RegisterModel(new(scmlogic.ItemRequestDetailEngine), "item/request/detail"),
		s.RegisterModel(new(scmlogic.GoodReceiptEngine), "good-receipt"),
		s.RegisterModel(new(scmlogic.GoodReceiptBatchEngine), "good-receipt/batch"),
		s.RegisterModel(new(scmlogic.GoodReceiptSerialEngine), "good-receipt/serial"),
		s.RegisterModel(new(scmlogic.GoodIssueEngine), "good-issue"),
		s.RegisterModel(new(scmlogic.GoodIssueBatchEngine), "good-issue/batch"),
		s.RegisterModel(new(scmlogic.GoodIssueSerialEngine), "good-issue/serial"),
		// s.RegisterModel(new(scmlogic.PostingProfileHandler), "postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
		// 	Secret:           appConfig.Data.GetString("jwt_secret"),
		// 	GetSessionMethod: "NATS",
		// 	GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		// })),
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "new/postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
		// s.RegisterModel(new(scmlogic.InventTrxEngine), "inventory/trx").RegisterPostMWs(
		// 	scmlogic.MWPostInventTrx(),
		// ),
		s.RegisterModel(new(scmlogic.InventTrxEngine), "inventory/trx").RegisterMWs(scmlogic.MWPreAssignCompanyID()).RegisterPostMWs(
			scmlogic.MWPostInventTrx(),
		).DisableRoute("gets-filter"),
		s.RegisterModel(new(scmlogic.InventTrxEngine), "inventory/trx").RegisterMWs(scmlogic.MWPreAssignCompanyID()).AllowOnlyRoute("gets-filter"),
		s.RegisterModel(new(scmlogic.InventReceiveJournalEngine), "inventory/receive").DisableRoute("gridconfig", "gets-v1"),
		s.RegisterModel(new(scmlogic.ItemEngine), "item").RegisterMWs(scmlogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(scmlogic.ItemBalanceEngine), "item/balance"),
		s.RegisterModel(new(scmlogic.PurchaseRequestEngine), "purchase/request").DisableRoute("gets-v1"),
		s.RegisterModel(new(scmlogic.PurchaseRequestEngine), "purchase/request").AllowOnlyRoute("gets-v1").RegisterMWs(
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			rbaclogic.MWRbacFilterDim("", "jwt"),
			scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
		).RegisterPostMWs(
			scmlogic.MWPostPurchaseRequestGets(),
			scmlogic.MWPostPurchaseRequestGet(),
		),
		// s.RegisterModel(new(scmlogic.InventReceiveJournalEngine), "inventory/receive").DisableRoute("gets-v1"),
		s.RegisterModel(new(scmlogic.InventReceiveJournalEngine), "inventory/receive").AllowOnlyRoute("gets-v1").RegisterMWs(
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			scmlogic.MWPreAssignCompanyID(),
		).RegisterPostMWs(
			scmlogic.MWPostInventoryReceive(),
		),

		s.RegisterModel(new(scmlogic.AssetAcquisitionEngine), "asset-acquisition"),
		s.RegisterModel(new(scmlogic.AssetEngine), "asset"),
		s.RegisterModel(new(scmlogic.PurchaseEngine), "purchase"),
		s.RegisterModel(new(scmlogic.ApprovalAggregatorHandler), "approvalaggregator"),
		s.RegisterModel(new(scmlogic.PurchaseOrderEngine), "purchase/order").DisableRoute("gets-v1"),
		s.RegisterModel(new(scmlogic.PurchaseOrderEngine), "purchase/order").AllowOnlyRoute("gets-v1").RegisterMWs(
			kamis.JWT(kamis.JWTSetupOptions{
				Secret:           appConfig.Data.GetString("jwt_secret"),
				GetSessionMethod: "NATS",
				GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
			}),
			rbaclogic.MWRbacFilterDim("", "jwt"),
			scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
		).RegisterPostMWs(
			scmlogic.MWPostPurchaseOrderGets(),
		),
		s.RegisterModel(new(scmlogic.ItemRequestEngine), "item/request").
			RegisterMWs(
				kamis.JWT(kamis.JWTSetupOptions{
					Secret:           appConfig.Data.GetString("jwt_secret"),
					GetSessionMethod: "NATS",
					GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
				}),
				rbaclogic.MWRbacFilterDim("", "jwt"),
				scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
			).
			RegisterPostMWs(
				scmlogic.MWPostItemRequest(),
			),
		s.RegisterModel(new(scmlogic.ItemSpecEngine), "item/spec"),
		s.RegisterModel(new(scmlogic.InventJournalEngine), "inventory/journal").
			RegisterMWs(
				kamis.JWT(kamis.JWTSetupOptions{
					Secret:           appConfig.Data.GetString("jwt_secret"),
					GetSessionMethod: "NATS",
					GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
				}),
				rbaclogic.MWRbacFilterDim("", "jwt"),
				scmlogic.MWPreSiteHONoFilter(), // for User with HO Site only
			).
			RegisterPostMWs(
				scmlogic.MWPostInventoryJournal(),
			),
		s.RegisterModel(new(scmlogic.InventoryAdjustmentDetailEngine), "inventoryadjustment/detail"),
	)

	//-- custom api for ui
	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).
		RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})).
		Apply(
			s.RegisterModel(new(scmmodel.InventDimension), "movementtransaction/inventorydimension").AllowOnlyRoute("formconfig"),
			s.RegisterModel(new(scmmodel.InventDimension), "inventory/dimension"),
			s.RegisterModel(new(scmmodel.InventJournalLine), "inventory/journal/line"),
			s.RegisterModel(new(scmmodel.PurchaseJournalLine), "purchase/line"),
			s.RegisterModel(new(scmmodel.StockOpnameJournalLine), "stock-opname/line"),
			s.RegisterModel(new(scmmodel.BatchSN), "inventory/journal/batchserial"),
			s.RegisterModel(new(scmmodel.InventReceiveIssueLine), "inventory/receive/line").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.ItemRequestDetailLine), "item/request/detail/line"),
			s.RegisterModel(new(scmmodel.AssetItemTransfer), "asset-acquisition/item/transfer").AllowOnlyRoute("formconfig", "gridconfig"),
			s.RegisterModel(new(scmmodel.AssetRegister), "asset-acquisition/asset/register").AllowOnlyRoute("formconfig", "gridconfig"),
			s.RegisterModel(new(scmmodel.InventTrxReceipt), "inventory/trx/receipt").AllowOnlyRoute("formconfig", "gridconfig"),
			s.RegisterModel(new(scmmodel.InventTrxPerDimension), "inventory/trx/dimension").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.ItemRequestGridUI), "item/request").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.PurchaseOrderJournalGrid), "purchase/order").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.PurchaseRequestJournalGrid), "purchase/request").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.InventJournal), "inventory/journal").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.InventJournalMIMO), "inventory/journalmimo").AllowOnlyRoute("formconfig"),
			s.RegisterModel(new(scmmodel.InventReceiveIssueJournalGrid), "inventory/receive").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.ItemBalanceGrid), "item/balance").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmmodel.ItemBalanceWithName), "item/balance/dimension").AllowOnlyRoute("gridconfig"),
			s.RegisterModel(new(scmlogic.DisplayPrevPo), "display/previous/log").AllowOnlyRoute("gridconfig"),
		)

	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "new/postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})).AllowOnlyRoute("post"),
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})).AllowOnlyRoute("post"),
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "new/postingprofile").AllowOnlyRoute("map-source-data-to-url"),
		s.RegisterModel(new(scmlogic.PostingProfileHandlerV2), "postingprofile").AllowOnlyRoute("map-source-data-to-url"),
		s.RegisterModel(new(scmlogic.ItemBalanceEngine), "item/balance").AllowOnlyRoute("get-qty"),
		s.RegisterModel(new(scmlogic.InventReceiveJournalEngine), "inventory/receive").AllowOnlyRoute("check-gr-used"),
		s.RegisterModel(new(scmlogic.ItemEngine), "item").AllowOnlyRoute("usage-check"),
		s.RegisterModel(new(scmlogic.ItemSpecEngine), "item/spec").AllowOnlyRoute("usage-check"),
	)

	// model nats api
	s.Group().SetMod(modDB).SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(scmmodel.ItemRequest), "item/request").AllowOnlyRoute("insert").RegisterMWs(
			// scmlogic.MWPreAssignSequenceNoForNATSApi("ItemRequest", "_id"),
			scmlogic.MWPreAssignCompanyID(),
		),
		s.RegisterModel(new(scmmodel.ItemRequestDetail), "item/request/detail").AllowOnlyRoute("insert"),
		s.RegisterModel(new(scmmodel.PurchaseRequestJournal), "purchase/request").AllowOnlyRoute("insert").RegisterMWs(
			scmlogic.MWPreAssignSequenceNoForNATSApi("PurchaseRequest", "_id"),
			scmlogic.MWPreAssignCompanyID(),
		),
	)

	s.RegisterModel(new(scmlogic.Lab), "lab")
	s.RegisterModel(new(scmlogic.PreviewLogic), "preview")

	return nil
}
