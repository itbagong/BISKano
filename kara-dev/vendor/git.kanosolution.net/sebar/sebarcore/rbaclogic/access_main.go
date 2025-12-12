package rbaclogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
)

type GetAccessRequest struct {
	Scope      rbacmodel.RoleScope
	IsMemberOf string
}

type AccessLogic struct {
}

func (obj *AccessLogic) Get(ctx *kaos.Context, payload *GetAccessRequest) (string, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return "", errors.New("missingDBConn")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID == "" {
		return "", errors.New("unauthorized")
	}

	tenantID := ctx.Data().Get("jwt_tenant_id", "").(string)

	if payload.IsMemberOf != "" {
		role := new(rbacmodel.Role)
		if e := h.GetByID(role, payload.IsMemberOf); e != nil {
			return "", errors.New("unauthorized")
		}
		roleMember := rbacmodel.RoleMember{}
		var whereRM *dbflex.Filter
		if payload.Scope == rbacmodel.RoleScopeTenant {
			where1 := dbflex.Eqs("UserID", userID, "RoleID", payload.IsMemberOf, "TenantID", tenantID, "Scope", rbacmodel.RoleScopeTenant)
			where2 := dbflex.Eqs("UserID", userID, "RoleID", payload.IsMemberOf, "Scope", rbacmodel.RoleScopeGlobal)
			whereRM = dbflex.Or(where1, where2)
		} else {
			whereRM = dbflex.Eqs("UserID", userID, "RoleID", payload.IsMemberOf, "Scope", rbacmodel.RoleScopeGlobal)
		}
		if e := h.GetByFilter(&roleMember, whereRM); e != nil {
			return "", errors.New("unauthorized")
		}
	}

	return "OK", nil
}
