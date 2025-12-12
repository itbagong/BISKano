package main

import (
	"flag"
	"os"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/sheconfig"
	"git.kanosolution.net/sebar/she/shelogic"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/she"
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

	if err := serde.Serde(appConfig.Data, &shelogic.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	shelogic.Config.EventHub = ev
	shelogic.Config.SeqNumTopic = "/v1/tenant/numseq/claim-by-setup"
	if err := serde.Serde(appConfig.Data, &sheconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	sheconfig.Config.EventHub = ev
	sheconfig.Config.SeqNumTopic = "/v1/tenant/numseq/claim-by-setup"
	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.EventHub = ev

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")

	modDB := sebar.NewDBModFromContext()
	modUI := suim.New()

	func(models ...*kaos.ServiceModel) {
		publicEndPoints := []string{"formconfig", "gridconfig", "listconfig", "get", "gets", "find", "new"} // for activate Endpoint in suim
		protectedSaveEndPoints := []string{"insert", "update", "save"}
		protectedDeleteEndPoints := []string{"delete", "delete-many"}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			// for detail
			switch modelPtr.Name {
			case "coaching":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostCoaching(),
				)
			case "csms":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostCSMS(),
				)
			case "jsa":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostJSA(),
				)
			case "legalcompliance":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostLegalCompliance(),
				)
			case "mastersop":
				publicModels[index].RegisterMWs(
					tenantcorelogic.MWPreEmployeeFilter("Position", "JobPosition"),
				)
			case "mastersopsummary":
				publicModels[index].RegisterMWs(
					func(ctx *kaos.Context, i interface{}) (bool, error) {
						ctx.Data().Set("DBModFilter", []*dbflex.Filter{
							dbflex.Eq("IsActive", true),
						})
						return true, nil
					},
				)
			case "ibprtrx":
				publicModels[index].RegisterMWs(
					tenantcorelogic.MWPreEmployeeFilter("ID", "IBPRTeam"),
				)
			case "safetycard":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostSafetyCard(),
				)
			case "meeting":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostMeeting(),
				)
			case "legalregister":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostLegalRegister(),
				)
			case "investigasi":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostInvestigasi(),
				)
			case "mcutransaction":
				publicModels[index].RegisterPostMWs(
					shelogic.MWPostMCUTransaction(),
				)
			}
		}

		protectedSaveModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			protectedModel := new(kaos.ServiceModel)
			*protectedModel = *modelPtr
			s.AddServiceModel(protectedModel)
			protectedSaveModels[index] = protectedModel

			switch modelPtr.Name {
			case "bulletin":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Bulletin", false, "_id"),
				)
			case "coaching":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Coaching", false, "_id"),
				)
			case "mcutransaction":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("MCUTransaction", false, "_id"),
				)
			case "inspection":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Inspection", false, "_id"),
					shelogic.MWPreInspectionAssignDefault,
				)
			case "audit":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("SHEAudit", false, "_id"),
				)
			case "induction":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Induction", false, "_id"),
				)
			case "jsa":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("JSA", false, "_id"),
				)
			case "investigasi":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Investigation", false, "_id"),
				)
			case "csms":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("CSMS", false, "_id"),
				)
			case "p3k":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("P3K", false, "_id"),
				)
			case "sidak":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Sidak", false, "_id"),
				)
			case "mcumasterpackage":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("MCUPackage", false, "_id"),
				)
			case "observasi":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("Observation", false, "_id"),
				)
			case "masteribpr":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("MasterIBPR", false, "_id"),
				)
			case "masterrsca":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("MasterRSCA", false, "_id"),
				)
			case "rscatrx":
				protectedSaveModels[index].RegisterMWs(
					tenantcorelogic.MWPreAssignSequenceNo("RSCATransaction", false, "_id"),
				)
			}
		}

		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
			AllowOnlyRoute(publicEndPoints...).
			Apply(publicModels...)

		//--- save, inset, update
		s.Group().
			SetMod(modDB).
			SetDeployer(hd.DeployerName).
			RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
				Permission:          "TenantAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute(protectedSaveEndPoints...).
			Apply(protectedSaveModels...)

		//--- delete, delete-many
		s.Group().
			SetMod(modDB).
			SetDeployer(hd.DeployerName).
			RegisterMWs(kamis.NeedAccess(kamis.NeedAccessOptions{
				Permission:          "TenantAdmin",
				RequiredAccessLevel: 7,
				CheckFunction:       checkAccessFn,
			})).
			AllowOnlyRoute(protectedDeleteEndPoints...).
			Apply(models...)

	}(
		s.RegisterModel(new(shemodel.Pica), "pica"),
		s.RegisterModel(new(shemodel.WistleBlower), "wistleblower"),
		s.RegisterModel(new(shemodel.SafetyCard), "safetycard").DisableRoute("save"),
		s.RegisterModel(new(shemodel.Coaching), "coaching"),
		s.RegisterModel(new(shemodel.Meeting), "meeting").DisableRoute("save"),
		s.RegisterModel(new(shemodel.Sidak), "sidak").DisableRoute("gets"),
		s.RegisterModel(new(shemodel.Bulletin), "bulletin"),
		s.RegisterModel(new(shemodel.LegalRegister), "legalregister").DisableRoute("save"),
		s.RegisterModel(new(shemodel.LegalRegisterDetail), "legalregisterdetail").DisableRoute("save", "gets", "delete"),
		s.RegisterModel(new(shemodel.LegalCompliance), "legalcompliance"),
		s.RegisterModel(new(shemodel.MCUItemTemplate), "mcuitemtemplate"),
		s.RegisterModel(new(shemodel.MCUMasterPackage), "mcumasterpackage"),
		s.RegisterModel(new(shemodel.MCUTransaction), "mcutransaction"),
		s.RegisterModel(new(shemodel.MasterSOP), "mastersop").DisableRoute("save"),
		s.RegisterModel(new(shemodel.MasterSOPSummary), "mastersopsummary"),
		s.RegisterModel(new(shemodel.IBPR), "masteribpr"),
		s.RegisterModel(new(shemodel.IBPRTransaction), "ibprtrx").DisableRoute("save"),
		s.RegisterModel(new(shemodel.RSCA), "masterrsca"),
		s.RegisterModel(new(shemodel.RSCATransaction), "rscatrx"),
		s.RegisterModel(new(shemodel.Inspection), "inspection"),
		s.RegisterModel(new(shemodel.Csms), "csms"),
		s.RegisterModel(new(shemodel.Observasi), "observasi"),
		s.RegisterModel(new(shemodel.Audit), "audit"),
		s.RegisterModel(new(shemodel.Jsa), "jsa"),
		s.RegisterModel(new(shemodel.P3k), "p3k"),
		s.RegisterModel(new(shemodel.Induction), "induction"),
		s.RegisterModel(new(shemodel.Investigasi), "investigasi"),
	)

	shelogic.RegisterLab(s)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(shelogic.PicaLogic), "pica"),
		s.RegisterModel(new(shelogic.SCLogic), "safetycard"),
		s.RegisterModel(new(shelogic.MeetingLogic), "meeting"),
		s.RegisterModel(new(shelogic.SidakLogic), "sidak"),
		s.RegisterModel(new(shelogic.LRLogic), "legalregister"),
		s.RegisterModel(new(shelogic.LRDLogic), "legalregisterdetail"),
		s.RegisterModel(new(shelogic.MCUTransactionLogic), "mcutransaction"),
		s.RegisterModel(new(shelogic.MasterSOPLogic), "mastersop"),
		s.RegisterModel(new(shelogic.IBPRTransactionLogic), "ibprtrx"),
		s.RegisterModel(new(shelogic.PostingProfileHandler), "postingprofile"),
	)

	//-- custom api for ui
	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(shemodel.Fatigue), "sidak/fatigue").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.SpeedGun), "sidak/speedgun").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.Alcohol), "sidak/alcohol").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.Drug), "sidak/drug").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.MeetingResult), "meeting/results").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.MCUItemTemplateLine), "mcuitemtemplate/lines").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.AssessmentResult), "mcutransaction/assesment").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.MCUResult), "mcutransaction/result").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.MCUResultDetailFrom), "mcutransaction/resultdetail").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.MCUFollowUp), "mcutransaction/followup").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.IBPRLine), "masteribpr/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.GridConfigInitialRisk), "ibprtrx/initialrisk").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.CurrentAction), "ibprtrx/currentaction").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.GridConfigResidualRisk), "ibprtrx/residualrisk").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.GridConfigOpportunityAssessment), "ibprtrx/opportunity").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.RSCALine), "masterrsca/lines").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.RSCATrxLine), "rscatrx/lines").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.Instructions), "mcuitemtemplate/instruction").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.SMK3), "audit/smk3").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.JsaLine), "jsa/lines").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.MedLine), "p3k/medicines").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.P3kDetail), "p3k/detail").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.SMKP), "audit/smkp").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.SMKPAU), "audit/smkpau").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.InductionAttendee), "induction/attendee").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.InductionMaterial), "induction/materials").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.InductionAssessment), "induction/assesment").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.DetailsAccident), "investigasi/detailsaccident").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.Involvement), "investigasi/involvement").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.InvestigationTeam), "investigasi/investigationteam").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(shemodel.DirectCauseLine), "investigasi/directcause").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.BasicCauseLine), "investigasi/basiccause").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.RiskReductionLine), "investigasi/riskreduction").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.ExternalReportLine), "investigasi/externalreport").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.PICA), "investigasi/pica").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.AccidentTypeLine), "investigasi/accidenttypeline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.PersonLine), "investigasi/personline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.MedicalTreatmentLine), "investigasi/medicaltreatmentline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.AssetLine), "investigasi/assetline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.EnvironmentLine), "investigasi/environmentline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.TeamMemberLine), "investigasi/teammemberline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.CheckListDirectionLine), "investigasi/checklistdirectionline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.InjuredLine), "investigasi/injuredline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.PartEquipment), "investigasi/partequipmentline").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(shemodel.ItemPica), "pica/itempica").AllowOnlyRoute("formconfig", "gridconfig"),
	)

	return nil
}
