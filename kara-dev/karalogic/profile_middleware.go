package karalogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type GetUserByRequest struct {
	FindBy string
	FindID string
}
type GetUserByResult struct {
	ID          string `json:"_id" bson:"_id"`
	LoginID     string
	DisplayName string
	Email       string
}

func MWProfilePostFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	// fmt.Println("Start PostMW")
	ev, _ := ctx.DefaultEvent()
	// fmt.Println(ctx.Data().Get("FnResult", nil))
	objs, err := sebar.ExtractDbmodFindResult(ctx, karamodel.UserProfile{})
	if err != nil {
		fmt.Println(err.Error())
		return true, nil
	}
	userIDs := lo.Map(objs, func(obj karamodel.UserProfile, index int) interface{} {
		return obj.ID
	})
	// fmt.Println(len(userIDs))
	allResult := []karamodel.UserProfileExtended{}
	for idx, id := range userIDs {
		// fmt.Println("Checking _id", id)
		req := GetUserByRequest{FindBy: "_id", FindID: id.(string)}
		reply := []GetUserByResult{}
		err := ev.Publish("/v1/iam/user/find-by", req, &reply, nil)
		if err != nil {
			fmt.Println(err.Error())
			return false, err
		}
		newRes := karamodel.UserProfileExtended{}
		newRes.ID = objs[idx].ID
		newRes.UserID = id.(string)
		newRes.HolidayProfileID = objs[idx].HolidayProfileID
		newRes.Enable = objs[idx].Enable
		newRes.MinimumWorkingHour = objs[idx].MinimumWorkingHour
		newRes.LastUpdate = objs[idx].LastUpdate
		newRes.Username = reply[0].DisplayName
		newRes.Email = reply[0].Email
		allResult = append(allResult, newRes)
		// objs[idx].Username = reply.DisplayName
		// objs[idx].Email = reply.Email
	}
	// fmt.Println("Done Populating", len(objs))
	ctx.Data().Set("FnResult", &allResult)
	return true, nil
	//users, err := datahub.FindByFilter(ev, &rbacmodel.User{}, dbflex.In("_id", userIDs...))
}

func MWProfilePostGets(ctx *kaos.Context, payload interface{}) (bool, error) {
	ev, _ := ctx.DefaultEvent()
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, errors.New("missing: db")
	}

	holidays := []karamodel.HolidayProfile{}
	err := db.Gets(new(karamodel.HolidayProfile), dbflex.NewQueryParam(), &holidays)
	if err != nil {
		return false, err
	}

	mapHoliday := lo.Associate(holidays, func(h karamodel.HolidayProfile) (string, string) {
		return h.ID, h.Name
	})

	mres := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
	res := *(mres.Get("data", &[]karamodel.UserProfile{}).(*[]karamodel.UserProfile))
	for i := range res {
		req := GetUserByRequest{FindBy: "_id", FindID: res[i].UserID}
		reply := []GetUserByResult{}
		err := ev.Publish("/v1/iam/user/find-by", req, &reply, nil)
		if err != nil {
			return false, err
		}

		res[i].UserID = reply[0].DisplayName
		res[i].HolidayProfileID = mapHoliday[res[i].HolidayProfileID]
	}

	mres.Set("data", res)
	ctx.Data().Set("FnResult", mres)
	return true, nil
}
