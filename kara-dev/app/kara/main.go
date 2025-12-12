package main

import (
	"flag"
	"fmt"
	"os"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karaconfig"
	"git.kanosolution.net/sebar/kara/karalogic"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
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
	version     = "v1"
	serviceName = version + "/kara"
	logger      = sebar.LogWithPrefix(serviceName)
	config      = flag.String("config", "app.yml", "path to config file")
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "370", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	if err := serde.Serde(appConfig.Data, karaconfig.Config); err != nil {
		s.Log().Warningf("serde config: %s", err.Error())
	}
	karaconfig.Config.Ev = ev

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})
	s.RegisterMW(getJWT, "getJWT")
	//s.RegisterMW(kamis.NeedJWT(), "needJWT")

	dbm := sebar.NewDBModFromContext()
	uim := suim.New()

	s.Group().SetMod(uim).SetDeployer(hd.DeployerName).
		Apply(
			s.RegisterModel(new(rbacmodel.User), "/admin/user"),
			s.RegisterModel(new(karamodel.UserProfile), "admin/userprofile"),
			s.RegisterModel(new(karamodel.WorkLocationUser), "admin/worklocationuser"),
			s.RegisterModel(new(karamodel.WorkLocationUser_UserEntry), "worklocationuser"),
		)

	s.Group().SetMod(dbm).SetDeployer(hd.DeployerName).
		Apply(
			s.RegisterModel(new(rbacmodel.User), "/admin/user").
				DisableRoute("gets"),
			s.RegisterModel(new(rbacmodel.User), "/admin/user").
				RegisterMWs(karalogic.MWPreUserProfileFind).
				AllowOnlyRoute("gets"),
			s.RegisterModel(new(karamodel.UserProfile), "admin/userprofile").DisableRoute("gets"),
			s.RegisterModel(new(karamodel.UserProfile), "admin/userprofile").AllowOnlyRoute("gets"),
			s.RegisterModel(new(karamodel.WorkLocationUser), "admin/worklocationuser").DisableRoute("gets"),
			s.RegisterModel(new(karamodel.WorkLocationUser), "admin/worklocationuser").AllowOnlyRoute("gets").
				RegisterPostMW(func(ctx *kaos.Context, i interface{}) (bool, error) {
					ev, _ := ctx.DefaultEvent()
					if ev == nil {
						return false, fmt.Errorf("missing: event hub")
					}
					res := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
					recs := []karamodel.WorkLocationUser_UserEntry{}
					serde.Serde(res.Get("data", &[]karamodel.WorkLocationUser{}).(*[]karamodel.WorkLocationUser), &recs)
					user := rbacmodel.User{}
					for idx, rec := range recs {
						getUserReq := &rbaclogic.GetUserByRequest{
							FindBy: "userid",
							FindID: rec.UserID,
						}
						err := ev.Publish(karaconfig.Config.GetUserTopic, getUserReq, &user, nil)
						if err == nil {
							rec.UserID = fmt.Sprintf("%s<br/>%s", user.DisplayName, user.Email)
							rec.UserName = user.DisplayName
							rec.Email = user.Email
						} else {
							rec.UserName = rec.UserID
						}
						recs[idx] = rec
					}
					res.Set("data", recs)
					ctx.Data().Set("FnResult", res)
					return true, nil
				}, "LocationUserGetsPost"),
		)

	s.Group().SetMod(dbm, uim).SetDeployer(hd.DeployerName).
		Apply(
			s.RegisterModel(new(karamodel.WorkLocation), "admin/worklocation"),
			s.RegisterModel(new(karamodel.UserProfile), "profile").
				AllowOnlyRoute("find").
				RegisterPostMW(karalogic.MWProfilePostFind, "mw-post-find-profile").
				AllowOnlyRoute("gets").
				RegisterPostMW(karalogic.MWProfilePostGets, "mw-post-gets-profile"),
			s.RegisterModel(new(karamodel.UserProfile), "admin/profile"),
			s.RegisterModel(new(karamodel.AttendanceRule), "admin/rule"),
			s.RegisterModel(new(karamodel.RuleLine), "admin/ruleline"),
			s.RegisterModel(new(karamodel.LeaveRequest), "leave/request"),
			s.RegisterModel(new(karamodel.LeaveApproval), "leave/approval"),
			s.RegisterModel(new(karamodel.LeaveType), "leave/type"),
		)

	s.Group().SetMod(dbm, uim).
		AllowOnlyRoute("gets", "get", "find", "formconfig", "gridconfig").
		Apply(
			s.RegisterModel(new(karamodel.WorkLocation), "worklocation"),
			s.RegisterModel(new(karamodel.AttendanceRule), "rule"),
			s.RegisterModel(new(karamodel.RuleLine), "ruleline"),
		)

	s.Group().SetMod(dbm, uim).
		Apply(
			s.RegisterModel(new(karamodel.LeaveBalance), "leavebalance").DisableRoute("gets"),
			s.RegisterModel(new(karamodel.LeaveBalance), "leavebalance").
				AllowOnlyRoute("gets").
				RegisterPostMW(karalogic.LeaveBalanceRequestGetsPostMW, "LeaveBalanceRequestPostMW"),
			s.RegisterModel(new(karamodel.LeaveApprovalSetup), "leaveapprovalsetup"),
			s.RegisterModel(new(karamodel.LeaveRequest), "leave").DisableRoute("gets"),
			s.RegisterModel(new(karamodel.LeaveRequest), "leave").
				AllowOnlyRoute("gets").
				RegisterPostMW(karalogic.LeaveRequestGetsPostMW, "LeaveRequestPostMW"),
		)

	hiam, _ := s.HubManager().Get("iam", "")
	admEngine := karalogic.NewAdminEngine(hiam, ev)
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(admEngine, "admin"),
		s.RegisterModel(new(karalogic.UserProfile), "admin/userprofile"),
		s.RegisterModel(new(karalogic.UserTrx), "me/trx"),
		s.RegisterModel(new(karalogic.AdminTrx), "admin/trx"),
		s.RegisterModel(new(karalogic.ShiftAssignment), "shiftassign"),
		s.RegisterModel(new(karalogic.WorkLocationHandler), "admin/worklocation"),
	)

	s.Group().SetMod(dbm, uim).Apply(
		s.RegisterModel(new(karamodel.HolidayProfile), "holiday"),
		s.RegisterModel(new(karamodel.HolidayItem), "holidayitem"),
		s.RegisterModel(new(karamodel.AttendanceTrx), "attendancetrx").AllowOnlyRoute("find"),
		s.RegisterModel(new(karamodel.UserProfile), "profile").DisableRoute("find", "gets"),
	// s.RegisterModel(new(rbacmodel.App), "app"),
	)

	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(karalogic.MeLogic), "me"),
		s.RegisterModel(new(karalogic.LeaveRequestLogic), "leave"),
	)

	s.RegisterModel(new(karamodel.TrxGridView), "trx").SetDeployer(hd.DeployerName).SetMod(uim).AllowOnlyRoute("gridconfig")
	s.RegisterModel(new(karamodel.TrxRequest), "trx/request").SetDeployer(hd.DeployerName).SetMod(uim).AllowOnlyRoute("formconfig")
	s.RegisterModel(new(karamodel.AttendanceTrx), "admin/trx").
		SetDeployer(hd.DeployerName).SetMod(dbm).
		AllowOnlyRoute("gets").
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

	s.RegisterModel(new(karamodel.AttendancePhoto), "photo").SetDeployer(hd.DeployerName).SetMod(dbm).AllowOnlyRoute("get")
	s.RegisterModel(new(karamodel.AttendanceTrx), "admin/trx").SetDeployer(hd.DeployerName).SetMod(dbm).AllowOnlyRoute("get", "save", "delete")

	s.Group().SetDeployer(knats.DeployerName).Apply(
		s.RegisterModel(new(karamodel.AttendanceTrx), "attendancetrx").AllowOnlyRoute("find"),
	)

	return nil
}

func panicIfError(s *kaos.Service, e error) {
	if e != nil {
		s.Log().Errorf("error happened, app will be shutdown. %s", e.Error())
		os.Exit(1)
	}
}
