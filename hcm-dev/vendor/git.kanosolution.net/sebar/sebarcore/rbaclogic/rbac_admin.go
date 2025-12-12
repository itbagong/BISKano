package rbaclogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	. "git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type adminEngine struct {
	hub *datahub.Hub
	ev  kaos.EventHub
}

func NewAdminEngine(h *datahub.Hub, ev kaos.EventHub) *adminEngine {
	eng := new(adminEngine)
	eng.hub = h
	eng.ev = ev
	return eng
}

type ChangeUserPasswordRequest struct {
	UserID   string
	Password string
}

func (obj *adminEngine) ChangePassword(ctx *kaos.Context, payload *ChangeUserPasswordRequest) (string, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return "", errors.New("missing: db")
	}
	return obj.ChangeUserPassword(h, payload)
}

func (obj *adminEngine) ChangeUserPassword(h *datahub.Hub, payload *ChangeUserPasswordRequest) (string, error) {
	if h == nil {
		return "", errors.New("missing: db")
	}

	if e := h.GetByID(new(User), payload.UserID); e != nil {
		return "", errors.New("invalid user")
	}

	password := codekit.ShaString(payload.Password, "")
	up := UserPassword{ID: payload.UserID, Password: password}
	if e := h.Save(&up); e != nil {
		return "", fmt.Errorf("save password error. %s", e.Error())
	}

	return payload.UserID, nil
}

type EnableUserRequest struct {
	UserID string
	Enable bool
}

func (obj *adminEngine) EnableUser(payload *EnableUserRequest) (string, error) {
	h := obj.hub
	if h == nil {
		return "", errors.New("missing: db")
	}

	user := new(User)
	if e := h.GetByID(user, payload.UserID); e != nil {
		return "", errors.New("invalid user")
	}
	user.Enable = payload.Enable

	if e := h.Update(user, "Enable"); e != nil {
		return "", e
	}

	return payload.UserID, nil
}

type ChangeUserStatusRequest struct {
	UserID string
	Status string
}

func (obj *adminEngine) ChangeUserStatus(payload *ChangeUserStatusRequest) (string, error) {
	h := obj.hub
	if h == nil {
		return "", errors.New("missing: db")
	}

	user := new(User)
	if e := h.GetByID(user, payload.UserID); e != nil {
		return "", errors.New("invalid user")
	}
	user.Status = payload.Status

	if e := h.Update(user, "Status"); e != nil {
		return "", e
	}

	return payload.UserID, nil
}

type AddFeaturesToRoleRequest struct {
	RoleID   string
	Features []RoleFeature
}

func (obj *adminEngine) AddFeatureToRole(payload *RoleFeature) (string, error) {
	return obj.AddFeaturesToRole(&AddFeaturesToRoleRequest{
		RoleID:   payload.RoleID,
		Features: []RoleFeature{*payload},
	})
}

func (obj *adminEngine) AddFeaturesToRole(payload *AddFeaturesToRoleRequest) (string, error) {
	h := obj.hub
	if h == nil {
		return "", errors.New("missing: db")
	}

	h.DeleteQuery(new(RoleFeature), dbflex.Eq("RoleID", payload.RoleID))

	for _, feature := range payload.Features {
		f := dbflex.Eqs("RoleID", payload.RoleID, "FeatureID", feature.FeatureID)
		rf := RoleFeature{}
		if e := h.GetByFilter(&rf, f); e != nil {
			rf.RoleID = payload.RoleID
			rf.FeatureID = feature.FeatureID
		}
		rf.Create = feature.Create
		rf.Read = feature.Read
		rf.Update = feature.Update
		rf.Delete = feature.Delete
		rf.Posting = feature.Posting
		rf.Special1 = feature.Special1
		rf.Special2 = feature.Special2
		rf.All = feature.All
		if e := h.Save(&rf); e != nil {
			return "", e
		}
	}

	return "", nil
}

func (obj *adminEngine) AddUserToRole(payload *RoleMember) (string, error) {
	h := obj.hub
	if h == nil {
		return "", errors.New("missing: db")
	}

	rm := RoleMember{}
	if e := h.GetByFilter(&rm, dbflex.Eqs("UserID", payload.UserID, "RoleID", payload.RoleID,
		"Hash", payload.Dimension.Hash())); e == nil {
		return "", nil
	}
	e := h.Save(payload)

	return payload.ID, e
}
