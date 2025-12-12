package mfglogic_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/mfg/mfglogic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/strikememongo"
	"github.com/ariefdarmawan/strikememongo/strikememongolog"
	"github.com/ariefdarmawan/suim"
	"github.com/kanoteknologi/knats"
	natssvr "github.com/nats-io/nats-server/v2/server"
)

type injectDataFn func(ctx *kaos.Context) error

var (
	dbSvr           *strikememongo.Server
	defaultConnStr  = "mongodb://localhost:27017/dbapp"
	tenantDBConnStr = "%s/db_%s"
	testCoID1       = "testco1"
	testCoID2       = "testco2"
	//testDbPort      = codekit.RandInt(1000) + 28000
	testDbPort = 27888
	natsPort   = 4333
	svc        = kaos.NewService().SetBasePoint("/v1/mfg")
	svcMFGNats = kaos.NewService().SetBasePoint("/v1/mfg")
	svcTenant  = kaos.NewService().SetBasePoint("/v1/tenant")
	svcSCM     = kaos.NewService().SetBasePoint("/v1/scm")

	WHSID = "WHS001"
	SECID = "PLANT"
)

// NOTE: pastikan sdh setup ENV MEMONGO_MONGOD_BIN mengarah ke mongod
func TestMain(m *testing.M) {
	// mock nats
	opts := &natssvr.Options{Port: natsPort}
	ns, err := natssvr.NewServer(opts)
	if err != nil {
		panic(err)
	}
	go ns.Start()

	// PubSub
	ev := knats.NewEventHub(fmt.Sprintf("nats://localhost:%d", natsPort), byter.NewByter("")).SetSignature("testco")
	defer ev.Close()

	ficoconfig.Config.EventHub = ev
	ficoconfig.Config.SeqNumTopic = "/v1/tenant/numseq/claim-by-setup"
	mfglogic.Config.EventHub = ev

	// naming pattern
	kaos.NamingType = kaos.NamingIsLower
	kaos.NamingJoiner = "-"

	svc.Log().Info("initiate test")
	svc.Log().Infof("prepare test db server")
	dbSvr, err = strikememongo.StartWithOptions(&strikememongo.Options{
		// comment line dibawah apabila MEMONGO_MONGOD_BIN belum diset
		// MongodBin: "/Applications/mongodb/bin/mongod",
		// MongodBin: "C:\\Program Files\\MongoDB\\Server\\6.0\\bin\\mongod.exe",
		// Port:      testDbPort,
		// LogLevel: strikememongolog.LogLevelDebug,
		LogLevel:       strikememongolog.LogLevelInfo,
		StartupTimeout: 20 * time.Second,
		TempDirFolder:  "D:\\temp\\memongo",
	})
	if err != nil {
		log.Fatal(err)
	}

	svc.Log().Infof("prepare datahub")
	hm := kaos.NewHubManager(nil)
	defaultConn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(defaultConnStr, false), true, 50)
	defaultConn.SetAutoCloseDuration(2 * time.Second)
	hm.Set("default", "", defaultConn)
	hm.SetHubBuilder(func(key, group string) (*datahub.Hub, error) {
		vTenantConnStr := fmt.Sprintf(tenantDBConnStr, dbSvr.URI(), key)
		hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(vTenantConnStr, false), true, 100)
		hconn.SetAutoCloseDuration(2 * time.Second)
		//hconn.Cache(&datahub.CacheOpts{Provider: kvms}) // this can be turned on to enable cache
		return hconn, nil
	})
	defer hm.Close()
	svc.SetHubManager(hm)
	svc.RegisterEventHub(ev, "default", "testco")

	dbmod := sebar.NewDBModFromContext()
	uimod := suim.New()
	svc.Log().Infof("prepare mock registration")
	mfglogic.RegisterJournalAPI(svc, dbmod, uimod)
	svc.PrepareRoutes("")
	routes := svc.Routes()
	svc.Log().Infof("routes registered: %d", len(routes))

	// mock supporting service
	svcMFGNats.RegisterEventHub(ev, "default", "testco")
	svcMFGNats.SetHubManager(hm)
	mfglogic.RegisterNATSAPI(svcMFGNats, dbmod, uimod)
	svcMFGNats.PrepareRoutes("")
	if e := knats.NewDeployer(ev).Deploy(svcMFGNats, nil); e != nil {
		dbSvr.Stop()
		log.Fatal(e.Error())
		os.Exit(-1)
	}

	// mock supporting service
	svcTenant.RegisterEventHub(ev, "default", "testco")
	svcTenant.SetHubManager(hm)
	sequenceLogic := tenantcorelogic.DefaultSequence()
	svcTenant.Group().SetDeployer(knats.DeployerName).Apply(svcTenant.RegisterModel(sequenceLogic, "numseq"))

	svcTenant.PrepareRoutes("")
	if e := knats.NewDeployer(ev).Deploy(svcTenant, nil); e != nil {
		dbSvr.Stop()
		log.Fatal(e.Error())
		os.Exit(-1)
	}

	// mock supporting data
	ctx := kaos.NewContextFromService(svc, nil)
	svc.Log().Infof("prepare data")
	fns := []injectDataFn{
		injectLedgerBankMasterData,
		injectLedgerBankConfigData,
		injectItem,
		injectUnit,
		injectSCMData,
		injectMFGData,
		injectMFGDataBOM,
	}
	for _, fn := range fns {
		if err := fn(ctx); err != nil {
			svc.Log().Error(err.Error())
			dbSvr.Stop()
			os.Exit(-1)
		}
	}
	runResult := m.Run()
	dbSvr.Stop()
	os.Exit(runResult)
}
