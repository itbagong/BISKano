package rbaclogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
)

type PublicUserAPI struct {
}

type HttpCreateUserRequest struct {
	Email    string
	Password string
}

func (obj *PublicUserAPI) Create(ctx *kaos.Context, payload *HttpCreateUserRequest) (string, error) {
	req := CreateUserRequest{
		User: &rbacmodel.User{
			Email:       payload.Email,
			LoginID:     payload.Email,
			DisplayName: payload.Email,
			Status:      "Registered",
			Enable:      true,
		},
		Password: payload.Password,
	}
	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", errors.New("nil: EventHub")
	}
	userid := ""
	err := ev.Publish("/v1/iam/user/create", &req, &userid, nil)
	if err != nil {
		return "", err
	}
	return userid, nil
}
