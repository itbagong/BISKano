package flowlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/flow/flowmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func createRequestTask(ctx *kaos.Context, request *flowmodel.Request, tasks flowmodel.Tasks) error {
	h := sebar.GetTenantDBFromContext(ctx)
	for _, task := range tasks {
		//-- validate task
		_, ok := lo.Find(request.Template.Tasks, func(t flowmodel.Task) bool {
			return t.ID == task.ID
		})
		if !ok {
			continue
		}

		rt, _ := datahub.GetByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", request.ID, "Version", request.Version, "TaskID", task.ID))
		if rt == nil {
			return fmt.Errorf("duplicate: request task: reqid %s, version %d, task %s", request.ID, request.Version, task.Name)
		}

		rt = new(flowmodel.RequestTask)
		rt.RequestID = request.ID
		rt.Version = request.Version
		rt.Setup = task
		rt.Assigned = time.Now()
		rt.Status = flowmodel.TaskRunning
		if e := h.Save(rt); e != nil {
			return fmt.Errorf("save request task: reqid %s, version %d, task %s: %s", request.ID, request.Version, task.Name, e.Error())
		}

		if rt.Setup.TaskType == flowmodel.TaskProcess {
			runTask(ctx, rt)
		}
	}
	return nil
}

func runTask(ctx *kaos.Context, task *flowmodel.RequestTask) error {
	if task.Status != flowmodel.TaskRunning {
		return fmt.Errorf("invalid: task status: %s, %s: %s", task.RequestID, task.Setup.Name, string(task.Status))
	}
	switch task.Setup.TaskType {
	case flowmodel.TaskProcess:

	default:
		return nil
	}
	if task.Status == flowmodel.TaskSuccess || task.Status == flowmodel.TaskFail {
		go nextTask(ctx, task)
	}
	return nil
}

func nextTask(ctx *kaos.Context, task *flowmodel.RequestTask) error {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("missing: db")
	}
	request, _ := datahub.GetByID(h, new(flowmodel.Request), task.RequestID)
	if request == nil {
		return errors.New("missing: request")
	}
	if task.Status == flowmodel.TaskFail && task.Setup.StopRequestIfFail {
		return stopRequest(ctx, request, flowmodel.RequestFail, "")
	}
	if task.Status == flowmodel.TaskSuccess && task.Setup.StopRequestIfSuccess {
		return stopRequest(ctx, request, flowmodel.RequestSuccess, task.Error)
	}

	routeToIDs := lo.FindUniques(lo.Map(lo.Filter(request.Template.Routes, func(r flowmodel.Route, index int) bool {
		return r.FromID == task.Setup.ID
	}), func(r flowmodel.Route, index int) string {
		return r.ToID
	}))

	tasks := lo.Filter(lo.Map(routeToIDs, func(id string, index int) flowmodel.Task {
		task, ok := lo.Find(request.Template.Tasks, func(t flowmodel.Task) bool {
			return t.ID == id
		})
		if !ok {
			return flowmodel.Task{ID: ""}
		}
		return task
	}), func(t flowmodel.Task, i int) bool {
		return t.ID != ""
	})

	routeTasks := lo.Map(tasks, func(t flowmodel.Task, i int) *flowmodel.RequestTask {
		rt := new(flowmodel.RequestTask)
		rt.RequestID = request.ID
		rt.Setup = t
		rt.Version = request.Version
		rt.Assigned = time.Now()
		rt.Status = flowmodel.TaskRunning
		return rt
	})

	for _, rt := range routeTasks {
		if e := h.Save(rt); e != nil {
			return fmt.Errorf("save: request task: %s, %s: %s", request.ID, rt.Setup.ID, e.Error())
		}
	}

	return nil
}

func stopRequest(ctx *kaos.Context, request *flowmodel.Request, status flowmodel.RequestStatus, reason string) error {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("missing: db")
	}
	request.Status = status
	request.Reason = reason
	err := h.Save(request, "Status", "Reason")
	if err != nil {
		return fmt.Errorf("stopping request: %s, %s: %s", request.ID, request.Name, err.Error())
	}
	tasks, _ := datahub.FindByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", request.ID, "Version", request.Version))
	for _, task := range tasks {
		if task.Status != flowmodel.TaskRunning {
			continue
		}
		task.Status = flowmodel.TaskCancel
		task.Text = "request is stopped"
		h.Save(task, "Status", "Text")
	}
	return nil
}

func reviewTask(ctx *kaos.Context, task *flowmodel.RequestTask, approve bool, reason string) error {
	if task.Status != flowmodel.TaskRunning {
		return fmt.Errorf("invalid: task status: %s, %s: %s", task.RequestID, task.Setup.Name, string(task.Status))
	}

	if task.Setup.TaskType != flowmodel.TaskReview {
		return fmt.Errorf("invalid: task type: %s, %s: %s", task.RequestID, task.Setup.Name, string(task.Setup.TaskType))
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("missing: db")
	}

	userid := sebar.GetUserIDFromCtx(ctx)

	// validate user
	_, found := lo.Find(task.Setup.Users, func(u flowmodel.TaskUser) bool {
		return u.UserID == userid
	})
	if !found {
		return fmt.Errorf("invalid: user: request %s, task %s, user: %s", task.RequestID, task.Setup.ID, userid)
	}

	// check if user already review
	_, hasReview := lo.Find(task.Approval, func(a flowmodel.UserApproval) bool {
		return a.UserID == userid
	})
	if hasReview {
		return errors.New("duplicate: user review")
	}
	review := flowmodel.UserApproval{
		UserID:   userid,
		Approval: approve,
		Reason:   reason,
		Time:     time.Now(),
	}
	task.Approval = append(task.Approval, review)

	moveToNext := false
	if review.Approval && len(task.Approval) == len(task.Setup.Users) {
		task.Status = flowmodel.TaskSuccess
		moveToNext = true
	} else if !review.Approval {
		task.Status = flowmodel.TaskFail
		task.Text = review.Reason
		moveToNext = true
	}
	if err := h.Save(task); err != nil {
		return fmt.Errorf("update task status: %s, %s: %s", task.RequestID, task.Setup.Name, err.Error())
	}
	if moveToNext {
		return nextTask(ctx, task)
	}

	return nil
}
