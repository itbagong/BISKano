package sdplogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type SalesOpportunityEngine struct {
}

func (o *SalesOpportunityEngine) upsert(h *datahub.Hub, model *sdpmodel.SalesOpportunity) error {
	if e := h.GetByID(new(sdpmodel.SalesOpportunity), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Sales Order : " + e.Error())
		}
	} else {
		if e := h.Save(model); e != nil {
			return errors.New("error update Sales Order : " + e.Error())
		}
	}

	return nil
}

func (o *SalesOpportunityEngine) Insert(ctx *kaos.Context, payload *sdpmodel.SalesOpportunity) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Customer == "" {
		return nil, errors.New("missing: payload")
	}

	// check Event hub
	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, errors.New("Event not null")
	}

	OP := payload

	last := []sdpmodel.SalesOpportunity{}
	err = h.Gets(new(sdpmodel.SalesOpportunity), dbflex.NewQueryParam().SetWhere(dbflex.Gte("Created", time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC))).SetSort("-No").SetTake(1), &last)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %v Last sales quotation", err))
	}

	if len(last) > 0 {
		OP.No = last[0].No + 1
		OP.OpportunityNo = fmt.Sprintf("OP/%04d/BDM-HO/%02d/%d", OP.No, (int(time.Now().Month())), time.Now().Year())
	} else {
		OP.No = 1
		OP.OpportunityNo = fmt.Sprintf("OP/%04d/BDM-HO/%02d/%d", OP.No, (int(time.Now().Month())), time.Now().Year())
	}

	err = o.upsert(h, OP)
	if err != nil {
		return nil, err
	}

	return OP, nil
}

func (o *SalesOpportunityEngine) Update(ctx *kaos.Context, payload *sdpmodel.SalesOpportunity) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Customer == "" {
		return nil, errors.New("missing: payload")
	}

	// check Event hub
	event, err := ctx.DefaultEvent()
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, errors.New("Event not null")
	}

	OP := payload

	err = o.upsert(h, OP)
	if err != nil {
		return nil, err
	}

	return OP, nil
}
