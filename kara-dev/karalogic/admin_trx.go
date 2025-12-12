package karalogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"github.com/sebarcode/codekit"

	"git.kanosolution.net/kano/kaos"
)

type AdminTrx struct {
	appConfig *sebar.AppConfig
}

func NewAdminTrxLogic(appConfig *sebar.AppConfig, opt *rbaclogic.AuthOptions) *AdminTrx {
	return &AdminTrx{
		appConfig,
	}
}

func (u *AdminTrx) Create(ctx *kaos.Context, payload *karamodel.TrxRequest) (*karamodel.AttendanceTrx, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}
	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("missing: event hub")
	}

	var paramTrx karamodel.AttendanceTrx
	paramTrx.UserID = payload.UserID
	paramTrx.Op = karamodel.OpCode(payload.Op)
	paramTrx.WorkLocationID = payload.WorkLocationID
	paramTrx.TrxDate = codekit.DateOnly(payload.TrxDate)
	if payload.TrxTime != "" {
		newDateTimeStr := fmt.Sprintf("%sT%s%s", payload.TrxDate.Format("2006-01-02"), payload.TrxTime, payload.TrxDate.Format("-0700"))
		newDateTime, err := time.Parse("2006-01-02T15:04-0700", newDateTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid: time: %s", err.Error())
		}
		paramTrx.TrxDate = newDateTime
	}

	res, err := saveTrx(h, ev, &paramTrx, payload.ConfirmForReview)
	if err != nil {
		return nil, err
	}

	return res, nil
}
