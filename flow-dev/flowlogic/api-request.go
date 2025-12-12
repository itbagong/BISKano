package flowlogic

import (
	"errors"
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/flow/flowmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type Request struct {
}

type RequestPayload struct {
	TemplateID string
	Title      string
	Start      bool
	Payload    codekit.M
}

func (obj *Request) Create(ctx *kaos.Context, payload *RequestPayload) (*flowmodel.Request, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	ft, err := datahub.GetByID(h, new(flowmodel.FlowTemplate), payload.TemplateID)
	if err != nil {
		return nil, ctx.Log().Error2("invalid: flow template", "invalid: flow template: %s, %s", payload.TemplateID, err.Error())
	}

	uid := sebar.GetUserIDFromCtx(ctx)
	if uid == "" {
		return nil, errors.New("missing: user")
	}

	cid := ctx.Data().Get("CompanyID", "").(string)
	if cid == "" {
		return nil, errors.New("missing: company")
	}

	if len(ft.Tasks) == 0 {
		return nil, fmt.Errorf("invalid: tasks length is 0")
	}

	res := new(flowmodel.Request)
	res.Template = ft
	res.Name = tenantcorelogic.TernaryString(payload.Title, ft.Name)
	res.CompanyID = cid
	res.CreatedBy = uid
	res.Status = flowmodel.RequestDraft
	res.Version = 0
	res.Data = payload.Payload

	if err = h.Save(res); err != nil {
		return nil, ctx.Log().Error2("fail: create request", "fail: create request: %s, %s", payload.TemplateID, err.Error())
	}

	if payload.Start {
		return obj.Start(ctx, res)
	}

	return res, nil
}

func (obj *Request) Start(ctx *kaos.Context, payload *flowmodel.Request) (*flowmodel.Request, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	res, err := datahub.GetByID(h, new(flowmodel.Request), payload.ID)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("missing: request: %s, %s", payload.ID, payload.Name)
		} else {
			return nil, ctx.Log().Error2("fail: start request", "start request: %s, %s: %s", payload.ID, payload.Name, err.Error())
		}
	}
	if res.Status == flowmodel.RequestSuccess || res.Status == flowmodel.RequestFail || res.Status == flowmodel.RequestRunning {
		return nil, fmt.Errorf("invalid status: %s, %s", res.ID, res.Status)
	}

	if res.Status != flowmodel.RequestDraft {
		rh := new(flowmodel.RequestHistory)
		rh.RequestID = res.ID
		rh.Request = new(flowmodel.Request)
		*rh.Request = *res
		h.Save(rh)
	}

	res.Version++
	res.Status = flowmodel.RequestRunning
	if err = h.Save(res, "Version", "Status"); err != nil {
		return nil, ctx.Log().Error2("fail: update request status", "update request status: %s: %s", res.ID, err.Error())
	}

	initialTasks := res.Template.GetGenesisTask()
	if len(initialTasks) == 0 {
		return nil, fmt.Errorf("missing: starting tasks: request %s, template %s", res.ID, res.Template.Name)
	}
	if err = createRequestTask(ctx, res, initialTasks); err != nil {
		return nil, fmt.Errorf("fail: create task: %s, %s", res.ID, err.Error())
	}
	return res, nil
}

type CancelPayload struct {
	ID     string
	Reason string
}

func (obj *Request) Cancel(ctx *kaos.Context, payload *CancelPayload) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: db")
	}

	req, _ := datahub.GetByID(h, new(flowmodel.Request), payload.ID)
	if req == nil {
		return "", fmt.Errorf("missing: request")
	}
	if req.Status != flowmodel.RequestRunning {
		return "", fmt.Errorf("invalid: request status: %s", req.Status)
	}

	// get all request task and cancel it
	tasks, _ := datahub.FindByFilter(h, new(flowmodel.RequestTask), dbflex.Eqs("RequestID", req.ID, "Version", req.Version))
	for _, task := range tasks {
		task.Error = "Request is cancelled"
		task.Status = flowmodel.TaskCancel
		if err := h.Save(task); err != nil {
			return "", err
		}
	}

	// cancel the request
	req.Status = flowmodel.RequestCancel
	req.Reason = payload.Reason
	if err := h.Save(req, "Status", "Reason"); err != nil {
		return "", err
	}

	return "OK", nil
}

type ReviewPayload struct {
	TaskID   string
	Approval bool
	Reason   string
}

func (obj *Request) Review(ctx *kaos.Context, payload *ReviewPayload) (*flowmodel.RequestTask, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	task, err := datahub.GetByID(h, new(flowmodel.RequestTask), payload.TaskID)
	if err != nil {
		return nil, errors.New("missing: task")
	}
	if err = reviewTask(ctx, task, payload.Approval, payload.Reason); err != nil {
		return nil, err
	}
	return task, nil
}
