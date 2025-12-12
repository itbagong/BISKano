package karalogic

import (
	"errors"
	"fmt"
	"math"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karaconfig"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func LeavaRequestGetPreMW(ctx *kaos.Context, payload interface{}) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, errors.New("missing: db")
	}

	req, ok := payload.(*karamodel.LeaveRequest)
	if !ok {
		return false, fmt.Errorf("invalid payload, got %t", payload)
	}
	jwtUserID := sebar.GetUserIDFromCtx(ctx)
	req.UserID = lo.Ternary(req.UserID == "", jwtUserID, req.UserID)

	if req.ID == "" {
		req.Status = "Draft"
	} else {

	}

	if err := db.Save(req); err != nil {
		return false, fmt.Errorf("save request: %s", err.Error())
	}

	return true, nil
}

func LeaveRequestGetsPostMW(ctx *kaos.Context, payload interface{}) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, errors.New("missing: db")
	}
	ev, err := ctx.DefaultEvent()
	if err != nil {
		return false, errors.New("missing: event hub")
	}
	userMap := map[string]*rbacmodel.User{}

	mres := ctx.Data().Get("FnResult", codekit.M{}).(codekit.M)
	res := *(mres.Get("data", &[]karamodel.LeaveRequest{}).(*[]karamodel.LeaveRequest))
	ids := make([]string, len(res))
	for index, re := range res {
		user := userMap[re.UserID]
		if user == nil {
			if err := ev.Publish(karaconfig.Config.GetUserTopic,
				&rbaclogic.GetUserByRequest{
					FindID: re.UserID,
					FindBy: "id",
				}, &user, nil); err == nil {
				userMap[re.UserID] = user
				re.UserID = user.DisplayName
			}
		} else {
			re.UserID = user.DisplayName
		}
		res[index] = re
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

type LeaveRequestLogic struct {
}

func (obj *LeaveRequestLogic) GetScreenStat(ctx *kaos.Context, payload string) ([]string, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	req, _ := datahub.GetByID(db, new(karamodel.LeaveRequest), payload)
	if req == nil {
		return nil, errors.New("missing: request")
	}

	res := []string{}
	jwtUserID := sebar.GetUserIDFromCtx(ctx)

	switch req.Status {
	case "Draft":
		res = append(res, "Submit")

	case "Submitted":
		if req.CurrentApprovers == nil {
			req.CurrentApprovers = []string{}
		}
		if lo.IndexOf(req.CurrentApprovers, jwtUserID) >= 0 {
			res = append(res, "Approve", "Reject")
		}
	}
	return res, nil
}

type GetLeaveTypeApproversRequest struct {
	LeaveTypeID string
	UserID      string
}

func (obj *LeaveRequestLogic) GetLeaveTypeApprovers(ctx *kaos.Context, payload *GetLeaveTypeApproversRequest) ([]string, error) {
	res := []string{}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return res, errors.New("missing: db")
	}

	if payload.UserID == "" {
		payload.UserID = sebar.GetUserIDFromCtx(ctx)
	}

	leaveType, err := datahub.GetByID(db, new(karamodel.LeaveType), payload.LeaveTypeID)
	if err != nil {
		return res, fmt.Errorf("missing: leave type: %s", payload.LeaveTypeID)
	}
	siteID := leaveType.DefaultSite

	emp, err := datahub.GetByFilter(db, new(tenantcoremodel.Employee), dbflex.Eqs("UserID", payload.UserID))
	if err == nil {
		empSiteID := emp.Dimension.Get("Site")
		if empSiteID != "" {
			siteID = empSiteID
		}
	}

	if siteID == "" {
		return res, fmt.Errorf("missing: site")
	}

	apv, err := datahub.GetByFilter(db, new(karamodel.LeaveApprovalSetup), dbflex.Eqs("LeaveTypeID", payload.LeaveTypeID, "Site", siteID))
	if err != nil {
		return res, fmt.Errorf("get approvers error: %s", err.Error())
	}

	res = apv.ApproverIDs
	return res, nil
}

func (obj *LeaveRequestLogic) Submit(ctx *kaos.Context, payload string) (*karamodel.LeaveRequest, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	leave, err := datahub.GetByID(db, new(karamodel.LeaveRequest), payload)
	if err != nil {
		return nil, errors.New("missing: leave request")
	}

	if len(leave.Approvers) == 0 {
		return nil, errors.New("missing: approver")
	}

	if leave.Status != "Draft" {
		return nil, errors.New("invalid status")
	}

	userID := sebar.GetUserIDFromCtx(ctx)

	v := int(0)
	leave.Status = "Submitted"
	leave.CurrentApproverIndex = &v
	leave.CurrentApprovers = leave.Approvers[*leave.CurrentApproverIndex].UserIDs
	leave.Approvals = append(leave.Approvals, karamodel.LineApproval{
		UserID:    userID,
		Timestamp: time.Now(),
		Op:        "Submit",
	})
	db.Save(leave)
	return leave, nil
}

type ApprovalRequest struct {
	ID     string
	Op     string
	Reason string
}

func (obj *LeaveRequestLogic) Approve(ctx *kaos.Context, payload *ApprovalRequest) (*karamodel.LeaveRequest, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	leave, err := datahub.GetByID(db, new(karamodel.LeaveRequest), payload.ID)
	if err != nil {
		return nil, errors.New("missing: leave request")
	}

	if leave.Status != "Submitted" {
		return nil, errors.New("invalid status")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if lo.IndexOf(leave.CurrentApprovers, userID) < 0 {
		return nil, errors.New("invalid approver")
	}

	switch payload.Op {
	case "Approve":
		// kurangi balance
		days := math.Ceil(leave.LeaveTo.Sub(leave.LeaveFrom).Hours() / 24)
		bal, _ := datahub.GetByFilter(db, new(karamodel.LeaveBalance), dbflex.Eqs("UserID", leave.UserID, "LeaveTypeID", leave.LeaveTypeID))
		if bal == nil {
			return nil, fmt.Errorf("no leave balance")
		}
		if bal.Balance < (int(days) + 1) {
			return nil, fmt.Errorf("not enough  balance")
		}

		lineApprovalsCount := len(leave.Approvers)
		if *(leave.CurrentApproverIndex) == lineApprovalsCount-1 {
			leave.Status = karamodel.LeaveRequestIsApproved
		} else {
			currentIndex := *(leave.CurrentApproverIndex) + 1
			leave.CurrentApproverIndex = &currentIndex
			leave.CurrentApprovers = leave.Approvers[currentIndex].UserIDs
		}
		leave.Approvals = append(leave.Approvals, karamodel.LineApproval{
			UserID:    userID,
			Timestamp: time.Now(),
			Op:        "Approve",
		})
		db.Save(leave)

		bal.Balance = bal.Balance - (int(days) + 1)
		db.Save(bal)

	case "Reject":
		if payload.Reason == "" {
			return nil, errors.New("missing: reason for rejection")
		}
		leave.Status = karamodel.LeaveRequestIsRejected
		leave.Reason = payload.Reason
		leave.Approvals = append(leave.Approvals, karamodel.LineApproval{
			UserID:    userID,
			Timestamp: time.Now(),
			Op:        "Reject",
			Reason:    payload.Reason,
		})
		db.Save(leave)

	default:
		return nil, fmt.Errorf("invalid op: %s", payload.Op)
	}

	return leave, nil
}
