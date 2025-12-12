package ficologic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/sebarcode/codekit"
)

type NotificationHandler struct {
}

type NotificationHandlerShowsRequest struct {
}

func (m *NotificationHandler) Shows(ctx *kaos.Context, param *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
		if ctx.Data().Get("UserID", "").(string) != "" {
			userID = ctx.Data().Get("UserID", "").(string)
		}
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	param = param.MergeWhere(false, []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
		dbflex.Eq("UserTo", userID),
	}...)

	notifs := []ficomodel.Notification{}
	err := h.Gets(new(ficomodel.Notification), param.SetSort("IsRead", "-Created"), &notifs)
	if err != nil {
		return nil, fmt.Errorf("error when get notification: %s", err.Error())
	}

	return codekit.M{"data": notifs}, nil
}

func (m *NotificationHandler) Count(ctx *kaos.Context, _ *NotificationHandler) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
		if ctx.Data().Get("UserID", "").(string) != "" {
			userID = ctx.Data().Get("UserID", "").(string)
		}
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	count, err := h.Count(new(ficomodel.Notification), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("CompanyID", coID),
			dbflex.Eq("UserTo", userID),
			dbflex.Eq("IsRead", false),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error when get notification: %s", err.Error())
	}

	return codekit.M{"count": count}, nil
}

type NotificationHandlerSaveRequest struct {
	ID string
}

func (m *NotificationHandler) Save(ctx *kaos.Context, p *NotificationHandlerSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	mdl := new(ficomodel.Notification)
	err := h.GetByID(mdl, p.ID)
	if err != nil {
		return nil, fmt.Errorf("error when get notification: %s", err.Error())
	}

	mdl.IsRead = true
	err = h.Save(mdl)
	if err != nil {
		return nil, fmt.Errorf("error when save notification: %s", err.Error())
	}

	return "success", nil
}
