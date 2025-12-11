package scmlogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
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
					MWPreJournal,
				),
			s.RegisterModel(new(scmmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(publicFn...),
			s.RegisterModel(new(scmmodel.InventReceiveIssueJournal), "inventreceiveissuejournal").AllowOnlyRoute(restrictedFn...).
				RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("InventReceiceJournal", true, "_id"),
					MWPreReceieveIssueJournal,
				),
			s.RegisterModel(new(scmmodel.PurchaseRequestJournal), "purchase/request"),
			s.RegisterModel(new(scmmodel.PurchaseOrderJournal), "purchase/order"),
			s.RegisterModel(new(scmmodel.StockOpnameJournal), "stock-opname"),
			s.RegisterModel(new(scmmodel.AssetAcquisitionJournal), "asset-acquisition"),
		)

	s.RegisterModel(new(PostingProfileHandler), "postingprofile")
	s.RegisterModel(new(StockOpnameEngine), "stock-opname")
	s.RegisterModel(new(PostingProfileHandlerV2), "new/postingprofile")
	s.RegisterModel(new(PreviewLogic), "preview")

	return nil
}
