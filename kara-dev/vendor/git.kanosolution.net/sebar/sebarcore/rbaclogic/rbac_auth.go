package rbaclogic

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/golang-jwt/jwt"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/siam"
)

type SendEmailFn func(ctx *kaos.Context, from string, to []string, cc []string, subject, content string, isHtml bool, attachments ...string) (string, error)
type EnrichFn func(ctx *kaos.Context, user *rbacmodel.User) (codekit.M, codekit.M, error)

type AuthOptions struct {
	SendEmail      SendEmailFn
	Enrich         EnrichFn
	SiamMgr        *siam.Manager
	SignMethod     jwt.SigningMethod
	SignSecret     string
	SecondLifetime int
}

type authEngine struct {
	opts *AuthOptions
	rbac RbacService
}

func NewAuthEngine(rbac RbacService, opts *AuthOptions) (*authEngine, error) {
	if opts == nil {
		return nil, errors.New("options can not be nil")
	}

	if opts.SiamMgr == nil {
		return nil, errors.New("SiamMgr can not be nil")
	}

	ae := new(authEngine)
	ae.opts = opts
	return ae, nil
}

type HttpAuthRequest struct {
	CheckName      string
	SecondLifeTime int
}

type AppAuthRequest struct {
	UserID         string
	CheckName      string
	Password       string
	SecondLifeTime int
}

type AuthResponse struct {
	Token      string
	ExpireTime time.Time
	Data       codekit.M
}

func (obj *authEngine) HttpAuth(ctx *kaos.Context, payload *HttpAuthRequest) (*AuthResponse, error) {
	r, ok := ctx.Data().Get("http_request", nil).(*http.Request)
	if !ok {
		return nil, errors.New("this function should only be called by http request only")
	}
	user, password, ok := r.BasicAuth()
	if !ok {
		return nil, errors.New("invalid credentials (0)")
	}
	return obj.authenticate(ctx, user, payload.CheckName, password, payload.SecondLifeTime)
}

func (obj *authEngine) AppAuth(ctx *kaos.Context, payload *AppAuthRequest) (*AuthResponse, error) {
	return obj.authenticate(ctx, payload.UserID, payload.CheckName, payload.Password, payload.SecondLifeTime)
}

func (obj *authEngine) Logout(ctx *kaos.Context, payload string) (string, error) {
	var err error
	sessionID := ctx.Data().Get("jwt_session_id", "").(string)

	if sessionID == "" {
		token := payload

		if r, isHttp := ctx.Data().Get("http_request", nil).(*http.Request); isHttp {
			token = strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
		}

		if token == "" {
			return "", errors.New("invalid access token")
		}

		sessionID, _, err = JwtToSession(token, obj.opts.SignSecret)
		if err != nil {
			return "", err
		}
	}

	return obj.opts.SiamMgr.Remove(ctx, codekit.M{}.Set("ID", sessionID))
}

func (obj *authEngine) Validate(ctx *kaos.Context, payload codekit.M) (*siam.Session, error) {
	sessionID := payload.GetString("ID")
	walletAddress := payload.GetString("WalletAddress")
	needLoginWithPassword := payload.GetBool("NeedPassword")
	requiredRole := payload.GetString("Role")

	sess, err := obj.opts.SiamMgr.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if walletAddress != "" && sess.Data.GetString("WalletAddress") != walletAddress {
		return nil, errors.New("invalid session")
	}

	if needLoginWithPassword && !sess.Data.GetBool("LoggedIn") {
		return nil, errors.New("invalid session")
	}

	if requiredRole != "" && !obj.rbac.HasRole(sess.ReferenceID, requiredRole) {
		return nil, errors.New("no access")
	}

	return sess, nil
}

func (obj *authEngine) ChangeData(ctx *kaos.Context, payload *rbacmodel.AuthChangeDataRequest) (*AuthResponse, error) {
	token := ""
	if token == "" {
		token = payload.Token
	}

	if token == "" {
		return nil, errors.New("invalid session")
	}

	scope := payload.Scope
	if scope == "" {
		scope = rbacmodel.AuthScopeIsBoth
	}

	sessionID, jwtData, err := JwtToSession(token, obj.opts.SignSecret)
	if err != nil {
		return nil, ctx.Log().Error2("unable to convert token to session", "change data: jwt to session: %s", err.Error())
	}

	sess, err := obj.opts.SiamMgr.Get(ctx, sessionID)
	if err != nil {
		return nil, ctx.Log().Error2("unable to get session", "change data: get session: %s", err.Error())
	}

	for k, v := range payload.Data {
		if scope == rbacmodel.AuthScopeIsBoth || scope == rbacmodel.AuthScopeIsSession {
			if v == "$unset" {
				sess.Data.Unset(k)
			} else {
				sess.Data.Set(k, v)
			}
		}

		if scope == rbacmodel.AuthScopeIsBoth || scope == rbacmodel.AuthScopeIsJWT {
			if v == "$unset" {
				jwtData.Unset(k)
			} else {
				jwtData.Set(k, v)
			}
		}
	}

	newToken, err := obj.sessionToJWT(obj.opts.SiamMgr, ctx, sess.ReferenceID, obj.opts.SecondLifetime, sess.Data, jwtData)
	if err != nil {
		return nil, ctx.Log().Error2("fail change session data", "change data: %s", err.Error())
	}

	return &AuthResponse{
		Token:      newToken,
		ExpireTime: time.Now().Add(time.Duration(sess.Duration) * time.Second),
		Data:       jwtData,
	}, nil
}
