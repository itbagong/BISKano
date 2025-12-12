package main

import (
	"flag"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/flow/flowconfig"
	"git.kanosolution.net/sebar/flow/flowlogic"
	"git.kanosolution.net/sebar/sebar"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
	"github.com/kanoteknologi/hd"
	"github.com/sebarcode/kamis"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/flow"
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
	//modDB := sebar.NewDBModFromContext()
	//modUI := suim.New()

	getJWTWithoutIAMVerification := kamis.JWT(kamis.JWTSetupOptions{
		Secret: appConfig.Data.GetString("jwt_secret"),
	})
	s.RegisterMW(getJWTWithoutIAMVerification, "getJWTWithoutIAMVerification")
	serde.Serde(appConfig.Data, flowconfig.Config)

	s.RegisterModel(new(flowlogic.Request), "request").SetDeployer(hd.DeployerName)
	return nil
}
