package karalogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func LeaveBalanceRequestGetsPostMW(ctx *kaos.Context, payload interface{}) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, errors.New("missing: db")
	}

	mres := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
	res := *(mres.Get("data", &[]karamodel.LeaveBalance{}).(*[]karamodel.LeaveBalance))
	ids := make([]string, len(res))
	for index, re := range res {
		ids[index] = re.LeaveTypeID
	}

	leaves := []karamodel.LeaveType{}
	e := db.GetsByFilter(new(karamodel.LeaveType), dbflex.In("_id", ids...), &leaves)
	if e != nil {
		ctx.Log().Errorf("error when get leave type: %s", e.Error())
	}

	mapLeave := lo.Associate(leaves, func(leave karamodel.LeaveType) (string, string) {
		return leave.ID, leave.Name
	})

	for i := range res {
		if v, ok := mapLeave[res[i].LeaveTypeID]; ok {
			res[i].LeaveTypeID = v
		}
	}

	mres.Set("data", res)
	ctx.Data().Set("FnResult", mres)
	return true, nil
}
