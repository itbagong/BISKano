package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"xibarCoreApp/xibarconfig"
	"xibarCoreApp/xibarlogic"
	"xibarCoreApp/xibarmodel"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/serde"
	"github.com/ariefdarmawan/suim"
	"github.com/golang-jwt/jwt"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/dbmod"
	"github.com/sebarcode/kamis"
	"github.com/sebarcode/siam"
	"github.com/sebarcode/siam/storage/jsonstore"

	_ "github.com/ariefdarmawan/flexmgo"
)

var (
	version         = "v1"
	serviceName     = version + "/iam"
	logger          = sebar.LogWithPrefix(serviceName)
	config          = flag.String("config", "app.yml", "path to config file")
	sessionLifeTime int
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	// variables from appconfig
	e := sebar.CopyConfigDataToService(appConfig, s,
		"addr_web_user", "addr_web_user_activation", "addr_web_user_2fa",
		"topic_send_message_template", "topic_change_session_data")
	panicIfError(s, e)

	serde.Serde(appConfig.Data, xibarconfig.Config)
	xibarconfig.Config.Ev = ev

	sessionLifeTime = appConfig.Data.GetInt("session_lifetime")
	if sessionLifeTime == 0 {
		sessionLifeTime = 60 * 60 * 24 //24 hour
	}

	// jwt
	getJWT := kamis.JWT(kamis.JWTSetupOptions{
		Secret:           appConfig.Data.GetString("jwt_secret"),
		GetSessionMethod: "NATS",
		DisableExpiry:    true,
		GetSessionTopic:  appConfig.Data.GetString("jwt_validate_topic"),
	})

	// iam dan identity storage setup
	storage := jsonstore.NewStorage(appConfig.Data.GetString("iam_folder"))
	siamOpts := &siam.Options{
		Storage: storage,
	}
	siamMgr := siam.New(logger, sessionLifeTime, siamOpts)
	if e := siamMgr.Load(); e != nil {
		fmt.Println("fail to load iam.", e.Error())
		os.Exit(1)
	}

	// authentication setup
	authOpts := &rbaclogic.AuthOptions{
		SignMethod:     jwt.GetSigningMethod(appConfig.Data.GetString("jwt_method")),
		SignSecret:     appConfig.Data.GetString("jwt_secret"),
		SiamMgr:        siamMgr,
		Logger:         logger,
		SecondLifetime: sessionLifeTime,
		SyncPeriod:     1 * time.Minute,
		Enrich: func(ctx *kaos.Context, user *rbacmodel.User) (codekit.M, codekit.M, error) {
			defaultTenantID := xibarconfig.Config.DefaultTenantID
			sessData := codekit.M{
				"DisplayName": user.DisplayName,
				"Email":       user.Email,
				"Status":      user.Status,
				"Dimension":   user.Dimension,
			}
			jwtData := codekit.M{
				"DisplayName": user.DisplayName,
				"Status":      user.Status,
				"Email":       user.Email,
				"Dimension":   user.Dimension,
			}
			if defaultTenantID != "" {
				sessData.Set("TenantID", defaultTenantID)
				jwtData.Set("TenantID", defaultTenantID)
			}

			return sessData, jwtData, nil
		},
	}

	authEngine, err := rbaclogic.NewAuthEngine(rbaclogic.NewRbac(s.HubManager().GetMust("iam", "")), authOpts)
	if err != nil {
		logger.Errorf("create auth engine error. %s", err.Error())
		os.Exit(1)
	}

	// rbac
	rbacOpts := &rbaclogic.RbacOptions{}
	rbacOpts.PreCreateUserFn = func(ctx *kaos.Context, req *rbaclogic.CreateUserRequest) error {
		if req.User.DisplayName == "" {
			req.User.DisplayName = req.User.Email
		}
		req.User.LoginID = req.User.Email
		req.User.Enable = true
		req.User.Status = "Registered"
		return nil
	}

	s.Group().SetDeployer(hd.DeployerName).
		RegisterMWs(getJWT).
		Apply(
			s.RegisterModel(authEngine, "").DisableRoute("validate"),
			s.RegisterModel(rbaclogic.NewPublicUserEvent(rbacOpts), "user").AllowOnlyRoute("get-by", "find-by", "update-dimension"),
			s.RegisterModel(new(rbaclogic.PublicUserAPI), "user"),
			s.RegisterModel(new(rbaclogic.UserMeLogic), "user").DisableRoute("activate", "resend-activation-email"),
			s.RegisterModel(xibarlogic.NewUserLogic(appConfig, authOpts), "user"),
			s.RegisterModel(new(rbaclogic.TenantLogic), "tenant"),
			s.RegisterModel(new(rbaclogic.AccessLogic), "access"),
			s.RegisterModel(new(rbaclogic.TenantJoinLogic), "tenantjoin"),
			s.RegisterModel(new(rbaclogic.AppLogic), "app"),
		)

	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(rbaclogic.UserMeLogic), "user").AllowOnlyRoute("activate", "resend-activation-email"),
	)

	s.Group().SetDeployer(knats.DeployerName).
		RegisterMWs(getJWT).
		Apply(
			s.RegisterModel(authEngine, "auth").AllowOnlyRoute("validate", "change-data"),
			s.RegisterModel(rbaclogic.NewPublicUserEvent(rbacOpts), "user"),
		)

	registerScreen(s)

	modSuim := suim.New()
	modDb := dbmod.New()

	s.Group().SetMod(modSuim, modDb).SetDeployer(hd.DeployerName).
		RegisterMWs(getJWT).
		Apply(
			s.RegisterModel(&rbacmodel.TenantJoin{}, "tenantjoin").AllowOnlyRoute("formconfig"),
			s.RegisterModel(&rbacmodel.TenantJoinExtended{}, "tenantjoin").
				AllowOnlyRoute("find").
				RegisterMW(kaos.MWFunc(rbaclogic.MWPreTenantJoinRequestFind), "mw-pre-tenantjoin-request-find").
				RegisterPostMW(kaos.MWFunc(rbaclogic.MWPostTenantJoinRequestFind), "mw-post-tenantjoin-request-find"),
			s.RegisterModel(&rbacmodel.TenantJoinExtended{}, "tenantjoin").
				AllowOnlyRoute("find").SetAlias("find", "review").
				RegisterMWs(rbaclogic.MWPreTenantJoinReviewFind).
				RegisterPostMWs(rbaclogic.MWPostTenantJoinReviewFind),
		)

	s.RegisterModel(new(xibarmodel.User), "user").SetMod(modDb).AllowOnlyRoute("find", "get", "gets")

	return func() {
		authEngine.Close()
		siamMgr.Close()
	}
}

func registerScreen(s *kaos.Service) {
	modSuim := suim.New()

	//-- forms
	s.Group().SetMod(modSuim).SetDeployer(hd.DeployerName).AllowOnlyRoute("formconfig").Apply(
		s.RegisterModel(new(rbacmodel.CreateTenantRequest), "/ui/create-tenant"),
		s.RegisterModel(new(rbacmodel.LoginScreen), "/ui/login"),
		s.RegisterModel(new(rbacmodel.UserRegisterScreen), "/ui/register"),
	)
}

func panicIfError(s *kaos.Service, e error) {
	if e != nil {
		s.Log().Errorf("error happened, app will be shutdown. %s", e.Error())
		os.Exit(1)
	}
}
