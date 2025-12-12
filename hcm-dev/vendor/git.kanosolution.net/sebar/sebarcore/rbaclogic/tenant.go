package rbaclogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type TenantLogic struct {
}

func (obj *TenantLogic) My(ctx *kaos.Context, payload string) ([]rbacmodel.Tenant, error) {
	res := []rbacmodel.Tenant{}

	h, _ := ctx.DefaultHub()
	if h == nil {
		return res, errors.New("db_conn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return res, errors.New("invalid: UserID")
	}

	tenantUsers := []rbacmodel.TenantUser{}
	where := dbflex.Eq("UserID", userID)
	err := h.Gets(new(rbacmodel.TenantUser), dbflex.NewQueryParam().SetWhere(where).SetSelect("TenantID"), &tenantUsers)
	if err != nil {
		return res, ctx.Log().Error2("unable to get data", "tenantlogic/my: get tenant users: %s", err.Error())
	}

	tenantIDs := lo.Map(tenantUsers, func(tu rbacmodel.TenantUser, index int) interface{} {
		return tu.TenantID
	})
	where = dbflex.In("_id", tenantIDs...)
	err = h.Gets(new(rbacmodel.Tenant), dbflex.NewQueryParam().SetWheres(where, dbflex.Eq("Enable", true)).SetSelect("_id", "FID", "Name", "Use2FA"), &res)
	if err != nil {
		return res, ctx.Log().Error2("unable to get data", "tenantlogic/my: get tenant users: %s", err.Error())
	}

	return res, nil
}

func (obj *TenantLogic) Create(ctx *kaos.Context, payload *rbacmodel.Tenant) (*rbacmodel.Tenant, error) {
	res := payload
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("db_conn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return res, errors.New("invalid: UserID")
	}

	res.Enable = false
	res.OwnerID = userID
	if err := h.Save(res); err != nil {
		return nil, ctx.Log().Error2("unable to create tenant", "save tenant: %s", err.Error())
	}

	tu := new(rbacmodel.TenantUser)
	tu.TenantID = res.ID
	tu.TenantName = res.Name
	tu.UserID = userID
	if err := h.Save(tu); err != nil {
		return nil, ctx.Log().Error2("unable to create tenant", "save tenant user: %s", err.Error())
	}

	return res, nil
}

func (obj *TenantLogic) RequestToJoin(ctx *kaos.Context, payload *rbacmodel.TenantJoin) (*rbacmodel.TenantJoin, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missing: dbconn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return nil, errors.New("missing: user")
	}
	user := &rbacmodel.User{}
	err := h.GetByID(user, userID)
	if err != nil {
		return nil, errors.New("invalid: User")
	}

	where := dbflex.Or(dbflex.Eq("_id", payload.TenantID), dbflex.Eq("FID", payload.TenantID))
	tenant := &rbacmodel.Tenant{}
	if err = h.GetByFilter(tenant, where); err != nil {
		return nil, errors.New("invalid: Tenant")
	}

	where = dbflex.Eqs("Status", "PENDING", "UserID", userID, "TenantID", tenant.ID)
	tenantJoin := &rbacmodel.TenantJoin{}
	if err = h.GetByFilter(tenantJoin, where); err == nil {
		return nil, errors.New("duplicate: pending request")
	}

	tenantJoin.JoinType = rbacmodel.TenantJoinRequest
	tenantJoin.UserID = userID
	tenantJoin.TenantID = tenant.ID
	tenantJoin.Status = "PENDING"
	if err = h.Save(tenantJoin); err != nil {
		return nil, errors.New("unable to join")
	}

	dataKit := codekit.M{"DisplayName": user.DisplayName, "TenantName": tenant.Name}
	if _, err := sebarcore.SendEmail(ctx, user.Email, "rbac-tenant-join-request", "en-us", dataKit); err != nil {
		return nil, err
	}

	return tenantJoin, nil
}

func (obj *TenantLogic) AddUser(ctx *kaos.Context, payload *rbacmodel.TenantJoin) (string, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return "", errors.New("missing: db")
	}

	tenant, err := datahub.Get(h, &rbacmodel.Tenant{ID: payload.TenantID})
	if err != nil {
		return "", errors.New("missing: Tenant")
	}
	ownerID := ctx.Data().Get("jwt_reference_id", "").(string)
	if ownerID != tenant.OwnerID {
		return "", errors.New("unauthorized")
	}

	user, err := datahub.Get(h, &rbacmodel.User{ID: payload.UserID})
	if err != nil {
		return "", errors.New("missing: User")
	}

	tu, err := datahub.GetByFilter(h, &rbacmodel.TenantUser{}, dbflex.Eqs("UserID", user.ID, "TenantID", tenant.ID))
	if err == nil {
		return "", errors.New("duplicate: Tenant User")
	}
	tu.UserID = user.ID
	tu.LoginID = user.LoginID
	tu.TenantID = tenant.ID
	tu.TenantName = tenant.Name
	if err := h.Save(tu); err != nil {
		return "", ctx.Log().Error2("unable to add user to tenant", "add user to tenant: %s", err.Error())
	}

	dataKit := codekit.M{"DisplayName": user.DisplayName, "TenantName": tenant.Name}
	if _, err := sebarcore.SendEmail(ctx, user.Email, "rbac-tenant-join", "en-us", dataKit); err != nil {
		return "Failed", err
	}

	return "OK", nil
}

func MWPostTenantGets(ctx *kaos.Context, payload interface{}) (bool, error) {
	res, tenants, err := sebar.ExtractDbmodGetsResult(ctx, rbacmodel.Tenant{})
	if err != nil {
		ctx.Log().Error(err.Error())
		return true, nil
	}

	userIDs := lo.Map(tenants, func(obj rbacmodel.Tenant, i int) interface{} {
		return obj.OwnerID
	})

	h, _ := ctx.DefaultHub()
	users := []rbacmodel.User{}
	h.GetsByFilter(&rbacmodel.User{}, dbflex.In("_id", userIDs...), &users)
	displayNames := lo.Associate(users, func(user rbacmodel.User) (string, string) {
		return user.ID, user.DisplayName
	})
	for index, tenant := range tenants {
		tenant.OwnerID = displayNames[tenant.OwnerID]
		tenants[index] = tenant
	}

	res["data"] = tenants
	ctx.Data().Set("FnResult", res)

	return true, nil
}
