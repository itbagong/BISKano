package mfglogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/kanoteknologi/knats"
)

func RegisterJournalAPI(s *kaos.Service, mods ...kaos.Mod) error {
	publicFn := []string{"get", "gets", "find"}
	restrictedFn := []string{"insert", "save", "update", "delete"}

	s.Group().RegisterMWs(ficologic.MWPreFilterCompanyID).
		SetMod(mods...).
		Apply(
			s.RegisterModel(new(scmmodel.InventJournal), "inventjournal").AllowOnlyRoute(publicFn...),
			s.RegisterModel(new(scmmodel.InventJournal), "inventjournal").AllowOnlyRoute(restrictedFn...).
				RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventJournal", true, "_id"),
					scmlogic.MWPreJournal,
				),
			s.RegisterModel(new(scmmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(publicFn...),
			s.RegisterModel(new(scmmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(restrictedFn...).
				RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventReceiceJournal", true, "_id"),
					scmlogic.MWPreReceieveIssueJournal,
				),
			// s.RegisterModel(new(scmmodel.PurchaseOrder), "purchaseorder"),
			// s.RegisterModel(new(scmmodel.PurchaseOrderDetail), "purchaseorder/detail"),
			s.RegisterModel(new(mfgmodel.WorkOrderJournal), "work-order-journal"),

			// Work Order Plan
			s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").DisableRoute("save", "gets"),
			s.RegisterModel(new(mfgmodel.WorkOrderPlan), "workorderplan").AllowOnlyRoute("gets").RegisterPostMWs(MWPostWorkOrderPlanGets()),
			s.RegisterModel(new(mfgmodel.WorkOrderPlanReport), "workorderplan/report"),
			s.RegisterModel(new(mfgmodel.WorkOrderPlanReportConsumption), "workorderplan/report/consumption"),
			s.RegisterModel(new(mfgmodel.WorkOrderPlanReportResource), "workorderplan/report/resource"),
			s.RegisterModel(new(mfgmodel.WorkOrderPlanReportOutput), "workorderplan/report/output"),
			s.RegisterModel(new(mfgmodel.WorkOrderSummaryMaterial), "workorderplan/summary/material"),
			s.RegisterModel(new(mfgmodel.WorkOrderSummaryResource), "workorderplan/summary/resource"),
			s.RegisterModel(new(mfgmodel.WorkOrderSummaryOutput), "workorderplan/summary/output"),
		)

	s.RegisterModel(new(PostingProfileHandler), "postingprofile")

	// custom api
	s.RegisterModel(new(WorkOrderPlanEngine), "workorderplan")
	s.RegisterModel(new(WorkOrderPlanReportEngine), "workorderplan/report")

	return nil
}

func RegisterNATSAPI(s *kaos.Service, mods ...kaos.Mod) error {
	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(PostingProfileHandler), "postingprofile"),
	)

	return nil
}
