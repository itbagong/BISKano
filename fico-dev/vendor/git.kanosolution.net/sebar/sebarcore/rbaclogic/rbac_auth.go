package rbaclogic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/golang-jwt/jwt"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
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
	SyncPeriod     time.Duration
	Logger         *logger.LogEngine
}

type authEngine struct {
	opts *AuthOptions
	rbac RbacService

	ctx      context.Context
	cancelFn context.CancelFunc

	lock *sync.RWMutex
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
	ae.lock = new(sync.RWMutex)
	ae.ctx, ae.cancelFn = context.WithCancel(context.Background())

	if opts.Logger == nil {
		opts.Logger = logger.NewLogEngine(true, false, "", "", "")
	}

	if int(opts.SyncPeriod) > 0 {
		go ae.Sync(opts.SyncPeriod)
	}

	return ae, nil
}

func (a *authEngine) Close() {
	a.cancelFn()
}

func (a *authEngine) Sync(every time.Duration) {
	a.clearSession()
	for {
		select {
		case <-a.ctx.Done():
			return

		case <-time.After(every):
			a.clearSession()
		}
	}
}

func (a *authEngine) clearSession() {
	keys := a.opts.SiamMgr.Keys()
	for _, key := range keys {
		sess, err := a.opts.SiamMgr.Get(nil, key)
		if err != nil {
			continue
		}
		if sess.LastUpdate.Add(time.Duration(sess.Duration * int(time.Second))).Before(time.Now()) {
			_, err := a.opts.SiamMgr.Remove(nil, codekit.M{}.Set("ID", key))
			if err != nil {
				a.opts.Logger.Errorf("remove session %s, %s error: %s", sess.SessionID, sess.ReferenceID, err.Error())
			} else {
				a.opts.Logger.Infof("remove session %s, %s success", sess.SessionID, sess.ReferenceID)
			}
		}
	}
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
	res, err := obj.authenticate(ctx, user, payload.CheckName, password, obj.opts.SecondLifetime)
	if err == nil {
		obj.opts.Logger.Infof("add session %s, expired on %v", res.Data.GetString("DisplayName"), res.ExpireTime)
	}
	return res, err
}

func (obj *authEngine) AppAuth(ctx *kaos.Context, payload *AppAuthRequest) (*AuthResponse, error) {
	res, err := obj.authenticate(ctx, payload.UserID, payload.CheckName, payload.Password, obj.opts.SecondLifetime)
	if err == nil {
		obj.opts.Logger.Infof("add session %s, expired on %v", res.Data.GetString("DisplayName"), res.ExpireTime)
	}
	return res, err
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
			if _, err := obj.opts.SiamMgr.Remove(ctx, codekit.M{}.Set("ID", sessionID)); err != nil {
				obj.opts.Logger.Errorf("remove session %s because of logout fail. %s", sessionID, err.Error())
			} else {
				obj.opts.Logger.Infof("remove session %s because of logout", sessionID)
			}
			return "", nil
		}
	}

	if res, err := obj.opts.SiamMgr.Remove(ctx, codekit.M{}.Set("ID", sessionID)); err != nil {
		obj.opts.Logger.Errorf("remove session %s because of logout fail. %s", sessionID, err.Error())
	} else {
		obj.opts.Logger.Infof("remove session %s because of logout", sessionID)
		return res, nil
	}

	return "", errors.New("unkown error")
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
	token := payload.Token
	if token == "" {
		token = ctx.Data().Get("jwt_token", "").(string)
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

	ctxUserID := sess.ReferenceID
	if payload.ImpersonateAs != "" {
		ctxUserID = payload.ImpersonateAs
	}
	newToken, err := obj.sessionToJWT(obj.opts.SiamMgr, ctx, ctxUserID, obj.opts.SecondLifetime, sess.Data, jwtData)
	if err != nil {
		return nil, ctx.Log().Error2("fail change session data", "change data: %s", err.Error())
	}

	return &AuthResponse{
		Token:      newToken,
		ExpireTime: time.Now().Add(time.Duration(sess.Duration) * time.Second),
		Data:       jwtData,
	}, nil
}

func (obj *authEngine) GetSessions(ctx *kaos.Context, payload string) ([]string, error) {
	res := obj.opts.SiamMgr.Keys()
	return res, nil
}

func (obj *authEngine) GetSession(ctx *kaos.Context, payload string) (*siam.Session, error) {
	sess, err := obj.opts.SiamMgr.Get(ctx, payload)
	return sess, err
}

func (obj *authEngine) Deimpersonate(ctx *kaos.Context, userID string) (*AuthResponse, error) {
	sessData := ctx.Data().Get("jwt_session_data", codekit.M{}).(codekit.M)
	originalUserID := sessData.GetString("OriginalUserID")
	if originalUserID == "" {
		return nil, errors.New("unauthorized")
	}

	db, _ := ctx.GetHub("iam", "")
	if db == nil {
		return nil, errors.New("missing: db")
	}

	originalUser, _ := datahub.GetByID(db, new(rbacmodel.User), originalUserID)
	if originalUser == nil {
		return nil, errors.New("invalid: user")
	}

	tenantID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("TenantID")
	impersonateRBAC, err := GetUserMatrix(db, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user matrix: %s: %s", userID, err.Error())
	}

	res, err := obj.ChangeData(ctx, &rbacmodel.AuthChangeDataRequest{
		Scope: rbacmodel.AuthScopeIsBoth,
		Data: codekit.M{}.
			Set("OriginalUserID", "$unset").
			Set("DisplayName", originalUser.DisplayName).
			Set("Dimension", originalUser.Dimension).
			Set("Email", originalUser.Email).
			Set("RBAC", impersonateRBAC),
		ImpersonateAs: originalUserID,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (obj *authEngine) Impersonate(ctx *kaos.Context, userID string) (*AuthResponse, error) {
	_, allowToImpersonate, _ := checkAccesFromJWT(ctx, string(JwtSourceServer), &GetAccessRequest{
		AccessType: GetAccessByFeature,
		AccessID:   "Impersonate",
		Level:      1,
	})
	if !allowToImpersonate {
		return nil, errors.New("unauthorized")
	}

	db, _ := ctx.GetHub("iam", "")
	if db == nil {
		return nil, errors.New("missing: db")
	}
	impersonateUser, _ := datahub.GetByID(db, new(rbacmodel.User), userID)
	if impersonateUser == nil {
		return nil, errors.New("invalid: user")
	}

	tenantID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("TenantID")
	impersonateRBAC, err := GetUserMatrix(db, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get user matrix: %s: %s", userID, err.Error())
	}

	originalUserID := ctx.Data().Get("jwt_reference_id", "").(string)

	res, err := obj.ChangeData(ctx, &rbacmodel.AuthChangeDataRequest{
		Scope: rbacmodel.AuthScopeIsBoth,
		Data: codekit.M{}.Set("OriginalUserID", originalUserID).
			Set("DisplayName", impersonateUser.DisplayName).
			Set("Dimension", impersonateUser.Dimension).
			Set("Email", impersonateUser.Email).
			Set("RBAC", impersonateRBAC),
		ImpersonateAs: userID,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

type JwtSource string

const (
	JwtSourceClient JwtSource = "jwt"
	JwtSourceServer JwtSource = "session"
)

func getRbacFromJWT(ctx *kaos.Context, source string) (*rbacmodel.UserMatrix, error) {
	mSource := codekit.M{}
	switch source {
	case string(JwtSourceClient):
		mSource = ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)

	case string(JwtSourceServer):
		mSource = ctx.Data().Get("jwt_session_data", codekit.M{}).(codekit.M)
	}

	_, ok := mSource["RBAC"]
	if !ok {
		return new(rbacmodel.UserMatrix), nil
	}

	rbac := new(rbacmodel.UserMatrix)
	err := serde.Serde(mSource["RBAC"], rbac)
	if err != nil {
		return nil, fmt.Errorf("missing: jwt source: %s", err.Error())
	}

	return rbac, nil
}

func checkAccesFromJWT(ctx *kaos.Context, source string, req *GetAccessRequest) (int, bool, error) {
	rbac, err := getRbacFromJWT(ctx, source)
	if err != nil {
		return 0, false, err
	}

	level, ok := CheckAccess(rbac, req)
	return level, ok, nil
}
