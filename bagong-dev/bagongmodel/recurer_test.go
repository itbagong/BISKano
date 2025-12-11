package bagongmodel_test

import (
	"io/ioutil"
	"testing"
	"time"

	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"gopkg.in/yaml.v2"
)

func ReadConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{}
	configByte, err := ioutil.ReadFile("app.yml")
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(configByte, &config)
	return config, nil
}

func TestRecurrence(t *testing.T) {
	conf, err := ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	// t.Log(conf["Config"])
	conf2 := conf["Config"].(map[interface{}]interface{})
	connStr := conf2["db"].(string)
	poolsize := conf2["PoolSize"].(int)
	hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)
	vendorSubmission := bagongmodel.VendorSubmission{}
	vendorSubmission.TrxDate = time.Now()
	vendorSubmission.CompanyID = "OXSDSDPD"
	vendorSubmission.CurrencyID = "PDSPDS"
	vendorSubmission.InvoiceID = "POPOSDSD"
	recuringParam := bagongmodel.RecuringParam{DateStart: time.Now(), DateEnd: time.Now().AddDate(0, 1, 0), Freq: bagongmodel.WEEKLY}
	hconn.Save(&vendorSubmission)
	bagongmodel.Recur(hconn, &vendorSubmission, recuringParam)

}
