package rbaclogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	. "git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/sebarcode/codekit"
)

type CreateUserFn func(ctx *kaos.Context, req *CreateUserRequest) error

type RbacOptions struct {
	PreCreateUserFn  CreateUserFn
	PostCreateUserFn CreateUserFn
}

type publicUserEvent struct {
	opts *RbacOptions
}

func NewPublicUserEvent(opts *RbacOptions) *publicUserEvent {
	eng := new(publicUserEvent)
	eng.opts = opts
	if eng.opts == nil {
		eng.opts = &RbacOptions{}
	}
	return eng
}

type CreateUserRequest struct {
	User     *User
	Password string
}

func (obj *publicUserEvent) Create(ctx *kaos.Context, payload *CreateUserRequest) (string, error) {
	h, _ := ctx.GetHub("iam", "")
	if h == nil {
		return "", errors.New("missingDBConn")
	}

	otherUser := rbacmodel.User{}
	h.GetByFilter(&otherUser, dbflex.Eq("Email", payload.User.Email))
	if otherUser.ID != "" {
		return "", fmt.Errorf("duplicate: User: %s", payload.User.Email)
	}

	if obj.opts.PreCreateUserFn != nil {
		if e := obj.opts.PreCreateUserFn(ctx, payload); e != nil {
			return "", e
		}
	}

	payload.User.Enable = true
	if e := h.Save(payload.User); e != nil {
		return "", e
	}

	password := codekit.ShaString(payload.Password, "")
	up := UserPassword{ID: payload.User.ID, Password: password}
	if e := h.Save(&up); e != nil {
		return "", ctx.Log().Error2("unable to save password", "save password error. %s", e.Error())
	}

	postCreateUserFn := obj.opts.PostCreateUserFn
	if postCreateUserFn == nil {
		postCreateUserFn = DefaultPostCreateUser
	}
	if e := postCreateUserFn(ctx, &CreateUserRequest{User: payload.User, Password: password}); e != nil {
		h.DeleteAny(new(rbacmodel.UserPassword).TableName(), dbflex.Eq("_id", payload.User.ID))
		h.Delete(payload.User)
		return "", ctx.Log().Error2(
			"user has been created but post create function return error. Please contact system administrators",
			"post create user %s, %s error. %s", payload.User.ID, payload.User.LoginID, e.Error())
	}

	return payload.User.ID, nil
}

func (obj *publicUserEvent) Get(ctx *kaos.Context, payload string) (*User, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missingDBConn")
	}

	user := new(User)
	if err := h.GetByID(user, payload); err != nil {
		return nil, err
	}
	return user, nil
}

type GetUserByRequest struct {
	FindBy string
	FindID string
}

func (obj *publicUserEvent) GetBy(ctx *kaos.Context, payload *GetUserByRequest) (*User, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missingDBConn")
	}

	if payload == nil || payload.FindBy == "" {
		payload = &GetUserByRequest{FindBy: "userid"}
	}

	//-- check for http-req and query
	if req, ok := sebar.GetHTTPRequest(ctx); ok {
		if findBy := req.URL.Query().Get("findby"); findBy != "" {
			payload.FindBy = findBy
		}

		if findID := req.URL.Query().Get("findid"); findID != "" {
			payload.FindID = findID
		}
	}

	user := new(User)
	var err error
	switch strings.ToLower(payload.FindBy) {
	case "id", "_id", "userid":
		err = h.GetByID(user, payload.FindID)
	case "loginid":
		err = h.GetByFilter(user, dbflex.Eq("LoginID", payload.FindID))
	case "email":
		err = h.GetByFilter(user, dbflex.Eq("Email", payload.FindID))
	default:
		err = errors.New("invalid getby parameters")
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

type FindUserByRespond struct {
	ID          string `json:"_id" bson:"_id"`
	LoginID     string
	DisplayName string
	Email       string
}

type FindUserByRequest struct {
	FindBy         string
	FindID         string
	IncludeDisable bool
	Take           int
	Where          dbflex.Filter
}

func (obj *publicUserEvent) FindBy(ctx *kaos.Context, payload *FindUserByRequest) ([]FindUserByRespond, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("db_conn")
	}

	if payload == nil || payload.FindBy == "" && payload.Where.Field == "" {
		payload = &FindUserByRequest{FindBy: "userid", Take: 20}
	} else if payload.FindBy == "" && payload.Where.Field != "" {
		payload = &FindUserByRequest{FindBy: "loginid", Take: payload.Take, FindID: setFindID(payload.Where.Value)}
	}

	//-- check for http-req and query
	//-- check for http-req and query
	if req, ok := sebar.GetHTTPRequest(ctx); ok {
		if findBy := req.URL.Query().Get("findby"); findBy != "" {
			payload.FindBy = findBy
		}

		if findID := req.URL.Query().Get("findid"); findID != "" {
			payload.FindID = findID
		}
	}

	res := []FindUserByRespond{}
	if payload.Take == 0 || payload.Take > 100 {
		payload.Take = 20
	}
	user := new(User)
	var (
		filters = []*dbflex.Filter{}
		err     error
	)
	if !payload.IncludeDisable {
		filters = append(filters, dbflex.Eq("Enable", true))
	}

	switch strings.ToLower(payload.FindBy) {
	case "id", "_id", "userid":
		filters = append(filters, dbflex.StartWith("_id", payload.FindID))
	case "loginid":
		filters = append(filters, dbflex.StartWith("LoginID", payload.FindID))
	case "email":
		filters = append(filters, dbflex.StartWith("Email", payload.FindID))
	default:
		return nil, errors.New("invalid getby parameters")
	}
	err = h.Gets(user, dbflex.NewQueryParam().SetWhere(dbflex.And(filters...)).SetTake(payload.Take), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func setFindID(o interface{}) string {
	aInterface := o.([]interface{})
	aString := make([]string, len(aInterface))
	for i, v := range aInterface {
		aString[i] = v.(string)
	}
	return strings.Join(aString, " ")
}

func DefaultPostCreateUser(ctx *kaos.Context, req *CreateUserRequest) error {
	return sendUserActivation(ctx, req.User)
}

func sendUserActivation(ctx *kaos.Context, user *rbacmodel.User) error {
	if user.ID == "" {
		return errors.New("mandatory: User.ID")
	}

	//-- create activation token
	addrWebUser := ctx.Data().Get("service_addr_web_user", "").(string)
	addrActivationUserLink := ctx.Data().Get("service_addr_web_user_activation", "").(string)
	if addrWebUser == "" || addrActivationUserLink == "" {
		return errors.New("mandatory: activation link")
	}
	activationLink := fmt.Sprintf("%s%s", addrWebUser, addrActivationUserLink)

	tokenEng := NewTokenEngine(nil, 16)
	tokenCode, err := tokenEng.Create(ctx, &CreateTokenRequest{
		App:            "Xibar",
		Kind:           "rbac-new-user",
		UserID:         user.ID,
		ExpiryInSecond: 7 * 24 * 60 * 60,
	})
	if err != nil {
		return err
	}

	//-- send email
	_, err = sebarcore.SendEmail(ctx, user.Email, "rbac-new-user", "en-us", codekit.M{
		"Email":          user.Email,
		"ActivationCode": tokenCode,
		"ActivationLink": activationLink,
	})
	return err
}
