package shelogic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

type IBPRTransactionLogic struct {
}

func (obj *IBPRTransactionLogic) Save(ctx *kaos.Context, payload *shemodel.IBPRTransaction) (*shemodel.IBPRTransaction, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	now := time.Now()

	// generate number ibpr transaction
	tenantcorelogic.MWPreAssignSequenceNo("IBPRTransaction", false, "_id")(ctx, &payload)
	if e := h.GetByID(new(shemodel.IBPRTransaction), payload.ID); e != nil {
		payload.ID = "BG/" + payload.Dimension.Get("CC") + "/IBPR/" + payload.ID
		//update created/updated by and created/updated time
		if len(payload.Lines) > 0 {
			for i, _ := range payload.Lines {
				payload.Lines[i].CreatedBy = userID
				payload.Lines[i].CreatedTime = now
				payload.Lines[i].UpdatedBy = userID
				payload.Lines[i].UpdatedTime = now

				//update isupdated to default false
				payload.Lines[i].IsUpdated = false
			}
		}
		if e := h.Insert(payload); e != nil {
			return nil, errors.New("error insert ibpr transaction: " + e.Error())
		}
	} else {
		//update updated by and updated time
		if len(payload.Lines) > 0 {
			for i, _ := range payload.Lines {
				if payload.Lines[i].IsUpdated {
					payload.Lines[i].UpdatedBy = userID
					payload.Lines[i].UpdatedTime = now

					//update isupdated to default false
					payload.Lines[i].IsUpdated = false
				}
			}
		}
		if e := h.Save(payload); e != nil {
			return nil, errors.New("error update ibpr transaction: " + e.Error())
		}
	}

	return payload, nil
}
