package rbaclogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type UserMeLogic struct {
}

func (obj *UserMeLogic) Activate(ctx *kaos.Context, payload string) (codekit.M, error) {
	res := codekit.M{}

	h, _ := ctx.DefaultHub()
	if h == nil {
		return res, errors.New("db_conn")
	}

	// get user
	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return res, errors.New("invalid: UserID")
	}

	user := rbacmodel.User{}
	if err := h.GetByID(&user, userID); err != nil {
		return res, errors.New("invalid: UserID")
	}

	if user.Status != "Registered" {
		return res, errors.New("invalid: User.Status")
	}

	//-- get token
	token := &rbacmodel.Token{}
	where := dbflex.Eqs("UserID", userID, "Kind", "rbac-new-user", "Status", "Reserved")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))
	if token.ID == "" {
		return res, errors.New("not-found: Token")
	}

	if token.Token != payload {
		return res, errors.New("invalid: Token")
	}

	if token.Expiry.Before(time.Now()) {
		return res, errors.New("expired: Token")
	}

	token.Status = "Claimed"
	token.ClaimDate = time.Now()
	h.Save(token)

	user.Status = "Active"
	h.Update(&user, "Status")
	sebarcore.SendEmail(ctx, user.Email, "rbac-user-activated", "en-us", codekit.M{"Email": user.Email})

	return res, nil
}

func (obj *UserMeLogic) ResendActivationEmail(ctx *kaos.Context, payload string) (codekit.M, error) {
	res := codekit.M{}

	h, _ := ctx.DefaultHub()
	if h == nil {
		return res, errors.New("db_conn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return res, errors.New("invalid: UserID")
	}

	user := rbacmodel.User{}
	if err := h.GetByID(&user, userID); err != nil {
		return res, errors.New("invalid: UserID")
	}

	if user.Status != "Registered" {
		return res, errors.New("invalid: User.Status")
	}

	//-- get token
	token := &rbacmodel.Token{}
	where := dbflex.Eqs("UserID", userID, "Kind", "rbac-new-user", "Status", "Reserved", "Enable", true)
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))
	if token.ID == "" {
		return res, errors.New("not-found: Token")
	}

	//-- send the token
	addrWebUser := ctx.Data().Get("service_addr_web_user", "").(string)
	addrActivationUserLink := ctx.Data().Get("service_addr_web_user_activation", "").(string)
	if addrWebUser == "" || addrActivationUserLink == "" {
		return res, errors.New("mandatory: activation link")
	}
	activationLink := fmt.Sprintf("%s%s", addrWebUser, addrActivationUserLink)

	//-- send email
	_, err := sebarcore.SendEmail(ctx, user.Email, "rbac-new-user", "en-us", codekit.M{
		"Email":          user.Email,
		"ActivationCode": token.Token,
		"ActivationLink": activationLink,
	})
	if err != nil {
		return res, fmt.Errorf("fail send email: %s", err.Error())
	}

	return res, nil
}

func (obj *UserMeLogic) ChangeTenant(ctx *kaos.Context, payload *rbacmodel.Tenant) (*AuthResponse, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missing: dbconn")
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("missing: event")
	}

	// Two Factor Authentication (2FA)
	if payload.Use2FA {
		userID := ctx.Data().Get("jwt_reference_id", "").(string)
		if userID == "" {
			return nil, errors.New("invalid: UserID")
		}

		user := rbacmodel.User{}
		if err := h.GetByID(&user, userID); err != nil {
			return nil, errors.New("invalid: UserID")
		}

		kind2FA := "rbac-tenant-2fa"
		_, err := set2FA(ctx, &user, h, kind2FA)
		if err != nil {
			return nil, err
		}
	}

	request := rbacmodel.AuthChangeDataRequest{
		Scope: rbacmodel.AuthScopeIsBoth,
		Token: ctx.Data().Get("jwt_token", "").(string),
		Data: codekit.M{
			"TenantID":   payload.ID,
			"TenantName": payload.Name,
			"CheckName":  "Tenant",
		},
	}
	if request.Token == "" {
		return nil, errors.New("missing: jwt token")
	}

	topic := ctx.Data().Get("service_topic_change_session_data", "").(string)
	resp := AuthResponse{}
	err := ev.Publish(topic, &request, &resp, nil)
	if err != nil {
		return nil, ctx.Log().Errorf("fail to change session data", "change-tenant: change session data: %s", err.Error())
	}
	return &resp, nil
}

func (obj *UserMeLogic) Apps(ctx *kaos.Context, payload string) ([]*rbacmodel.App, error) {
	res := []*rbacmodel.App{}

	h, _ := ctx.DefaultHub()
	if h == nil {
		return res, errors.New("missingDBConn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	tenantID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("TenantID")

	if userID == "" {
		return res, errors.New("missing: User")
	}

	//-- get public tenants
	res, _ = datahub.FindByFilter(h, &rbacmodel.App{}, dbflex.Eqs("Public", true, "Enable", true))
	if tenantID == "" {
		return res, nil
	}

	tenantApps, _ := datahub.FindByFilter(h, &rbacmodel.TenantApp{}, dbflex.Eq("TenantID", tenantID))
	for _, tenantApp := range tenantApps {
		app, err := datahub.Get(h, &rbacmodel.App{ID: tenantApp.AppID})
		if err != nil {
			continue
		}
		if !app.Enable {
			continue
		}
		if !app.UseRoleToAccess {
			res = append(res, app)
			continue
		}
		if app.RoleID == "" {
			continue
		}
		if _, err := datahub.GetByFilter(h, &rbacmodel.RoleMember{}, dbflex.Eqs("UserID", userID, "RoleID", app.RoleID)); err == nil {
			res = append(res, app)
		}
	}

	return res, nil
}
