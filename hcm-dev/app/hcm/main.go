package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/hcm/hcmlogic"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
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
	serviceName = "v1/hcm"
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
	// if e := sebar.ConfigHasData(appConfig, "addr_auth_validation", "addr_access_validation"); e != nil {
	// 	s.Log().Error(e.Error())
	// 	os.Exit(1)
	// }

	if err := serde.Serde(appConfig.Data, &hcmlogic.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	hcmlogic.Config.EventHub = ev

	if err := serde.Serde(appConfig.Data, &ficoconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	ficoconfig.Config.EventHub = ev

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
		protectedEndPoints := []string{"insert", "update", "save", "delete", "delete-many"}
		approvalProtectedEndPoints := []string{"insert", "update", "save"}

		//-- public
		// model harus dibuat ulang, agar tidak mereference ke memory object yang sama
		publicModels := make([]*kaos.ServiceModel, len(models))
		tenantModels := make([]*kaos.ServiceModel, len(models))
		for index, modelPtr := range models {
			publicModel := new(kaos.ServiceModel)
			*publicModel = *modelPtr
			s.AddServiceModel(publicModel)
			publicModels[index] = publicModel

			// for detail
			mdl := new(kaos.ServiceModel)
			*mdl = *publicModel
			switch modelPtr.Name {
			case "overtime", "talentdevelopment", "businesstrip", "leavecompensation", "loan",
				"worktermination", "coachingviolation", "talentdevelopmentassesment", "manpowerrequest",
				"contract", "olplotting", "talentdevelopmentsk", "tdc":
				mdl.DisableRoute(approvalProtectedEndPoints...)
			}

			s.AddServiceModel(mdl)
			tenantModels[index] = mdl
		}

		s.Group().
			SetMod(modDB, modUI).
			SetDeployer(hd.DeployerName).
			AllowOnlyRoute(publicEndPoints...).Apply(publicModels...)

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
			Apply(tenantModels...)
	}(
		s.RegisterModel(new(hcmmodel.Overtime), "overtime").DisableRoute("gets", "get"),
		s.RegisterModel(new(hcmmodel.Overtime), "overtime").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsOvertime()),
		s.RegisterModel(new(hcmmodel.Overtime), "overtime").AllowOnlyRoute("get").
			RegisterPostMWs(hcmlogic.MWPostGetOvertime()),
		s.RegisterModel(new(hcmmodel.TalentDevelopment), "talentdevelopment").DisableRoute("gets", "formconfig"),
		s.RegisterModel(new(hcmmodel.TalentDevelopment), "talentdevelopment").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsTalentDevelopment()),
		s.RegisterModel(new(hcmmodel.BusinessTrip), "businesstrip").DisableRoute("gets", "get"),
		s.RegisterModel(new(hcmmodel.BusinessTrip), "businesstrip").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsBusinessTrip()),
		s.RegisterModel(new(hcmmodel.BusinessTrip), "businesstrip").AllowOnlyRoute("get").
			RegisterPostMWs(hcmlogic.MWPostGetBusinessTrip()),
		s.RegisterModel(new(hcmmodel.LeaveCompensation), "leavecompensation").DisableRoute("gets", "formconfig"),
		s.RegisterModel(new(hcmmodel.LeaveCompensation), "leavecompensation").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsLeaveCompensation()),
		s.RegisterModel(new(hcmmodel.Loan), "loan").DisableRoute("gets", "formconfig"),
		s.RegisterModel(new(hcmmodel.Loan), "loan").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsLoan()),
		s.RegisterModel(new(hcmmodel.WorkTermination), "worktermination").DisableRoute("gets", "formconfig"),
		s.RegisterModel(new(hcmmodel.WorkTermination), "worktermination").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsWorkTermination()),
		s.RegisterModel(new(hcmmodel.CoachingViolation), "coachingviolation").DisableRoute("gets", "formconfig"),
		s.RegisterModel(new(hcmmodel.CoachingViolation), "coachingviolation").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsCoachingViolation()),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentDetail), "talentdevelopmentdetail"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentAssesment), "talentdevelopmentassesment"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentSK), "talentdevelopmentsk"),
		s.RegisterModel(new(hcmmodel.ManpowerRequest), "manpowerrequest").DisableRoute("gets"),
		s.RegisterModel(new(hcmmodel.ManpowerRequest), "manpowerrequest").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsManpower()),
		s.RegisterModel(new(hcmmodel.MultipleChoiceAnswerHistory), "multiplechoiceanswer"),
		s.RegisterModel(new(hcmmodel.Contract), "contract"),
		s.RegisterModel(new(hcmmodel.Screening), "screening"),
		s.RegisterModel(new(hcmmodel.PsychologicalTest), "psychologicaltest"),
		s.RegisterModel(new(hcmmodel.Interview), "interview"),
		s.RegisterModel(new(hcmmodel.OLPlotting), "olplotting"),
		s.RegisterModel(new(hcmmodel.TechnicalInterview), "techinal-interview"),
		s.RegisterModel(new(hcmmodel.PKWTT), "pkwtt"),
		s.RegisterModel(new(hcmmodel.MCU), "mcu"),
		s.RegisterModel(new(hcmmodel.Training), "training").DisableRoute("gets"),
		s.RegisterModel(new(hcmmodel.Training), "training").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsTraining()),
		s.RegisterModel(new(hcmmodel.OnBoarding), "onboarding"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopment), "tdc").DisableRoute("gets"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopment), "tdc").AllowOnlyRoute("gets").
			RegisterMWs(func(ctx *kaos.Context, i interface{}) (bool, error) {
				param := i.(*dbflex.QueryParam)
				if param.Where != nil {
					if len(param.Where.Items) > 0 {
						if len(param.Where.Items[0].Items) > 0 {
							// get training title
							h := sebar.GetTenantDBFromContext(ctx)
							if h == nil {
								return true, nil
							}

							titles := make([]hcmmodel.TrainingDevelopmentTitle, 0)
							err := h.Gets(new(hcmmodel.TrainingDevelopmentTitle), dbflex.NewQueryParam().SetWhere(
								dbflex.Contains("Name", param.Where.Items[0].Items[1].Value.([]interface{})[0].(string)),
							), &titles)
							if err != nil {
								return true, nil
							}
							fmt.Println(titles)
							if len(titles) > 0 {
								ids := lo.Map(titles, func(t hcmmodel.TrainingDevelopmentTitle, _ int) interface{} {
									return t.ID
								})

								param.Where.Items[0].Items[1].Value = ids
								i = param
							}
						}
					}
				}
				return true, nil
			}).
			RegisterPostMWs(hcmlogic.MWPostGetsTrainingDevelopment()),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentDetail), "tdcdetail"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentParticipant), "tdcparticipant").DisableRoute("gets"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentParticipant), "tdcparticipant").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsTrainingDevelopmentParticipant()),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentAttendance), "tdcattendance").DisableRoute("gets"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentAttendance), "tdcattendance").AllowOnlyRoute("gets").
			RegisterPostMWs(hcmlogic.MWPostGetsTrainingDevelopmentAttendance()),
		s.RegisterModel(new(hcmmodel.JournalType), "journaltype"),
		s.RegisterModel(new(hcmmodel.TDCJournalType), "tdcjournaltype"),
		s.RegisterModel(new(hcmmodel.TestSchedule), "testschedule"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentPracticeTestStaff), "tdcpracticestaff"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentPracticeScore), "tdcpracticescore"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentPracticeDuration), "tdcpracticeduration").DisableRoute("save"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentTitle), "tdctitle"),
	)

	s.Group().
		SetMod(modDB).
		SetDeployer(hd.DeployerName).
		AllowOnlyRoute("insert", "save", "update").
		Apply(
			s.RegisterModel(new(hcmmodel.Overtime), "overtime").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("Overtime", false, ""),
			),
			s.RegisterModel(new(hcmmodel.TalentDevelopment), "talentdevelopment").RegisterMWs(
				hcmlogic.MWPreAssignSequenceNo(),
			),
			s.RegisterModel(new(hcmmodel.BusinessTrip), "businesstrip").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("BusinessTrip", false, ""),
			),
			s.RegisterModel(new(hcmmodel.LeaveCompensation), "leavecompensation").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("LeaveCompensation", false, ""),
			),
			s.RegisterModel(new(hcmmodel.WorkTermination), "worktermination").RegisterMWs(
				hcmlogic.MWPreAssignSequenceNo(),
			).DisableRoute("insert"),
			s.RegisterModel(new(hcmmodel.CoachingViolation), "coachingviolation").RegisterMWs(
				hcmlogic.MWPreAssignSequenceNo(),
			),
			s.RegisterModel(new(hcmmodel.TalentDevelopmentAssesment), "talentdevelopmentassesment").RegisterMWs(
				hcmlogic.MWPreAssignSequenceNo(),
			),
			s.RegisterModel(new(hcmmodel.ManpowerRequest), "manpowerrequest").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("ManpowerRequest", false, ""),
			),
			s.RegisterModel(new(hcmmodel.OLPlotting), "olplotting").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("OLPlotting", false, ""),
			),
			s.RegisterModel(new(hcmmodel.Loan), "loan").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("Loan", false, ""),
			),
			s.RegisterModel(new(hcmmodel.Contract), "contract").RegisterMWs(
				tenantcorelogic.MWPreAssignSequenceNo("Contract", false, ""),
			),
			s.RegisterModel(new(hcmmodel.TalentDevelopmentSK), "talentdevelopmentsk").RegisterMWs(
				hcmlogic.MWPreAssignSequenceNo(),
			),
			s.RegisterModel(new(hcmmodel.TrainingDevelopment), "tdc").RegisterMWs(
				func(ctx *kaos.Context, i interface{}) (bool, error) {
					h := sebar.GetTenantDBFromContext(ctx)
					if h == nil {
						return false, errors.New("missing: connection")
					}

					now := time.Now()
					sequenceKind := fmt.Sprintf("Training%d%02d", now.Year(), int(now.Month()))
					err := h.Get(new(tenantcoremodel.NumberSequence))
					if err == io.EOF {
						seq := new(tenantcoremodel.NumberSequence)
						seq.ID = sequenceKind
						seq.Name = sequenceKind
						seq.OutFormat = "${counter:%03d}/${roman_month}/${dt:2006}/${tdc_title}"
						seq.LastNo = 0
						h.Save(seq)

						coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
						if coID == "DEMO" || coID == "" {
							return false, errors.New("missing: Company, please relogin")
						}
						seqSetup := new(tenantcoremodel.NumberSequenceSetup)
						seqSetup.ID = sequenceKind
						seqSetup.CompanyID = coID
						seqSetup.Kind = sequenceKind
						seqSetup.Label = sequenceKind
						seqSetup.NumSeqID = sequenceKind
						h.Save(seqSetup)
					}
					tenantcorelogic.MWPreAssignCustomSequenceNo(sequenceKind)(ctx, i)
					return true, nil
				},
			),
		)

	//-- custom api
	s.Group().SetDeployer(hd.DeployerName).Apply(
		// s.RegisterModel(new(hcmlogic.OvertimeEngine), "overtime"),
		s.RegisterModel(new(hcmlogic.TalentDevelopmentHandler), "talentdevelopment").DisableRoute("formconfig", "gridconfig"),
		s.RegisterModel(new(hcmlogic.AuthHandler), "auth"),
		s.RegisterModel(new(hcmlogic.TrackingHandler), "tracking"),
		s.RegisterModel(new(hcmlogic.TrainingDevelopmentHandler), "tdc"),
		s.RegisterModel(new(hcmlogic.PsychologicalTestHandler), "psychologicaltest"),
		s.RegisterModel(new(hcmlogic.TestScheduleHandler), "testschedule"),
		s.RegisterModel(new(hcmlogic.ManpowerHandler), "manpowerrequest"),
		s.RegisterModel(new(hcmlogic.ContractHandler), "contract"),
		s.RegisterModel(new(hcmlogic.PostingProfileHandler), "postingprofile"),
		s.RegisterModel(new(hcmlogic.LoanHandler), "loan"),
		s.RegisterModel(new(hcmlogic.WorkTerminationHandler), "worktermination"),
		s.RegisterModel(new(hcmlogic.TDCPracticeDurationHandler), "tdcpracticeduration"),
	)

	//-- custom api for ui
	s.Group().SetMod(modUI).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(hcmmodel.OvertimeLine), "overtime/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(hcmmodel.BusinessTripLineDetail), "businesstrip/line/detail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(hcmmodel.BusinessTripLine), "businesstrip/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(hcmmodel.BusinessTripLine), "businesstrip/line").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.BusinessTripForm), "businesstrip/form").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentForm), "talentdevelopment").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.WorkTerminationForm), "worktermination").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.CoachingViolationForm), "coachingviolation").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.LeaveCompensationForm), "leavecompensation").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.LoanForm), "loan").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.LoanLine), "loan/line").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(hcmmodel.UserRegisterScreen), "register").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentMapping), "talentdevelopment/mapping").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.AssesmentDetail), "talentdevelopmentassesment/assesment").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.AssesmentInterview), "talentdevelopmentassesment/interview").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentParticipantDetail), "tdcparticipant/detail").AllowOnlyRoute("gridconfig"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentSKActingForm), "talentdevelopmentsk/acting").AllowOnlyRoute("formconfig"),
		s.RegisterModel(new(hcmmodel.TalentDevelopmentSKPermanentForm), "talentdevelopmentsk/permanent").AllowOnlyRoute("formconfig"),
	)

	s.RegisterModel(new(hcmlogic.PreviewLogic), "preview")

	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(hcmmodel.TrainingDevelopmentTitle), "tdctitle").SetMod(modDB).AllowOnlyRoute("get"),
	)

	return nil
}
