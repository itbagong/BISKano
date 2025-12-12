package sebar

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

type CheckAccessRequest struct {
	Param        codekit.M
	PermissionID string
	AccessLevel  int
}

func CheckAccess(ctx *kaos.Context, request *CheckAccessRequest) error {
	/*
		ev, _ := ctx.DefaultEvent()
		if ev == nil {
			return errors.New("missing: eventhub")
		}

		validationResult := ""
		err := ev.Publish(ficologic.Config.AddrAccessValidation,
			parm, &validationResult,
			sebar.CopyContextDataToPublishOptions(ctx, &kaos.PublishOpts{}))
		if err != nil {
			return fmt.Errorf("access validation error: %s: %s", permission, err.Error())
		}
		return nil
	*/
	return nil
}
