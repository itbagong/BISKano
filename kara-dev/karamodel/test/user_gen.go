package karamodel_test

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/brianvoe/gofakeit/v6"
)

type RegUserDummy struct {
	Email    string
	Password string
}

func GenerateUser(seed int64, num int, evhub kaos.EventHub, HolidayGroup []karamodel.HolidayProfile) ([]*karamodel.UserProfile, error) {
	gofakeit.Seed(seed)
	all := []*karamodel.UserProfile{}
	workingHour := []int{10, 20, 40}
	for i := 0; i < num; i++ {
		payload := &RegUserDummy{
			Email:    gofakeit.Email(),
			Password: "Password.1",
		}
		req := rbaclogic.CreateUserRequest{
			User: &rbacmodel.User{
				Email:       payload.Email,
				LoginID:     payload.Email,
				DisplayName: payload.Email,
				Status:      "Registered",
				Enable:      true,
			},
			Password: payload.Password,
		}

		userid := ""
		err := evhub.Publish("/v1/iam/user/create", &req, &userid, nil)
		if err != nil {
			return nil, err
		}
		newUserProfile := &karamodel.UserProfile{}
		newUserProfile.ID = userid
		newUserProfile.UserID = userid
		newUserProfile.MinimumWorkingHour = gofakeit.RandomInt(workingHour)
		newUserProfile.HolidayProfileID = HolidayGroup[gofakeit.IntRange(0, len(HolidayGroup)-1)].ID
		newUserProfile.Enable = true
		all = append(all, newUserProfile)
	}
	return all, nil
}
func GenerateUserProfile(seed int64, num int, HolidayGroup []karamodel.HolidayProfile) []*karamodel.UserProfile {
	all := []*karamodel.UserProfile{}
	gofakeit.Seed(seed)
	workingHour := []int{10, 20, 40}
	for i := 0; i < num; i++ {
		newUserProfile := &karamodel.UserProfile{}
		newUserProfile.ID = gofakeit.UUID()
		newUserProfile.MinimumWorkingHour = gofakeit.RandomInt(workingHour)
		newUserProfile.HolidayProfileID = HolidayGroup[gofakeit.IntRange(0, len(HolidayGroup))].ID
		newUserProfile.Enable = true
		all = append(all, newUserProfile)
	}
	return all
}
func GenerateRule() karamodel.AttendanceRule {
	result := karamodel.AttendanceRule{}
	result.ID = gofakeit.UUID()
	result.Name = gofakeit.FarmAnimal()
	result.IsShift = true
	result.PeriodStart = time.Now()
	result.PeriodEnd = result.PeriodStart.AddDate(0, 1, 0)
	return result
}
func GenerateRuleLine(attendanceRule karamodel.AttendanceRule) []karamodel.RuleLine {
	ruleLines := []karamodel.RuleLine{}
	shiftTimeStart := []string{"00:00", "08:00", "16:00"}
	shiftTimeEnd := []string{"00:30", "08:30", "16:30"}

	for idx, _ := range shiftTimeStart {
		newRuleLine := karamodel.RuleLine{}
		newRuleLine.CheckinStart = shiftTimeStart[idx]
		newRuleLine.CheckinEnd = shiftTimeEnd[idx]
		newRuleLine.RuleID = attendanceRule.ID
		newRuleLine.CheckoutStart = shiftTimeStart[(idx+1)%len(shiftTimeEnd)]
		newRuleLine.CheckoutEnd = shiftTimeEnd[(idx+1)%len(shiftTimeEnd)]
		newRuleLine.ID = gofakeit.UUID()
		newRuleLine.PersonPerBlock = 5
		newRuleLine.Name = gofakeit.PetName()
		newRuleLine.MinimumHour = 8
		newRuleLine.WorkStart = shiftTimeStart[idx]
		newRuleLine.WorkEnd = shiftTimeEnd[idx]
		newRuleLine.Days = []int{1, 2, 3, 4, 5} //senin-jumat
		ruleLines = append(ruleLines, newRuleLine)
	}
	return ruleLines
}
