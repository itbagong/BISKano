package karalogic

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/dbflex"
	karamodel_test "git.kanosolution.net/sebar/kara/karamodel/test"

	"git.kanosolution.net/sebar/kara/karamodel"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
)

func TestScheduling(t *testing.T) {
	conf, err := karamodel_test.ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	t.Log(conf["Config"])
	conf2 := conf["Config"].(map[string]interface{})
	connStr := conf2["db"].(string)
	poolsize := conf2["PoolSize"].(int)
	hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)
	startDate := time.Date(2023, 8, 9, 0, 0, 0, 0, time.Local)
	endDate := time.Date(2023, 8, 9, 0, 0, 0, 0, time.Local)
	locIds := []string{"c44709ac-14cc-4eae-b75f-13228072e7b3",
		"62a582b7-e28e-426d-aeb5-6920f90f6473",
	}
	for _, locId := range locIds {
		fmt.Println("Processing", locId)
		location := karamodel.WorkLocation{ID: locId}
		err = GenerateShift(hconn, startDate, endDate, location)
		if err != nil && err != NotEnoughError {
			fmt.Println(err.Error())
			t.FailNow()
		}
		if err != nil && err == NotEnoughError {
			fmt.Println(err.Error())
			t.FailNow()
		}
	}
	//verify stuff
	HolidayGroupID := []string{"35994333-fe36-4147-8293-0f12c53f0d3c"}
	shiftPlans := []karamodel.ShiftPlan{}
	plan := karamodel.ShiftPlan{}
	users := []karamodel.UserProfile{}
	user := karamodel.UserProfile{}
	queryParm := dbflex.NewQueryParam()
	queryParm.Take = 30
	err = hconn.Gets(&plan, queryParm, &shiftPlans)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	userIds := []string{}
	for _, p := range shiftPlans {
		userIds = append(userIds, p.UserID)
	}
	queryParm.Where = dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
		dbflex.NewFilter("_id", dbflex.OpIn, userIds, nil),
		dbflex.NewFilter("HolidayProfileID", dbflex.OpIn, HolidayGroupID, nil),
	})
	err = hconn.Gets(&user, queryParm, &users)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if len(users) != 0 {
		t.Log("Salah assign")
		t.FailNow()
	}
}
func TestFetch3(t *testing.T) {
	conf, err := karamodel_test.ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	userProfile := karamodel.UserProfile{}
	t.Log(conf["Config"])
	conf2 := conf["Config"].(map[string]interface{})
	connStr := conf2["db"].(string)
	t.Log(connStr)
	poolsize := conf2["PoolSize"].(int)
	hub := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)
	uniqueRuleUserKey := []string{"20230816144052415PNLABXE726NO9GS", "20230816144052423QHPU3Z7D1EC8B91"}
	skippedHolidayProfileId := []string{"20230816144052415PNLABXE726NO9GS"}
	filter := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
		dbflex.NewFilter("_id", dbflex.OpIn, uniqueRuleUserKey, nil),
		// dbflex.NewFilter("HolidayProfileID", dbflex.OpNin, skippedHolidayProfileId, nil),
	})
	if len(skippedHolidayProfileId) > 0 {
		filter.Items = append(filter.Items, dbflex.NewFilter("HolidayProfileID", dbflex.OpNin, skippedHolidayProfileId, nil))
	}
	kk := []karamodel.UserProfile{}
	hub.GetAnyByFilter(userProfile.TableName(), filter, &kk)
}
func TestScheduling_filterDate2(t *testing.T) {
	conf, err := karamodel_test.ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	t.Log(conf["Config"])
	conf2 := conf["Config"].(map[string]interface{})
	connStr := conf2["db"].(string)
	t.Log(connStr)
	poolsize := conf2["PoolSize"].(int)
	hub := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)

	startDate := time.Date(2023, 8, 9, 0, 0, 0, 0, time.Local)
	// endDate := time.Date(2023, 8, 9, 0, 0, 0, 0, time.Local)
	locIds := []string{"2d729620-d411-44ff-9a94-d36e5d3e01b4",
		"49c23844-9051-4077-b414-534be3f70472",
		"e4df10ff-e267-4112-a838-8fd0426a627d",
		"ed5213a5-6675-473c-89f6-5fc1f46927eb",
		"46fbf6d1-4bc9-4cfd-9bab-e862d482ca0c",
	}
	workLocationUser := &karamodel.WorkLocationUser{}
	wsus := []karamodel.WorkLocationUser{}
	for _, locId := range locIds {
		i := startDate
		query := &dbflex.Filter{}
		query.Op = dbflex.OpAnd
		beforeQuery := dbflex.NewFilter("From", dbflex.OpLte, i, nil)
		afterQuery := dbflex.NewFilter("To", dbflex.OpGte, i, nil)
		locationQuery := dbflex.NewFilter("WorkLocationID", dbflex.OpEq, locId, nil)
		assignedUser := dbflex.NewFilter("UserID", dbflex.OpNin, []string{}, nil)
		query.Items = append(query.Items, locationQuery, beforeQuery, afterQuery, assignedUser)
		paramHoliday := dbflex.NewQueryParam()
		queryHoliday := dbflex.NewFilter("Date", dbflex.OpEq, i, nil)
		paramHoliday.Where = queryHoliday
		holidayDetail := karamodel.HolidayItem{}
		holidayDetails := []karamodel.HolidayItem{}
		err := hub.Gets(&holidayDetail, paramHoliday, &holidayDetails)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
			return
		}
		skippedHolidayProfileIdMap := map[string]bool{}
		skippedHolidayProfileId := []string{}
		for _, val := range holidayDetails {
			if _, ok := skippedHolidayProfileIdMap[val.HolidayGroupID]; !ok {
				skippedHolidayProfileIdMap[val.HolidayGroupID] = true
				skippedHolidayProfileId = append(skippedHolidayProfileId, val.HolidayGroupID)
			}

		}
		param := dbflex.NewQueryParam()
		param.Where = query
		param.Sort = append(param.Sort, "UserID")

		err = hub.Gets(workLocationUser, param, &wsus)
		if err != nil {
			t.Log(err.Error())
			t.FailNow()
			return
		}
		t.Log(len(wsus))
	}
}
func TestScheduling_filterDate(t *testing.T) {
	conf, err := karamodel_test.ReadConfig()
	if err != nil {
		t.FailNow()
		return
	}
	t.Log(conf["Config"])
	conf2 := conf["Config"].(map[string]interface{})
	connStr := conf2["db"].(string)
	poolsize := conf2["PoolSize"].(int)
	hconn := datahub.NewHub(datahub.GeneralDbConnBuilderWithTx(connStr, false), true, poolsize)
	//hub := hconn.GetClassicConnection()
	//2023-08-09T00:00:00+07:00 RuleLineID ff48b0b1-7349-4bf4-85ea-66ff840e2ec6
	startDate, _ := time.Parse(time.RFC3339, "2023-08-09T00:00:00+07:00") //time.Date(2023, 8, 9, 0, 0, 0, 0, time.Local)
	fmt.Println("CurDate", startDate)
	queryAlreadyAssignedUser := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
		dbflex.NewFilter("ShiftDate", dbflex.OpEq, startDate, nil),
		dbflex.NewFilter("RuleLineID", dbflex.OpEq, "ff48b0b1-7349-4bf4-85ea-66ff840e2ec6", nil),
	})
	shiftPlan := karamodel.ShiftPlan{}
	plannedAssignment := []karamodel.ShiftPlan{}
	hconn.GetsByFilter(&shiftPlan, queryAlreadyAssignedUser, &plannedAssignment)
	if len(plannedAssignment) == 0 {
		t.Fail()
	}
	fmt.Println(startDate.Format(time.RFC3339))
	fmt.Println(len(plannedAssignment))
}
