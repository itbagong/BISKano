package main

import (
	"flag"
	"os"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/sdp/sdpconfig"
	"git.kanosolution.net/sebar/sdp/sdplogic"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
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
	serviceName = "v1/sdp"
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
	if e := sebar.ConfigHasData(appConfig, "addr_auth_validation", "addr_access_validation"); e != nil {
		s.Log().Error(e.Error())
		os.Exit(1)
	}

	if err := serde.Serde(appConfig.Data, &sdpconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	sdpconfig.Config.EventHub = ev
	ficoconfig.Config.EventHub = ev

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
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"} // for activate Endpoint in suim
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}
		// modelName := []string{"salespricebook", "unitcalendar", "salesorder", "salesquotation", "salesopportunity", "measuringproject"}

		rbacEndPoints := []string{"get", "gets", "find"}
		rbacModelList := []string{
			"salesorder",
			"salesopportunity",
			"salesquotation",
			"measuringproject",
			"contract-checklist",
			"unitcalendar",
			// TODO: tambahin yang lain
		}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			switch modelPtr.Name {
			case "RegisterPostMWs":
				publicModels[index].RegisterPostMWs(tenantcorelogic.MWPostVendorName())
			}

			if lo.Contains(rbacModelList, modelPtr.Name) {
				// for public
				newPublicEndPoints, _ := lo.Difference(publicEndPoints, rbacEndPoints)
				publicModels[index].AllowOnlyRoute(newPublicEndPoints...)

				// for limited access public
				rbacModel := new(kaos.ServiceModel)
				*rbacModel = *modelPtr
				s.AddServiceModel(rbacModel)
				rbacModel.AllowOnlyRoute(rbacEndPoints...).RegisterMWs(
					// TODO: tambahkan filter companyID
					rbaclogic.MWRbacFilterDim("", "jwt"),
				)
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
			case "salesorder":
				protectedModels[index].RegisterMWs(
					sdplogic.MWPreSalesOrder(),
				)
			case "salesopportunity":
				protectedModels[index].RegisterMWs(
					sdplogic.MWPreSalesOpportunity(),
				)
			case "measuringproject":
				protectedModels[index].RegisterMWs(
					sdplogic.MWPreMeasuringProject(),
				)
			}
		}
		// public
		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
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

		// global
	}(
		s.RegisterModel(new(sdpmodel.SalesQuotation), "salesquotation").AllowOnlyRoute("gets", "get", "find").DisableRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity").AllowOnlyRoute("gets", "get", "find"),
		s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder").AllowOnlyRoute("gets", "get", "find"),
		s.RegisterModel(new(sdpmodel.MeasuringProject), "measuringproject").AllowOnlyRoute("gets", "get", "find"),
		s.RegisterModel(new(sdpmodel.SalesPriceBook), "salespricebook"),
		s.RegisterModel(new(sdpmodel.SalesOrderJournalType), "salesorderjournaltype"),
		s.RegisterModel(new(sdpmodel.UnitCalendar), "unitcalendar").AllowOnlyRoute("gets", "get", "find").DisableRoute("gridconfig", "formconfig"), //.DisableRoute("formconfig", "gridconfig", "insert", "update"),
		s.RegisterModel(new(sdpmodel.UnitCalendarSite), "unitcalendar/site").AllowOnlyRoute("gets"),
		s.RegisterModel(new(sdpmodel.DocumentUnitChecklist), "documentunitchecklist").DisableRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklist), "contract-checklist").AllowOnlyRoute("gets", "get", "find").DisableRoute("gridconfig", "formconfig"),
	)

	// for form and grid only
	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(sdpmodel.SalesOpportunityGridView), "salesopportunity/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderGrid), "salesorder/grid").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderLineGrid), "salesorder/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderLineForm), "salesorder/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderLinePreviewGrid), "salesorder/line/preview").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderEditorForm), "salesorder/editor").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderBreakdownCost), "salesorder/breakdowncost").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderBreakdownCost), "salesorder/breakdowncost").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderManPower), "salesorder/manpower").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderManPower), "salesorder/manpower").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.LinesOpportunity), "opportunity/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.LinesOpportunity), "opportunity/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.CompetitorOpportunity), "opportunity/competitor").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.CompetitorOpportunity), "opportunity/competitor").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.EventOpportunity), "opportunity/event").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.EventOpportunity), "opportunity/event").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.BondOpportunity), "opportunity/bond").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.BondOpportunity), "opportunity/bond").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationGrid), "salesquotation").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationForm), "salesquotation").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationLineForm), "salesquotation/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationLineGrid), "salesquotation/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationLinePreviewGrid), "salesquotation/line/preview").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesQuotationEditorForm), "salesquotation/editor").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.SalesPriceBookLineGrid), "salespricebook/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesPriceBookLineForm), "salespricebook/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.LinesMeasuringProject), "measuringproject/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.LinesMeasuringProject), "measuringproject/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.JournalTypeContext), "journaltypecontext").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.UnitCalendarForm), "unitcalendar").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.UnitCalendarLineGrid), "unitcalendar/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.DocumentUnitChecklistGrid), "documentunitchecklist").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistGrid), "contract-checklist").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistForm), "contract-checklist").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistGridTab), "contract-checklist/checked").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistBastForm1), "contract-checklist/bast1").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistBastForm2), "contract-checklist/bast2").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistBastForm3), "contract-checklist/bast3").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(sdpmodel.ContractChecklistBastForm5), "contract-checklist/bast5").AllowOnlyRoute("formconfig"),
	)

	s.Group().SetMod(modDB).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(sdpmodel.ContractChecklist), "contract-checklist").DisableRoute("gridconfig", "formconfig", "gets", "get", "find", "new"),
		s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity").DisableRoute("gets", "get", "find", "insert", "update", "new").RegisterMWs(sdplogic.MWPreSalesOpportunity()),
		s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder").DisableRoute("gets", "get", "find", "insert", "update", "new").RegisterMWs(sdplogic.MWPreSalesOrder()), // insert, update, get, delete
		s.RegisterModel(new(sdpmodel.MeasuringProject), "measuringproject").DisableRoute("gets", "get", "find", "new").RegisterMWs(sdplogic.MWPreMeasuringProject()),   // insert, update, get, delete
		s.RegisterModel(new(sdpmodel.SalesQuotation), "salesquotation").DisableRoute("gets", "insert", "update", "new", "find", "get"),                                 // insert, update, get, delete
	)
	// //-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(sdplogic.SalesQuotationEngine), "salesquotation").RegisterMWs(sdplogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(sdplogic.UnitCalendarEngine), "unitcalendar"),
		s.RegisterModel(new(sdplogic.SalesOrderEngine), "salesorder").RegisterMWs(sdplogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(sdplogic.SalesOpportunityEngine), "salesopportunity").RegisterMWs(sdplogic.MWPreAssignCompanyID()),
		s.RegisterModel(new(sdplogic.SalesPriceBookEngine), "salespricebook"),
		s.RegisterModel(new(sdplogic.DocumentUnitChecklistEngine), "documentunitchecklist"),
		s.RegisterModel(new(sdplogic.PostingProfileHandler), "new/postingprofile").RegisterMWs(kamis.JWT(kamis.JWTSetupOptions{
			Secret:           appConfig.Data.GetString("jwt_secret"),
			GetSessionMethod: "NATS",
			GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
		})),
	)

	// for NATS Router
	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(sdplogic.DocumentUnitChecklistEngine), "documentunitchecklist").AllowOnlyRoute("save-from-wo"),
		s.RegisterModel(new(sdplogic.PostingProfileHandler), "postingprofile").AllowOnlyRoute("map-source-data-to-url"),
	)

	return nil
}
