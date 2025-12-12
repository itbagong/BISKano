package rbaclogic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/sebarcode/codekit"
)

type NewTokenFn func(ctx *kaos.Context, req *CreateTokenRequest) (string, error)

type tokenEngine struct {
	newTokenFn    NewTokenFn
	defaultLength int
}

func NewTokenEngine(fn NewTokenFn, defaultLength int) *tokenEngine {
	if defaultLength == 0 {
		defaultLength = 8
	}
	te := new(tokenEngine)
	te.defaultLength = defaultLength
	te.newTokenFn = fn
	return te
}

type CreateTokenRequest struct {
	App            string
	Kind           string
	ExpiryInSecond int
	UserID         string
}

func (obj *tokenEngine) Create(ctx *kaos.Context, payload *CreateTokenRequest) (string, error) {
	h := GetRbacDb(ctx)
	if h == nil {
		return "", errors.New("missing: db")
	}

	userID := ""
	if payload.UserID == "" {
		userID = ctx.Data().Get("jwt_reference_id", "").(string)
		if userID != "" {
			return "", errors.New("invalid user")
		}
	} else {
		userID = payload.UserID
	}

	newToken := ""
	if obj.newTokenFn == nil {
		newToken = codekit.GenerateRandomString("abcdefghijklmnopqrstuvwxyz0123456789", obj.defaultLength)
	} else {
		var err error
		newToken, err = obj.newTokenFn(ctx, payload)
		if err != nil {
			return "", err
		}
	}

	tokenRecord := &rbacmodel.Token{
		App:     payload.App,
		Kind:    payload.Kind,
		Token:   newToken,
		Status:  "Reserved",
		UserID:  userID,
		Created: time.Now(),
		Expiry:  time.Now().Add(time.Duration(payload.ExpiryInSecond) * time.Second),
	}
	if e := h.Insert(tokenRecord); e != nil {
		return "", e
	}

	return tokenRecord.Token, nil
}

type ClaimRequest struct {
	App   string
	Kind  string
	Token string
}

func (obj *tokenEngine) Claim(ctx *kaos.Context, payload *ClaimRequest) (string, error) {
	h := GetRbacDb(ctx)
	if h == nil {
		return "", errors.New("missing: db")
	}

	userID := ctx.Data().Get("jwt_reference_id", "").(string)
	if userID != "" {
		return "", errors.New("invalid user")
	}

	where := dbflex.Eqs("App", payload.App, "Kind", payload.Kind, "Token", payload.Token)
	token := new(rbacmodel.Token)
	if e := h.GetByFilter(token, where); e != nil {
		return "", errors.New("invalid token")
	}
	if token.Status != "Reserved" || token.UserID != userID {
		return "", errors.New("invalid token")
	}
	if token.Expiry.After(time.Now()) {
		if token.Status != "Expired" {
			token.Status = "Expired"
			go h.Update(token, "Status")
		}
		return "", errors.New("expiry token, you need to restart the operation")
	}
	token.Status = "Claimed"
	token.ClaimDate = time.Now()
	if e := h.Update(token, "Status", "ClaimDate"); e != nil {
		return "", e
	}

	return token.ID, nil
}
