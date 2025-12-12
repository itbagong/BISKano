package karalogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/kara/karamodel"
	"github.com/ariefdarmawan/datahub"
)

// Deserialize decodes query string into a map
func Deserialize(str string, res interface{}) error {
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		return fmt.Errorf("%s:\n%s", err, str)
	}
	return err
}

type EmployeeClockHour struct {
	EmployeeID     string
	ClockedHours   float64
	HolidayProfile string
}

// generate shift for certain location. If err is NotEnoughError, meaning shift planning is done
// and created in database, but there might be some defficiency in resource allocated in the shift
func GenerateShift(hub *datahub.Hub, dateStart, dateEnd time.Time, location karamodel.WorkLocation) error {
	i := time.Date(dateStart.Year(), dateStart.Month(), dateStart.Day(), dateStart.Hour(), dateStart.Minute(), dateStart.Second(), dateStart.Nanosecond(), dateStart.Location())
	workLocationUser := &karamodel.WorkLocationUser{}
	wsus := []karamodel.WorkLocationUser{}
	//RuleForDays := map[int]karamodel.RuleLine{}
	for {
		if i.After(dateEnd) {
			break
		}
		fmt.Println("CurDate", i)
		date1 := time.Date(i.Year(), i.Month(), i.Day(), 0, 0, 0, 0, time.Local)
		fmt.Println("CurDate", date1)
		queryAlreadyAssignedUser := dbflex.NewFilter("ShiftDate", dbflex.OpEq, date1, nil)
		shiftPlan := karamodel.ShiftPlan{}
		plannedAssignment := []karamodel.ShiftPlan{}
		hub.GetsByFilter(&shiftPlan, queryAlreadyAssignedUser, &plannedAssignment)

		alreadyAssignedUser := []string{}
		for _, val := range plannedAssignment {
			alreadyAssignedUser = append(alreadyAssignedUser, val.UserID)
		}
		// filterout users on leave
		leaveBeforeQuery := dbflex.NewFilter("LeaveFrom", dbflex.OpLte, i, nil)
		leaveAfterQuery := dbflex.NewFilter("LeaveTo", dbflex.OpGte, i, nil)
		leaveStatus := dbflex.NewFilter("Status", dbflex.OpEq, karamodel.LeaveRequestIsApproved, nil)
		queryLeave := &dbflex.Filter{}
		queryLeave.Op = dbflex.OpAnd
		queryLeave.Items = append(queryLeave.Items, leaveBeforeQuery, leaveAfterQuery, leaveStatus)
		leave := karamodel.LeaveRequest{}
		leaveRequests := []karamodel.LeaveRequest{}
		err := hub.GetsByFilter(&leave, queryLeave, &leaveRequests)
		if err != nil {
			return err
		}
		usersOnLeave := []string{}
		for _, i := range leaveRequests {
			usersOnLeave = append(usersOnLeave, i.ID)
		}

		query := &dbflex.Filter{}
		query.Op = dbflex.OpAnd
		beforeQuery := dbflex.NewFilter("From", dbflex.OpLte, i, nil)
		afterQuery := dbflex.NewFilter("To", dbflex.OpGte, i, nil)
		locationQuery := dbflex.NewFilter("WorkLocationID", dbflex.OpEq, location.ID, nil)
		assignedUser := dbflex.NewFilter("UserID", dbflex.OpNin, alreadyAssignedUser, nil)
		usersOnLeaveFilter := dbflex.NewFilter("UserID", dbflex.OpNin, usersOnLeave, nil)
		// applicableToday := dbflex.NewFilter("Days", dbflex.OpIn, []interface{}{int(i.Weekday())}, nil)
		query.Items = append(query.Items, beforeQuery, afterQuery, locationQuery, assignedUser, usersOnLeaveFilter)
		// skip get holidayprofile that is applied on this date
		paramHoliday := dbflex.NewQueryParam()
		queryHoliday := dbflex.NewFilter("Date", dbflex.OpEq, i, nil)
		paramHoliday.Where = queryHoliday
		holidayDetail := karamodel.HolidayItem{}
		holidayDetails := []karamodel.HolidayItem{}
		err = hub.Gets(&holidayDetail, paramHoliday, &holidayDetails)
		if err != nil {
			return err
		}
		skippedHolidayProfileIdMap := map[string]bool{}
		skippedHolidayProfileId := []string{}
		for _, val := range holidayDetails {
			if _, ok := skippedHolidayProfileIdMap[val.HolidayGroupID]; !ok {
				skippedHolidayProfileIdMap[val.HolidayGroupID] = true
				skippedHolidayProfileId = append(skippedHolidayProfileId, val.HolidayGroupID)
			}

		}

		// fmt.Println(beforeQuery)
		// fmt.Println(afterQuery)
		// fmt.Println(applicableToday)
		param := dbflex.NewQueryParam()
		param.Where = query
		param.Sort = append(param.Sort, "UserID")
		err = hub.Gets(workLocationUser, param, &wsus)
		if err != nil {
			return err
		}
		mapRuleIdUserId := map[string]map[string]bool{}
		uniqueRuleUserId := map[string][]string{} //map[ruleid][]userId

		ruleLIneId := map[string][]karamodel.RuleLine{} //map[RuleId]RuleLine
		allRuleId := []string{}
		for _, v := range wsus {
			if _, ok := uniqueRuleUserId[v.RuleID]; !ok {
				uniqueRuleUserId[v.RuleID] = []string{}
				allRuleId = append(allRuleId, v.RuleID)
			}
			if mapRuleIdUserId[v.RuleID] == nil {
				mapRuleIdUserId[v.RuleID] = map[string]bool{}
				uniqueRuleUserId[v.RuleID] = append(uniqueRuleUserId[v.RuleID], v.UserID)
			} else {
				if _, ok := mapRuleIdUserId[v.RuleID][v.UserID]; !ok {
					uniqueRuleUserId[v.RuleID] = append(uniqueRuleUserId[v.RuleID], v.UserID)
					mapRuleIdUserId[v.RuleID][v.UserID] = true
				}
			}
		}

		// filter out rule that is not a shift
		rule := &karamodel.AttendanceRule{}
		rules := []karamodel.AttendanceRule{}
		err = hub.GetsByFilter(rule, dbflex.NewFilter("_id", dbflex.OpIn, allRuleId, nil), &rules)
		if err != nil {
			return err
		}
		allRuleId = []string{}
		for _, r := range rules {
			if r.IsShift {
				allRuleId = append(allRuleId, r.ID)
			}
		}
		weekday := i.Weekday()
		TodaysRuleLine := []karamodel.RuleLine{}
		queryRuleLine := dbflex.NewFilter("", dbflex.OpAnd, nil,
			[]*dbflex.Filter{
				dbflex.NewFilter("RuleID", dbflex.OpIn, allRuleId, nil),
				dbflex.NewFilter("Days", dbflex.OpEq, weekday, nil),
			},
		)
		ruleLine := karamodel.RuleLine{}
		err = hub.GetsByFilter(&ruleLine, queryRuleLine, &TodaysRuleLine)
		if err != nil {
			return err
		}
		// fmt.Println("wsus", len(wsus), int(weekday), "ruleline", len(TodaysRuleLine))
		for _, v := range TodaysRuleLine {
			if _, ok := ruleLIneId[v.RuleID]; !ok {
				ruleLIneId[v.RuleID] = []karamodel.RuleLine{}
			}
			ruleLIneId[v.RuleID] = append(ruleLIneId[v.RuleID], v)
		}

		for key, _ := range ruleLIneId {
			// ruleout userid that cannot works this day
			userProfile := karamodel.UserProfile{}
			uniqueRuleUserKey := uniqueRuleUserId[key]
			fmt.Println(len(uniqueRuleUserKey))
			filter := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
				dbflex.NewFilter("_id", dbflex.OpIn, uniqueRuleUserKey, nil),
				// dbflex.NewFilter("HolidayProfileID", dbflex.OpNin, skippedHolidayProfileId, nil),
			})
			if len(skippedHolidayProfileId) > 0 {
				filter.Items = append(filter.Items, dbflex.NewFilter("HolidayProfileID", dbflex.OpNin, skippedHolidayProfileId, nil))
			}
			kk := []karamodel.UserProfile{}
			hub.GetsByFilter(&userProfile, filter, &kk)
			validUser := []string{}
			for _, ss := range kk {
				validUser = append(validUser, ss.ID)
			}

			RuleForDays, err := processShiftForDayLocationRule(hub, i, validUser, ruleLIneId[key], location.ID)
			if err != nil && err != NotEnoughError {
				return err
			}
			// fmt.Println("RulesForDays", len(RuleForDays))
			for key, userIds := range RuleForDays {
				gg := strings.Split(key, "_")
				date_str := gg[0]
				rules_line_id := gg[1]
				planDate, _ := time.Parse(time.RFC3339, date_str)
				// fmt.Println(planDate.Format(time.RFC3339))
				for _, u := range userIds {
					sp := &karamodel.ShiftPlan{}
					sp.ShiftDate = planDate
					sp.UserID = u
					sp.WorkLocationID = location.ID
					sp.RuleLineID = rules_line_id

					err := hub.Save(sp)
					if err != nil {
						return err
					}
				}
			}
			if err == NotEnoughError {
				return err
			}
		}

		i = i.AddDate(0, 0, 1)
	}
	return nil

}
func GetRuleLines(hub *datahub.Hub, ruleid string, rules *map[string][]karamodel.RuleLine) (int, error) {
	filter := dbflex.NewFilter("RuleID", dbflex.OpEq, ruleid, nil)
	ruleLine := &karamodel.RuleLine{}
	lines := []karamodel.RuleLine{}
	err := hub.GetsByFilter(ruleLine, filter, &lines)
	if err != nil {
		return -1, err
	}
	totalAssignmentRequired := 0
	for _, i := range lines {
		if (*rules)[i.RuleID] == nil {
			(*rules)[i.RuleID] = []karamodel.RuleLine{}
		}
		totalAssignmentRequired += i.PersonPerBlock
		(*rules)[i.RuleID] = append((*rules)[i.RuleID], i)
	}
	return totalAssignmentRequired, nil
}

// check how many hour clocked by userid in the last 7 days
func CheckHour(hub *datahub.Hub, date time.Time, userid string) float64 {
	SevenDaysAgo := date.AddDate(0, 0, -7)
	trx := []karamodel.AttendanceTrx{}
	filter := dbflex.NewFilter("", dbflex.OpAnd, nil,
		[]*dbflex.Filter{
			dbflex.NewFilter("UserID", dbflex.OpEq, userid, nil),
			dbflex.NewFilter("TrxDate", dbflex.OpGte, SevenDaysAgo, nil),
		},
	)
	attendance := karamodel.AttendanceTrx{}
	hub.GetsByFilter(&attendance, filter, &trx)
	totalHour := 0.0
	for _, t := range trx {
		totalHour += t.Hours
	}
	return totalHour
}

var (
	NotEnoughError = errors.New("Not Enough employee")
)

func processShiftForDayLocationRule(hub *datahub.Hub, date time.Time, userid []string, rulelines []karamodel.RuleLine, worklocationid string) (map[string][]string, error) {
	assignedEmployee := map[string][]string{}
	// ruleTotalCount := map[string]int{}
	wlrs_count := 0
	empWorkHours := []EmployeeClockHour{}
	for _, ee := range userid {
		newEmpClockHour := EmployeeClockHour{}
		newEmpClockHour.EmployeeID = ee
		newEmpClockHour.ClockedHours = CheckHour(hub, date, ee)
		empWorkHours = append(empWorkHours, newEmpClockHour)
	}
	sort.Slice(empWorkHours, func(i, j int) bool {
		return empWorkHours[i].ClockedHours < empWorkHours[j].ClockedHours
	})
	sp := &karamodel.ShiftPlan{}
	notEnough := false
ruleLineLoop:
	for _, rule := range rulelines {
		// skip this rulelines if the plans already cover this
		filter := dbflex.NewFilter("", dbflex.OpAnd, nil,
			[]*dbflex.Filter{
				dbflex.NewFilter("ShiftDate", dbflex.OpEq, date, nil),
				dbflex.NewFilter("RuleLineID", dbflex.OpEq, rule.ID, nil),
				dbflex.NewFilter("WorkLocationID", dbflex.OpEq, worklocationid, nil),
			},
		)
		fmt.Println("ShiftDate", date.Format(time.RFC3339), "RuleLineID", rule.ID)
		plannedAlready := []*karamodel.ShiftPlan{}
		hub.GetsByFilter(sp, filter, &plannedAlready)
		if len(plannedAlready) == rule.PersonPerBlock {
			continue
		}
		left := rule.PersonPerBlock - len(plannedAlready)
		for i := 0; i < left; i++ {
			key := fmt.Sprintf("%s_%s", date.Format(time.RFC3339), rule.ID)
			// fmt.Println(key)
			employee := empWorkHours[wlrs_count]
			// if _, ok := workHour[employee]; !ok {
			// 	workHour[employee] = CheckHour(hub, date, employee)
			// }
			fmt.Println(employee.EmployeeID, wlrs_count)
			assignedEmployee[key] = append(assignedEmployee[key], employee.EmployeeID)
			wlrs_count += 1
			if wlrs_count == len(userid) {
				if i < left-1 {
					notEnough = true
				}
				fmt.Println("====")
				break ruleLineLoop
			}
		}
	}
	if notEnough {
		return assignedEmployee, NotEnoughError
	} else {
		return assignedEmployee, nil
	}

}
