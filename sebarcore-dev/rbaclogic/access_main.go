package rbaclogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type GetAccesType string

const (
	GetAccessByRole    GetAccesType = "Role"
	GetAccessByFeature GetAccesType = "Feature"
)

type GetAccessRequest struct {
	AccessType GetAccesType
	AccessID   string
	Scope      string
	Dim        rbacmodel.Dimension
	Level      int
}

type GetAccessRespond struct {
	Level int
}

func (a *GetAccessRespond) CanRead() bool {
	return a.Level&1 == 1
}

func (a *GetAccessRespond) CanCreate() bool {
	return a.Level&2 == 2
}

func (a *GetAccessRespond) CanUpdate() bool {
	return a.Level&4 == 4
}

func (a *GetAccessRespond) CanDelete() bool {
	return a.Level&8 == 8
}

func (a *GetAccessRespond) CanPost() bool {
	return a.Level&16 == 16
}

func (a *GetAccessRespond) CanS1() bool {
	return a.Level&32 == 32
}

func (a *GetAccessRespond) CanS2() bool {
	return a.Level&64 == 64
}

func (a *GetAccessRespond) CanS3() bool {
	return a.Level&128 == 128
}

type AccessLogic struct {
}

func (obj *AccessLogic) Get(ctx *kaos.Context, payload *GetAccessRequest) (*GetAccessRespond, error) {
	r := new(GetAccessRespond)

	rbac, err := GetRbacFromCtx(ctx, payload.Scope)
	if err != nil {
		return nil, err
	}

	level, ok := CheckAccess(rbac, payload)
	if !ok {
		return nil, fmt.Errorf("unauthorized: %s", codekit.JsonString(payload))
	}

	if level < payload.Level {
		return nil, fmt.Errorf("unauthorized: %s", codekit.JsonString(payload))
	}

	r.Level = level
	return r, nil
}

func CheckAccess(um *rbacmodel.UserMatrix, request *GetAccessRequest) (int, bool) {
	if um == nil || request == nil {
		return 0, false
	}

	switch request.AccessType {
	case GetAccessByRole:
		ok := um.MemberOf(request.AccessID)
		if ok {
			return 255, true
		} else {
			return 0, false
		}

	case GetAccessByFeature:
		level := um.AccessLevel(request.AccessID, request.Dim)
		if level == 0 {
			return 0, false
		}

		validateLevel := level & request.Level
		return validateLevel, validateLevel > 0
	}

	return 0, false
}

func GetUserMatrix(db *datahub.Hub, userID, tenantID string) (*rbacmodel.UserMatrix, error) {
	um := new(rbacmodel.UserMatrix)
	userAccess := rbacmodel.UserAccess{}

	user, err := datahub.GetByID(db, new(rbacmodel.User), userID)
	if err != nil {
		return nil, fmt.Errorf("missing: user: %s", userID)
	}

	rms, err := datahub.FindByFilter(db, new(rbacmodel.RoleMember), dbflex.Eq("UserID", userID))
	if err != nil {
		return nil, fmt.Errorf("get access role error: %s", err.Error())
	}
	rms = lo.Filter(rms, func(rm *rbacmodel.RoleMember, index int) bool {
		return rm.Scope == rbacmodel.RoleScopeGlobal || (rm.Scope == rbacmodel.RoleScopeTenant && rm.TenantID == tenantID)
	})
	roles := lo.Map(lo.FindUniquesBy(rms, func(rm *rbacmodel.RoleMember) string {
		return rm.RoleID
	}), func(rm *rbacmodel.RoleMember, index int) string {
		return rm.RoleID
	})

	for _, rm := range rms {
		dim := rbacmodel.Dimension{}

		switch rm.DimensionScope {
		case rbacmodel.DimensionUser:
			dim = user.Dimension

		case rbacmodel.DimensionCustom:
			dim = rm.Dimension
		}

		rfs, _ := datahub.FindByFilter(db, new(rbacmodel.RoleFeature), dbflex.Eq("RoleID", rm.RoleID))
		for _, rf := range rfs {
			a, ok := userAccess.Get(rf.FeatureID, dim)
			if ok && a.Level < rf.Level() {
				a.Level = rf.Level()
			} else if !ok {
				userAccess.Update(rf.FeatureID, dim, rf.Level())
			}
		}
	}

	um.Access = userAccess
	um.RoleIDs = roles

	return um, nil
}

func GetRbacDimFilter(ctx *kaos.Context, featureID, scope string) *dbflex.Filter {
	rbac, err := GetRbacFromCtx(ctx, scope)
	if err != nil {
		return nil
	}

	wheres := []*dbflex.Filter{}
	for _, ac := range rbac.Access {
		if ac.FeatureID == featureID {
			if len(ac.Dimension) == 0 {
				return nil
			}

			wheres = append(wheres, ac.Dimension.DbWhere())
		}
	}

	if len(wheres) == 0 {
		return nil
	}

	return dbflex.Or(wheres...)
}

func GetRbacFromCtx(ctx *kaos.Context, scope string) (*rbacmodel.UserMatrix, error) {
	var jwtData codekit.M
	if scope == "jwt" {
		jwtData = ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	} else if scope == "session" {
		jwtData = ctx.Data().Get("jwt_session_data", codekit.M{}).(codekit.M)
	} else {
		return nil, errors.New("invalid: access request source: valid opts only jwt or session")
	}
	rbac := new(rbacmodel.UserMatrix)
	err := serde.Serde(jwtData["RBAC"], rbac)
	if err != nil {
		return nil, errors.New("missing: user matrix")
	}
	return rbac, nil
}

func MWRbacFilterDim(featureID, scope string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		var jwtData codekit.M
		fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)
		if scope == "jwt" {
			jwtData = ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
		} else if scope == "session" {
			jwtData = ctx.Data().Get("jwt_session_data", codekit.M{}).(codekit.M)
		} else {
			return false, errors.New("invalid: access request source: valid opts only jwt or session")
		}

		// if featureID=="" read dimension from profile
		if featureID == "" {
			dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})
			if len(dimIface) == 0 {
				return true, nil
			}

			dim := rbacmodel.Dimension{}
			if err := serde.Serde(dimIface, &dim); err != nil {
				return false, fmt.Errorf("invalid: dimension access: %s", err)
			}

			fs = append(fs, dim.DbWhere())
			ctx.Data().Set("DBModFilter", fs)
			return true, nil
		}

		// else read from respective feature
		dimFilter := GetRbacDimFilter(ctx, featureID, scope)
		if dimFilter != nil {
			fs = append(fs, dimFilter)
			ctx.Data().Set("DBModFilter", fs)
		}

		return true, nil
	}
}
