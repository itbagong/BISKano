package sebar

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"reflect"
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
)

func ConfigToApp(configPath, portPrefix string) (*AppConfig, *kaos.HubManager, kaos.EventHub, error) {
	appConfig := new(AppConfig)
	AppConfigVariables = codekit.M{"port_prefix": portPrefix}
	e := ReadConfig(configPath, appConfig)
	if e != nil {
		return nil, nil, nil, e
	}

	appConfig.Parse()
	evServer := appConfig.EventServer

	// PubSub
	ev := knats.NewEventHub(evServer.Server, byter.NewByter("")).SetSignature(appConfig.EventServer.Group)
	if evServer.Timeout == 0 {
		ev.SetTimeout(30 * time.Second)
	} else {
		ev.SetTimeout(time.Second * time.Duration(evServer.Timeout))
	}

	// DataHub
	hm := kaos.NewHubManager(nil)
	vTenantConn := ""
	vTenantUseTx := false
	vTenantPoolSize := 100
	for k, v := range appConfig.Connections {
		if k == "tenant" {
			vTenantConn = v.Txt
			vTenantPoolSize = v.PoolSize
			vTenantUseTx = v.UseTx
			continue
		}
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(v.Txt, v.UseTx), true, v.PoolSize)
		hconn.SetAutoCloseDuration(2 * time.Second)
		hm.Set(k, "", hconn)
	}
	hm.SetHubBuilder(func(key, group string) (*datahub.Hub, error) {
		vTenantConnStr := fmt.Sprintf(vTenantConn, key)
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(vTenantConnStr, vTenantUseTx), true, vTenantPoolSize)
		hconn.SetAutoCloseDuration(2 * time.Second)
		return hconn, nil
	})

	return appConfig, hm, ev, nil
}

func Deploy(appConfig *AppConfig, s *kaos.Service, serviceName string, logger *logger.LogEngine, ev kaos.EventHub) {
	// event server
	if e := knats.NewDeployer(ev).Deploy(s, nil); e != nil {
		logger.Errorf("unable to deploy. %s", e.Error())
		os.Exit(1)
	}

	// rest server
	mux := http.NewServeMux()
	hostname := appConfig.Hosts[serviceName]

	if e := hd.NewHttpDeployer(WrapApiError).Deploy(s, mux); e != nil {
		logger.Errorf("unable to deploy. %s", e.Error())
		os.Exit(1)
	}
	logger.Infof("starting %v service on %s", serviceName, hostname)
	go http.ListenAndServe(hostname, mux)
}

func SyncModel(s *kaos.Service, hub *datahub.Hub, models ...orm.DataModel) {
	for _, mdl := range models {
		vt := reflect.TypeOf(mdl).Elem()
		if e := hub.EnsureDb(mdl); e != nil {
			s.Log().Errorf("fail to sync db %s. %s", vt.Name(), e.Error())
			os.Exit(1)
		}
		s.Log().Infof("syncing %s", vt.Name())
	}
}

func EnsurePathExist(p string, defaultPerm fs.FileMode) {
	if _, e := os.Stat(p); os.IsNotExist(e) {
		os.MkdirAll(p, defaultPerm)
	}
}
