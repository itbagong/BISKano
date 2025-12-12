package rbaclogic

import (
	"errors"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/golang-jwt/jwt"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/siam"
)

var (
	authResp = AuthResponse{}
)

func (obj *authEngine) authenticate(ctx *kaos.Context, id, checkName, password string, seconfLifetime int) (*AuthResponse, error) {
	h, _ := ctx.GetHub("iam", "")
	if h == nil {
		return nil, errors.New("db_conn")
	}

	if id == "" {
		return nil, errors.New("mandatory: id")
	}

	if !codekit.HasMember([]string{"LoginID", "Email", "WalletAddress"}, checkName) {
		return nil, errors.New("invalid: checkName")
	}

	user := new(rbacmodel.User)
	userWhere := dbflex.Eq(checkName, id)
	if err := h.GetByFilter(user, userWhere); err != nil {
		if err == io.EOF {
			return nil, errors.New("invalid credential (1)")
		}
		return nil, ctx.Log().Error2("unknown error when reading user profile", "auth get user error: %s, %s. %s", id, checkName, err.Error())
	}
	if !user.Enable {
		return nil, errors.New("user is not allowed to login, please contact system support")
	}

	// Two Factor Authentication (2FA)
	if user.Use2FA {
		kind2FA := "rbac-user-2fa"
		_, err := set2FA(ctx, user, h, kind2FA)
		if err != nil {
			return nil, err
		}
	}

	up := rbacmodel.UserPassword{}
	h.GetByID(&up, user.ID)
	if up.Password != codekit.ShaString(password, "") {
		return nil, errors.New("invalid credential (2)")
	}

	var err error
	sessionData := codekit.M{}
	jwtData := codekit.M{}
	if obj.opts.Enrich != nil {
		if sessionData, jwtData, err = obj.opts.Enrich(ctx, user); err != nil {
			return nil, ctx.Log().Error2("unknown error when enrich user profile", "auth enrich user error: %s, %s. %s", id, checkName, err.Error())
		}
	}

	expirySecond := seconfLifetime
	if expirySecond == 0 {
		expirySecond = obj.opts.SecondLifetime
	}
	token, err := obj.sessionToJWT(obj.opts.SiamMgr, ctx, user.ID, expirySecond, sessionData, jwtData)
	if err != nil {
		return nil, ctx.Log().Error2("unknown error when create jwt", "auth jwt error: %s, %s. %s", id, checkName, err.Error())
	}
	authResp.Token = token
	authResp.Data = jwtData
	authResp.ExpireTime = time.Now().Add(time.Duration(expirySecond * int(time.Second)))
	return &authResp, nil
}

func (obj *authEngine) sessionToJWT(siamMgr *siam.Manager, ctx *kaos.Context, userID string, lifeTime int, sessData, jwtData codekit.M) (string, error) {
	sess, err := siamMgr.FindOrCreate(ctx, codekit.M{}.Set("ID", userID).Set("Second", lifeTime), sessData)
	if err != nil {
		return "", errors.New("auth error: " + err.Error())
	}

	bc := new(siam.AuthJwt)
	bc.Id = sess.SessionID
	bc.ExpiresAt = time.Now().Add(time.Duration(sess.Duration) * time.Second).UnixMilli()
	bc.Data = jwtData

	token := jwt.NewWithClaims(obj.opts.SignMethod, bc)
	tokenString, err := token.SignedString([]byte(obj.opts.SignSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func JwtToSession(jwtCode, secret string) (string, codekit.M, error) {
	bc := siam.AuthJwt{}
	m := codekit.M{}
	tkn, e := jwt.ParseWithClaims(jwtCode, &bc, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if e != nil {
		return "", m, e
	}
	if !tkn.Valid {
		return "", m, errors.New("invalid_access_token")
	}
	m = bc.Data
	return bc.Id, m, nil
}
