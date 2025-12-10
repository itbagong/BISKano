package main

import (
	"flag"
	"os"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/dbmod"
)

var (
	version     = "v1"
	serviceName = version + "/admin"
	logger      = sebar.LogWithPrefix(serviceName)
	config      = flag.String("config", "app.yml", "path to config file")
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	err := sebar.CopyConfigDataToService(appConfig, s, "topic_send_message_template", "addr_web_user")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(-1)
	}

	dbm := dbmod.New()
	uim := suim.New()

	s.Group().SetMod(uim).SetDeployer(hd.DeployerName).
		Apply(
			s.RegisterModel(new(rbacmodel.Tenant), "tenant"),
			s.RegisterModel(new(rbacmodel.User), "user"),
			s.RegisterModel(new(rbacmodel.Role), "role"),
			s.RegisterModel(new(rbacmodel.RoleMember), "rolemember"),
			s.RegisterModel(new(rbacmodel.Feature), "feature"),
			s.RegisterModel(new(rbacmodel.AppMenu), "menu"),
			s.RegisterModel(new(rbacmodel.FeatureCategory), "featurecategory"),
			s.RegisterModel(new(rbacmodel.DimensionItem), "kv"),
		)

	s.Group().SetMod(dbm).
		SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(rbacmodel.User), "user"),
		s.RegisterModel(new(rbacmodel.Role), "role"),
		s.RegisterModel(new(rbacmodel.RoleMember), "rolemember").DisableRoute("gets"),
		s.RegisterModel(new(rbacmodel.RoleMember), "rolemember").AllowOnlyRoute("gets").RegisterPostMWs(rbaclogic.MWPostRoleMemberGets),
		s.RegisterModel(new(rbacmodel.Feature), "feature"),
		s.RegisterModel(new(rbacmodel.AppMenu), "menu"),
		s.RegisterModel(new(rbacmodel.AppMenu), "menu/she").
			AllowOnlyRoute("find").
			RegisterMWs(func(ctx *kaos.Context, i interface{}) (bool, error) {
				ctx.Data().Set("DBModFilter", []*dbflex.Filter{
					dbflex.And(
						dbflex.Eq("AppID", "bagong"),
						dbflex.Eq("Section", "Transaction"),
						dbflex.Contains("Uri", "/she/"),
					),
				})
				return true, nil
			}),
		s.RegisterModel(new(rbacmodel.RoleFeature), "rolefeature"),
		s.RegisterModel(new(rbacmodel.FeatureCategory), "featurecategory"),
		s.RegisterModel(new(rbacmodel.Tenant), "tenant").DisableRoute("gets"),
		s.RegisterModel(new(rbacmodel.Tenant), "tenant").AllowOnlyRoute("gets").RegisterPostMWs(rbaclogic.MWPostTenantGets),
		s.RegisterModel(new(rbacmodel.TenantUser), "tenantuser"),
	)

	s.Group().SetMod(dbm, uim).Apply(
		s.RegisterModel(new(rbacmodel.App), "app"),
		s.RegisterModel(new(rbacmodel.TenantApp), "tenantapp"),
	)

	hiam, _ := s.HubManager().Get("iam", "")
	admEngine := rbaclogic.NewAdminEngine(hiam, ev)
	s.Group().SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(admEngine, ""),
		s.RegisterModel(new(rbaclogic.AdminUserAPI), "user"),
	)
	s.RegisterRoute(func(ctx *kaos.Context, payload *rbaclogic.AddFeaturesToRoleRequest) (string, error) {
		return admEngine.AddFeaturesToRole(payload)
	}, "add-features-to-role")

	s.RegisterRoute(func(ctx *kaos.Context, payload *rbacmodel.RoleMember) (string, error) {
		return admEngine.AddUserToRole(payload)
	}, "add-user-to-role")

	s.Group().SetMod(dbm).SetDeployer(knats.DeployerName).Apply(
		//s.RegisterModel(new(logic.JobLogic), "job"),
		//s.RegisterModel(new(logic.AdminRefresher), "refresher"),
		s.RegisterModel(new(rbacmodel.AppMenu), "menu").AllowOnlyRoute("find"),
	)

	return nil
}
