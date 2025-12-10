package main

import (
	"flag"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/kmsg"
	ksmsg "github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/ariefdarmawan/kmsg/sender/smtper"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/dbmod"
)

var (
	version     = "v1"
	serviceName = version + "/msg"
	logger      = sebar.LogWithPrefix(serviceName)
	config      = flag.String("config", "app.yml", "path to config file")
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	smtpOpts := smtper.Options{
		Server:   appConfig.Data.GetString("smtp_server"),
		Port:     appConfig.Data.GetInt("smtp_port"),
		UID:      appConfig.Data.GetString("smtp_uid"),
		Password: appConfig.Data.GetString("smtp_password"),
		TLS:      appConfig.Data.GetBool("smtp_tls"),
	}
	sender := smtper.NewSender(smtpOpts)

	msgMgr := ksmsg.NewKaosModel()
	msgMgr.RegisterSender(sender, "SMTP")
	s.RegisterModel(msgMgr, "").SetDeployer(knats.DeployerName).AllowOnlyRoute("create", "send-template", "send-by-id")

	modDb := dbmod.New()
	modSuim := suim.New()

	s.RegisterModel(new(kmsg.Template), "template").SetMod(modDb).SetDeployer(hd.DeployerName)
	s.RegisterModel(new(kmsg.Template), "ui/template").SetMod(modSuim).SetDeployer(hd.DeployerName)

	return nil
}
