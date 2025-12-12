package rbaclogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type TenantJoinLogic struct {
}

type TenantJoinApproval struct {
	RequestID string
	Approve   bool
}

func (obj *TenantJoinLogic) Approve(ctx *kaos.Context, payload *TenantJoinApproval) (string, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return "", errors.New("missing: db")
	}

	joinRequest, err := datahub.Get(h, &rbacmodel.TenantJoin{ID: payload.RequestID})
	if err != nil {
		return "", errors.New("invalid: Join request id")
	}

	tenant, err := datahub.Get(h, &rbacmodel.Tenant{ID: joinRequest.TenantID})
	if err != nil {
		return "", errors.New("invalid: Tenant ID")
	}
	currentUserID := ctx.Data().Get("jwt_reference_id", "").(string)
	if currentUserID != tenant.OwnerID {
		return "", errors.New("unauthorized")
	}

	user, err := datahub.Get(h, &rbacmodel.User{ID: joinRequest.UserID})
	if err != nil {
		return "", errors.New("invalid: User ID")
	}

	var kind string = "rbac-tenant-join"
	if !payload.Approve {
		kind = "rbac-tenant-join-rejected"
		joinRequest.Status = "REJECTED"
		h.Save(joinRequest)
	}

	tenantUser, err := datahub.GetByFilter(h, &rbacmodel.TenantUser{}, dbflex.Eqs("UserID", user.ID, "TenantID", tenant.ID))
	if err == nil {
		h.Delete(joinRequest)
		return "", nil
	}
	joinRequest.Status = "APPROVED"
	h.Save(joinRequest)

	tenantUser.UserID = user.ID
	tenantUser.LoginID = user.LoginID
	tenantUser.TenantID = tenant.ID
	tenantUser.TenantName = tenant.Name
	h.Save(tenantUser)

	// TODO: kirim email
	if _, mailError := sebarcore.SendEmail(ctx, user.Email, kind, "en-us", codekit.M{"DisplayName": user.DisplayName, "TenantName": tenant.Name}); mailError != nil {
		return "Failed", mailError
	}

	return "", nil
}

func MWPreTenantJoinReviewFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	userID, userIDIsOK := ctx.Data().Get("jwt_reference_id", "").(string)
	tenantData, tenantIDIsOK := ctx.Data().Get("jwt_session_data", nil).(codekit.M)
	if !userIDIsOK {
		return false, errors.New("missing: User")
	}
	if !tenantIDIsOK {
		return false, errors.New("missing: Tenant")
	}

	tenantID := tenantData.GetString("TenantID")
	parm, ok := payload.(*dbflex.QueryParam)
	if !ok {
		return false, fmt.Errorf("invalid: Payload, got %t", payload)
	}

	h, _ := ctx.DefaultHub()
	if h == nil {
		return false, errors.New("missing: db")
	}
	tenant, err := datahub.Get(h, &rbacmodel.Tenant{ID: tenantID})
	if err != nil {
		return false, errors.New("missing: Tenant")
	}
	if tenant.OwnerID != userID {
		return false, errors.New("unauthorized")
	}
	parm.MergeWhere(false, dbflex.Eq("TenantID", tenant.ID))

	return true, nil
}

func MWPostTenantJoinReviewFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	h, _ := ctx.DefaultHub()

	objs, err := sebar.ExtractDbmodFindResult(ctx, rbacmodel.TenantJoinExtended{})
	if err != nil {
		return true, nil
	}

	userIDs := lo.Map(objs, func(obj rbacmodel.TenantJoinExtended, index int) interface{} {
		return obj.UserID
	})
	users, err := datahub.FindByFilter(h, &rbacmodel.User{}, dbflex.In("_id", userIDs...))
	if err != nil {
		ctx.Log().Errorf("mod MWPostTenantJoinReviewFind error: %s", err.Error())
		return true, nil
	}
	displayNameMaps := lo.Associate(users, func(obj *rbacmodel.User) (string, string) {
		return obj.ID, obj.DisplayName
	})

	for i, obj := range objs {
		displayName := displayNameMaps[obj.UserID]
		obj.DisplayName = displayName
		objs[i] = obj
	}
	ctx.Data().Set("FnResult", &objs)
	return true, nil
}

func MWPreTenantJoinRequestFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	userid := ctx.Data().Get("jwt_reference_id", "").(string)
	if userid == "" {
		return false, errors.New("missing: session")
	}

	var (
		parm *dbflex.QueryParam
		ok   bool
	)
	if payload == nil {
		parm = dbflex.NewQueryParam()
	} else {
		parm, ok = payload.(*dbflex.QueryParam)
		if !ok {
			ctx.Log().Errorf("midware error: MWPreTenantJoinRequestFind: invalid payload type, got %t", payload)
			return false, errors.New("invalid: payload")
		}
	}
	parm.SetWhere(dbflex.Eq("UserID", userid))

	return true, nil
}

func MWPostTenantJoinRequestFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	h, _ := ctx.DefaultHub()

	objs, err := sebar.ExtractDbmodFindResult(ctx, rbacmodel.TenantJoinExtended{})
	if err != nil {
		return true, nil
	}

	tenantIDs := lo.Map(objs, func(obj rbacmodel.TenantJoinExtended, index int) interface{} {
		return obj.TenantID
	})
	tenants, _ := datahub.FindByFilter(h, &rbacmodel.Tenant{}, dbflex.In("_id", tenantIDs...))
	tenantNames := lo.Associate(tenants, func(obj *rbacmodel.Tenant) (string, string) {
		return obj.ID, obj.Name
	})

	for i, obj := range objs {
		obj.TenantName = tenantNames[obj.TenantID]
		objs[i] = obj
	}
	ctx.Data().Set("FnResult", &objs)
	return true, nil
}
