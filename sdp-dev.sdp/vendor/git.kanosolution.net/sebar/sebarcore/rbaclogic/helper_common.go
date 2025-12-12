package rbaclogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebarcore"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

func send2FA(ctx *kaos.Context, user *rbacmodel.User, h *datahub.Hub, kind string) error {
	token := &rbacmodel.Token{}
	where := dbflex.Eqs("UserID", user.ID, "Kind", kind, "Status", "Reserved")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))
	if token.ID == "" {
		// create token
		token.Token = codekit.GenerateRandomString("", 16)
		token.Status = "Reserved"
		token.UserID = user.ID
		token.Kind = kind
		token.App = "Xibar"
		token.Expiry = time.Now().Add(2 * time.Hour)
	}
	h.Save(token)

	//-- send email
	_, err := sebarcore.SendEmail(ctx, user.Email, kind, "en-us", codekit.M{
		"Email":          user.Email,
		"ValidationCode": token.Token,
	})
	if err != nil {
		return fmt.Errorf("fail send email: %s", err.Error())
	}

	return nil
}

func set2FA(ctx *kaos.Context, user *rbacmodel.User, h *datahub.Hub, kind string) (*AuthResponse, error) {
	token := &rbacmodel.Token{}
	where := dbflex.Eqs("UserID", user.ID, "Kind", kind, "Status", "Claimed")
	h.GetByParm(token, dbflex.NewQueryParam().SetSort("-Expiry").SetWhere(where))
	if token.ID == "" {
		err := send2FA(ctx, user, h, kind)
		if err != nil {
			return nil, err
		}
	} else {
		token.Status = "Expired"
		h.Save(token)
	}

	if token.Status != "Expired" {
		return &authResp, errors.New("Use2FA is required")
	}

	return nil, nil
}
