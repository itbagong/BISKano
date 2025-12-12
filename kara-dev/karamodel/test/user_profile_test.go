package karamodel_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"git.kanosolution.net/sebar/kara/karamodel"
	"github.com/ariefdarmawan/byter"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/kanoteknologi/knats"
)

// pastikan dulu service IAM berjalan. Fungsi GenerateUser mengirim ke nats alamat /v1/iam/user/create untuk create user
// jangan lupa letakan file app.yml di folder yang sama dengan file test ini
// contoh app.yml:
//
// Config:
//
//	db: "mongodb://localhost:27017/kara_dev?idle=10&poolsize=500"
//	PoolSize: 500
//	EvServer: "nats://localhost:4222"
//	EvGroup: "xibar"
//	EventChangeTopic: /config/change
func TestCreateUser(t *testing.T) {
	conf, err := ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	t.Log(conf["Config"])
	conf2 := conf["Config"].(map[string]interface{})
	connStr := conf2["db"].(string)
	poolsize := conf2["PoolSize"].(int)
	// hm := kaos.NewHubManager(nil)
	evServer := conf2["EvServer"].(string)
	evGroup := conf2["EvGroup"].(string)
	ev := knats.NewEventHub(evServer, byter.NewByter("")).SetSignature(evGroup)
	hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)
	hconn.SetAutoCloseDuration(2 * time.Second)
	gofakeit.Seed(10)
	HolidayGroup := []karamodel.HolidayProfile{}
	for i := 0; i < 5; i++ {
		holGroup := karamodel.HolidayProfile{}
		holGroup.ID = gofakeit.UUID()
		holGroup.Name = gofakeit.Adverb() + " " + gofakeit.FarmAnimal()
		HolidayGroup = append(HolidayGroup, holGroup)
		hconn.Save(&holGroup)
	}
	locations := []karamodel.WorkLocation{}
	for i := 0; i < 5; i++ {
		workLocation := karamodel.WorkLocation{}
		workLocation.Address = gofakeit.Address().Address
		workLocation.Enable = true
		workLocation.ID = gofakeit.UUID()
		workLocation.Virtual = false
		workLocation.TimeLoc = "Asia/Jakarta"
		workLocation.Name = gofakeit.Company()

		err = hconn.Save(&workLocation)
		locations = append(locations, workLocation)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()

		}

	}
	rule := GenerateRule()
	err = hconn.Save(&rule)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()

	}
	ruleLines := GenerateRuleLine(rule)
	for _, v := range ruleLines {
		// rule := GenerateRule()
		err = hconn.Save(&v)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
		}
	}
	// hm.Set("tenant", "", hconn)
	newUsers, err := GenerateUser(2, 5000, ev, HolidayGroup)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	for _, u := range newUsers {
		err := hconn.Save(u)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()

		}
	}
	rand.Seed(34)
	locationDistArr := []int{1, 1, 1, 1, 1, 2, 2, 2, 3}
	startDate := time.Date(2023, 8, 1, 0, 0, 0, 0, time.Local)
	for _, val := range newUsers {
		workLocNum := gofakeit.RandomInt(locationDistArr) //math.Floor(rand.NormFloat64() * 1.2)
		fmt.Println(workLocNum)
		for i := 0; i < int(workLocNum); i++ {
			locIdx := rand.Int() % len(locations)
			loc := locations[locIdx]

			wlu := karamodel.WorkLocationUser{}
			wlu.ID = gofakeit.UUID()
			wlu.UserID = val.ID
			wlu.WorkLocationID = loc.ID
			wlu.RuleID = rule.ID
			wlu.From = startDate
			wlu.To = startDate.AddDate(0, 1, 0)
			err = hconn.Save(&wlu)
			if err != nil {
				t.Log(err.Error())
				t.FailNow()
			}
		}
	}

}
