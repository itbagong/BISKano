package hcmlogic

import (
	"errors"
	"fmt"
	"io"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/codekit"
)

type AuthHandler struct {
}

type SignUpPayload struct {
	CompanyID string
	Email     string
	Password  string
}

func (m *AuthHandler) SignUp(ctx *kaos.Context, payload *SignUpPayload) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", errors.New("nil: EventHub")
	}

	userReq := rbaclogic.GetUserByRequest{
		FindBy: "email",
		FindID: payload.Email,
	}
	existsCandidate := new(rbacmodel.User)
	err := ev.Publish("/v1/iam/user/get-by", &userReq, existsCandidate, nil)
	if err != nil && err.Error() != io.EOF.Error() {
		return "", fmt.Errorf("error when check user: %s", err.Error())
	}

	if existsCandidate.ID != "" {
		return nil, fmt.Errorf("email already exists")
	}

	employee := tenantcoremodel.Employee{
		Email:          payload.Email,
		EmploymentType: "CANDIDATE",
		IsLogin:        true,
		IsActive:       true,
		CompanyID:      payload.CompanyID,
	}
	ctx.Data().Set("jwt_data", codekit.M{"CompanyID": payload.CompanyID})
	tenantcorelogic.MWPreAssignSequenceNo("Employee", false, "_id")(ctx, &employee)

	if e := h.GetByID(new(tenantcoremodel.Employee), employee.ID); e != nil {
		if e := h.Insert(&employee); e != nil {
			return nil, errors.New("error insert Employee: " + e.Error())
		}
	} else {
		if e := h.Save(&employee); e != nil {
			return nil, errors.New("error update Employee: " + e.Error())
		}
	}

	req := rbaclogic.CreateUserRequest{
		User: &rbacmodel.User{
			ID:          employee.ID,
			Email:       payload.Email,
			LoginID:     payload.Email,
			DisplayName: payload.Email,
			Status:      "Registered",
			Enable:      true,
		},
		Password: payload.Password,
	}

	userid := ""
	err = ev.Publish("/v1/iam/user/create", &req, &userid, nil)
	if err != nil {
		return "", fmt.Errorf("error when create user: %s", err.Error())
	}

	roleReq := rbacmodel.RoleMember{
		UserID:         employee.ID,
		LoginID:        employee.Email,
		Scope:          rbacmodel.RoleScopeGlobal,
		RoleID:         "Candidate",
		RoleName:       "Candidate",
		DimensionScope: rbacmodel.DimensionNone,
	}
	memberID := ""
	err = ev.Publish("/v1/admin/add-user-to-role", &roleReq, &memberID, nil)
	if err != nil {
		return "", fmt.Errorf("error when add user to role member: %s", err.Error())
	}

	return "success", nil
}
