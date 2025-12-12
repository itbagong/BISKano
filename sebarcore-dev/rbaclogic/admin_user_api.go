package rbaclogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/sebarcode/codekit"
)

type AdminUserAPI struct {
}

func (obj *AdminUserAPI) ActivateUser(ctx *kaos.Context, payload string) (string, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return "", errors.New("db_conn")
	}

	// get user
	userID := payload
	if userID == "" {
		return "", errors.New("invalid: UserID")
	}

	user := rbacmodel.User{}
	if err := h.GetByID(&user, userID); err != nil {
		return userID, errors.New("invalid: UserID")
	}

	if user.Status != "Registered" {
		return userID, errors.New("invalid: User.Status")
	}

	user.Status = "Active"

	h.Update(&user, "Status")
	sebarcore.SendEmail(ctx, user.Email, "rbac-user-activated", "en-us", codekit.M{"Email": user.Email})

	return userID, nil
}

func (obj *AdminUserAPI) Activate(ctx *kaos.Context, payload string) (string, error) {
	userID, err := obj.ActivateUser(ctx, payload)
	return userID, err
}
