package sdplogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/sebarcode/kamis"
)

func checkAccessFn(ctx *kaos.Context, parm interface{}, permission string, accessLevel int) error {
	return nil
	/* use this for production: sebar.CheckAccess(ctx, sebar.CheckAccessRequest{
		Param: parm.(codekit.M), PermissionID: permission, AccessLevel: accessLevel
	})
	*/
}

func RegisterJournalAPI(s *kaos.Service, modUI kaos.Mod, modDB kaos.Mod) error {
	// publicFn := []string{"get", "gets", "find"}
	// restrictedFn := []string{"insert", "save", "update", "delete"}

	// s.Group().RegisterMWs(ficologic.MWPreFilterCompanyID).
	// 	SetMod(mods...).
	// 	Apply(
	// 		s.RegisterModel(new(sdpmodel.InventJournal), "inventjournal").AllowOnlyRoute(publicFn...),
	// 		s.RegisterModel(new(sdpmodel.InventJournal), "inventjournal").AllowOnlyRoute(restrictedFn...).
	// 			RegisterMWs(
	// 				tenantcorelogic.MWPreAssignSequenceNo("InventJournal", true, "_id"),
	// 				MWPreJournal,
	// 			),
	// 		s.RegisterModel(new(sdpmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(publicFn...),
	// 		s.RegisterModel(new(sdpmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(restrictedFn...).
	// 			RegisterMWs(
	// 				tenantcorelogic.MWPreAssignSequenceNo("InventReceiceJournal", true, "_id"),
	// 				MWPreReceieveIssueJournal,
	// 			),
	// 		s.RegisterModel(new(sdpmodel.PurchaseOrder), "purchaseorder"),
	// 		s.RegisterModel(new(sdpmodel.PurchaseOrderDetail), "purchaseorder/detail"),
	// 		s.RegisterModel(new(sdpmodel.StockOpnameJournal), "stock-opname"),
	// 	)

	// s.RegisterModel(new(PostingProfileHandler), "postingprofile")
	// s.RegisterModel(new(StockOpnameEngine), "stock-opname")

	func(models ...*kaos.ServiceModel) {
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"} // for activate Endpoint in suim
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			// for detail

		}

		s.Group().
			SetMod(modDB, modUI).
			AllowOnlyRoute(publicEndPoints...).Apply(publicModels...)

		//-- tenant admin
		s.Group().
			SetMod(modDB).
			RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
				Permission:          "TenantAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute(protectedEndPoints...).
			Apply(models...)

		// global
	}(
		// s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity"),
		// s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder"),
		s.RegisterModel(new(sdpmodel.SalesPriceBook), "salespricebook"),
		s.RegisterModel(new(sdpmodel.SalesOrderJournalType), "salesorderjournaltype"),
		s.RegisterModel(new(sdpmodel.UnitCalendar), "unitcalendar").DisableRoute("formconfig", "gridconfig"),
		s.RegisterModel(new(sdpmodel.DocumentUnitChecklist), "documentunitchecklist").DisableRoute("gridconfig"),
	)

	// for form and grid only
	s.Group().SetMod(modUI).Apply(
		s.RegisterModel(new(sdpmodel.SalesOrderLine), "salesorder/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrderLine), "salesorder/line").AllowOnlyRoute("formconfig"),
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
		s.RegisterModel(new(sdpmodel.MeasuringProject), "measuringproject").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity").AllowOnlyRoute("gridconfig", "formconfig"),
		s.RegisterModel(new(sdpmodel.DocumentUnitChecklistGrid), "documentunitchecklist").AllowOnlyRoute("gridconfig"),
	)

	// not for grid and form
	s.Group().SetMod(modDB).Apply(
		s.RegisterModel(new(sdpmodel.SalesQuotation), "salesquotation").
			AllowOnlyRoute("gets").RegisterPostMWs(tenantcorelogic.MWPostVendorName()), // gets
		s.RegisterModel(new(sdpmodel.SalesQuotation), "salesquotation").DisableRoute("gets"), // insert, update, get, delete
	)

	s.Group().SetMod(modDB).Apply(
		s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity").AllowOnlyRoute("gets", "get", "find"), // gets
		s.RegisterModel(new(sdpmodel.SalesOpportunity), "salesopportunity").DisableRoute("gets", "get", "find").RegisterMWs(MWPreSalesOpportunity()),
		s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder").AllowOnlyRoute("gets", "get", "find"),                                                // gets
		s.RegisterModel(new(sdpmodel.SalesOrder), "salesorder").DisableRoute("gets", "get", "find").RegisterMWs(MWPreSalesOrder()),                   // insert, update, get, delete
		s.RegisterModel(new(sdpmodel.MeasuringProject), "measuringproject").AllowOnlyRoute("gets", "get", "find"),                                    // gets
		s.RegisterModel(new(sdpmodel.MeasuringProject), "measuringproject").DisableRoute("gets", "get", "find").RegisterMWs(MWPreMeasuringProject()), // insert, update, get, delete
	)

	// //-- custom api
	s.Group().Apply(
		s.RegisterModel(new(SalesPriceBookEngine), "salespricebook"),
		s.RegisterModel(new(DocumentUnitChecklistEngine), "documentunitchecklist"),
	)

	return nil
}
