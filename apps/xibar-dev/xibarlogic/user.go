package xibarlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/kmsg"
	"github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/sebarcode/codekit"
)

type User struct {
	appConfig  *sebar.AppConfig
	authOption *rbaclogic.AuthOptions
}

func NewUserLogic(appConfig *sebar.AppConfig, opt *rbaclogic.AuthOptions) *User {
	return &User{
		appConfig,
		opt,
	}
}

type FindUserByRespond struct {
	ID          string `json:"_id" bson:"_id"`
	LoginID     string
	DisplayName string
	Email       string
}

func (u *User) FindByQuery(ctx *kaos.Context, payload *dbflex.QueryParam) ([]FindUserByRespond, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("db_conn")
	}

	res := []FindUserByRespond{}
	if payload.Take == 0 || payload.Take > 100 {
		payload.Take = 20
	}
	user := new(rbacmodel.User)
	var (
		err error
	)

	err = h.Gets(user, payload, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (u *User) ValidateTwofa(ctx *kaos.Context, payload string) (string, error) {

	h, _ := ctx.GetHub("iam", "")
	if h == nil {
		return "", errors.New("db_conn")
	}

	checkName := "LoginID"

	userToken := &rbacmodel.Token{}
	whereUserToken := dbflex.Eqs("Token", payload, "Kind", "rbac-user-2fa", "Status", "Reserved")
	h.GetByParm(userToken, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(whereUserToken))

	if userToken.Token != "" {
		userToken.Status = "Claimed"
		userToken.ClaimDate = time.Now()
		h.Save(userToken)
		return checkName, nil
	}

	tenantToken := &rbacmodel.Token{}
	whereTenantToken := dbflex.Eqs("Token", payload, "Kind", "rbac-tenant-2fa", "Status", "Reserved")
	h.GetByParm(tenantToken, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(whereTenantToken))

	if tenantToken.Token != "" {
		tenantToken.Status = "Claimed"
		tenantToken.ClaimDate = time.Now()
		h.Save(tenantToken)
		checkName = "Tenant"
		return checkName, nil
	}

	return "", errors.New("invalid token")
}

func (u *User) PublicResetPassword(ctx *kaos.Context, payload *RequestResetPasswordRequest) (string, error) {
	h, err := ctx.DefaultHub()
	if err != nil {
		return "", err
	}

	token := &rbacmodel.Token{}

	whereUser := dbflex.Eq("Email", payload.Email)
	user := &rbacmodel.User{}
	err = h.GetByParm(user, dbflex.NewQueryParam().SetWhere(whereUser))
	if err != nil {
		return "", err
	}

	where := dbflex.Eqs("UserID", user.ID, "Kind", "rbac-user-reset-password", "Status", "Reserved")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))

	if token.ID == "" {
		token.Token = codekit.GenerateRandomString("", 16)
		token.Status = "Reserved"
		token.UserID = user.ID
		token.Kind = "rbac-user-reset-password"
		token.App = "Xibar"
		token.Expiry = time.Now().Add(5 * time.Minute)
	} else {
		if time.Now().Before(token.Expiry) {
			return "", errors.New("please wait before trying to reset again")
		}
	}

	err = h.Save(token)
	if err != nil {
		return "", err
	}

	linkActivation := u.appConfig.Data.GetString("addr_web_user") + "/me/reset-password"
	msgTemplate := ksmsg.SendTemplateRequest{
		TemplateName: "reset_password",
		LanguageID:   "en-us",
		Message: &kmsg.Message{
			Kind:   "reset_password",
			Method: "SMTP",
			To:     user.Email,
		},
		Data: codekit.M{
			"DisplayName":       user.DisplayName,
			"ResetPasswordLink": linkActivation,
			"Token":             token.Token,
		},
	}
	ev, e := ctx.DefaultEvent()
	if e != nil {
		return "", e
	}
	resp := 0
	e = ev.Publish("/v1/msg/send-template", &msgTemplate, &resp, nil)
	if e != nil {
		return "NOK", e
	}

	return "", nil
}

type RequestResetPasswordRequest struct {
	Email string
	AppID string
}

func (u *User) RequestResetPassword(ctx *kaos.Context, payload *RequestResetPasswordRequest) (string, error) {
	h, err := ctx.DefaultHub()
	if err != nil {
		return "", err
	}

	token := &rbacmodel.Token{}
	// fmt.Println(ctx.Data().Keys())
	jwtToken := ctx.Data().Get("jwt_token", "").(string)

	if jwtToken == "" {
		return "", errors.New("Invalid Session")
	}

	_, data, err := rbaclogic.JwtToSession(jwtToken, u.appConfig.Data.GetString("jwt_secret"))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	// fmt.Println(data)

	whereUser := dbflex.Eq("Email", data.GetString("Email"))
	user := &rbacmodel.User{}
	err = h.GetByParm(user, dbflex.NewQueryParam().SetWhere(whereUser))
	if err != nil {
		// fmt.Println("Error querying user")
		return "", err
	}

	where := dbflex.Eqs("UserID", user.ID, "Kind", "rbac-reset-password", "Status", "Reserved")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))

	if token.ID == "" {
		//return "", errors.New("not-found: Token")
		// create token
		token.Token = codekit.GenerateRandomString("", 16)
		token.Status = "Reserved"
		token.UserID = user.ID
		token.Kind = "rbac-reset-password"
		token.App = "Xibar"
		token.Expiry = time.Now().Add(2 * time.Hour)
	}
	err = h.Save(token)
	if err != nil {
		return "", err
	}

	linkActivation := u.appConfig.Data.GetString("addr_web_user") + "/me/reset-password?uid=" + token.Token + "&AppID=" + payload.AppID

	msgTemplate := ksmsg.SendTemplateRequest{
		TemplateName: "reset_password",
		LanguageID:   "en-us",
		Message: &kmsg.Message{
			Kind:   "reset_password",
			Method: "SMTP",
			To:     user.Email,
		},
		Data: codekit.M{
			"DisplayName":       user.DisplayName,
			"ResetPasswordLink": linkActivation,
			"Token":             token.Token,
		},
	}
	ev, e := ctx.DefaultEvent()
	if e != nil {
		return "", e
	}
	resp := 0
	e = ev.Publish("/v1/msg/send-template", &msgTemplate, &resp, nil)
	if e != nil {
		return "NOK", e
	}
	// fmt.Println(resp)
	return "", nil
}

type ResetPasswordRequest struct {
	Token    string
	Password string
}

func (u *User) ResetPassword(ctx *kaos.Context, payload *ResetPasswordRequest) (codekit.M, error) {
	h, err := ctx.DefaultHub()
	if err != nil {
		return nil, err
	}
	result := codekit.M{}
	token := &rbacmodel.Token{}
	where := dbflex.Eqs("Token", payload.Token, "Kind", "rbac-reset-password", "Status", "Reserved")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))

	if token.Token == "" {
		return nil, errors.New("invalid token")
	}

	token.Status = "Claimed"
	token.ClaimDate = time.Now()
	userPwd := &rbacmodel.UserPassword{}
	whereUser := dbflex.Eq("_id", token.UserID)
	// fmt.Println("UID", token.UserID)
	err = h.GetByParm(userPwd, dbflex.NewQueryParam().SetWhere(whereUser))
	if err != nil {
		return nil, err
	}
	// fmt.Println(userPwd)
	userPwd.Password = codekit.ShaString(payload.Password, "")
	h.BeginTx()
	h.Save(token)
	h.Save(userPwd)
	h.Commit()
	return result, nil
}
