package afycorelogic_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/afycore/afycorelogic"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/strikememongo"
	"github.com/ariefdarmawan/strikememongo/strikememongolog"
	"github.com/kanoteknologi/knats"
	natssvr "github.com/nats-io/nats-server/v2/server"
	"github.com/sebarcode/codekit"
)

type injectDataFn func(ctx *kaos.Context) error

var (
	dbSvr           *strikememongo.Server
	defaultConnStr  = "mongodb://localhost:27017/dbapp"
	tenantDBConnStr = "%s/db_%s"
	testCoID1       = "testco1"
	testCoID2       = "testco2"
	testDbPort      = codekit.RandInt(1000) + 28000
	//testDbPort = 27888
	natsPort = 4333
	svc      = kaos.NewService().SetBasePoint("/v1/core")
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

	// naming pattern
	kaos.NamingType = kaos.NamingIsLower
	kaos.NamingJoiner = "-"

	svc.Log().Info("initiate test")
	svc.Log().Infof("prepare test db server")
	dbSvr, err = strikememongo.StartWithOptions(&strikememongo.Options{
		// comment line dibawah apabila MEMONGO_MONGOD_BIN belum diset
		//MongodBin: "/Applications/mongodb/bin/mongod",
		MongodBin:        "C:\\Program Files\\MongoDB\\Server\\5.0\\bin\\mongod.exe",
		Port:             testDbPort,
		ShouldUseReplica: false,
		//LogLevel:  strikememongolog.LogLevelDebug,
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

	svc.Log().Infof("prepare mock registration")
	afycorelogic.RegisterCore(svc)
	svc.PrepareRoutes("")

	// mock supporting data
	ctx := kaos.NewContextFromService(svc, nil)
	svc.Log().Infof("prepare data")
	fns := []injectDataFn{injectCoreData}
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

func insertJournal[M orm.DataModel](j M, uriPath, userid, coid string) (M, error) {
	sr := svc.GetRoute(uriPath)
	if sr == nil {
		return j, fmt.Errorf("missing: route: %s", uriPath)
	}
	ctx := prepareCtxData(kaos.NewContextFromService(svc, sr), userid, coid)
	e := svc.CallTo(uriPath, j, ctx, j)
	return j, e
}

func prepareCtxData(ctx *kaos.Context, userid, companyid string) *kaos.Context {
	if userid != "" {
		ctx.Data().Set("jwt_reference_id", userid)
	}

	if companyid == "" {
		companyid = testCoID1
	}
	ctx.Data().Set("jwt_data", codekit.M{}.Set("CompanyID", companyid))
	return ctx
}

func insertModel[T orm.DataModel](h *datahub.Hub, records []T) error {
	tableName := ""
	for _, obj := range records {
		if e := h.Insert(obj); e != nil {
			return fmt.Errorf("%s: %s: %s", obj.TableName(), modelID(obj), e.Error())
		}
		if tableName == "" {
			tableName = obj.TableName()
		}
	}
	return nil
}

func modelID(d orm.DataModel) string {
	_, ids := d.GetID(nil)
	if len(ids) == 0 {
		return "NoID"
	}
	return fmt.Sprintf("%v", ids[0])
}
