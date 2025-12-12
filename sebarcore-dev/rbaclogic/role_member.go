package rbaclogic

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func MWPostRoleMemberGets(ctx *kaos.Context, payload interface{}) (bool, error) {
	h, _ := ctx.DefaultHub()

	res, roleMembers, err := sebar.ExtractDbmodGetsResult(ctx, rbacmodel.RoleMember{})
	if err != nil {
		return true, nil
	}

	tenantIDs := lo.Map(roleMembers, func(obj rbacmodel.RoleMember, index int) interface{} {
		return obj.TenantID
	})
	tenants, _ := datahub.FindByFilter(h, &rbacmodel.Tenant{}, dbflex.In("_id", tenantIDs...))
	tenantNameMap := lo.Associate(tenants, func(obj *rbacmodel.Tenant) (string, string) {
		return obj.ID, obj.Name
	})
	for index, obj := range roleMembers {
		name := tenantNameMap[obj.TenantID]
		if name != "" {
			obj.TenantID = name
		}
		roleMembers[index] = obj
	}
	res.Set("data", roleMembers)
	return true, nil
}
