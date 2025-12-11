package main

import (
	"flag"
	"os"

	"git.kanosolution.net/sebar/bagong/bagongconfig"
	"git.kanosolution.net/sebar/bagong/bagonglogic"
	"git.kanosolution.net/sebar/sebar"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/serde"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/bagong-job"
	logger      = sebar.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()

	appConfig, hm, ev, err := sebar.ConfigToApp(*config, "370")
	if err != nil {
		logger.Errorf("fail to prepare app. %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		ev.Close()
		hm.Close()
	}()

	if err := serde.Serde(appConfig.Data, &bagongconfig.Config); err != nil {
		logger.Errorf("serde config: %s", err.Error())
		os.Exit(1)
	}
	bagongconfig.Config.EventHub = ev

	job := bagonglogic.ItemBalanceJob{}
	job.Run(hm)
}
