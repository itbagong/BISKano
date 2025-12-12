package karalogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type UserProfile struct {
	appConfig *sebar.AppConfig
}

func NewUserLogic(appConfig *sebar.AppConfig, opt *rbaclogic.AuthOptions) *UserProfile {
	return &UserProfile{
		appConfig,
	}
}

func MWPreUserProfileFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return false, errors.New("db_conn")
	}

	tenantData, tenantIDIsOK := ctx.Data().Get("jwt_session_data", nil).(codekit.M)
	if !tenantIDIsOK {
		return false, errors.New("missing: Tenant")
	}

	tenantID := tenantData.GetString("TenantID")

	parm, ok := payload.(*dbflex.QueryParam)
	if !ok {
		return false, fmt.Errorf("invalid: Payload, got %t", payload)
	}
	members, _ := datahub.FindByFilter(h, &rbacmodel.RoleMember{},
		dbflex.And(dbflex.Eq("TenantID", tenantID), dbflex.Eq("RoleID", "kara-users")))

	userIDs := lo.Map(members, func(obj *rbacmodel.RoleMember, index int) interface{} {
		return obj.UserID
	})

	parm.MergeWhere(false, dbflex.In("_id", userIDs...))

	return true, nil
}
