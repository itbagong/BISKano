package bagonglogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type CustomerHandler struct {
}

type CustomerPostRequest struct {
	tenantcoremodel.Customer
	Detail   bagongmodel.CustomerDetail
	Config   bagongmodel.CustomerConfiguration
	Contacts []tenantcoremodel.Contact
}

func (obj *CustomerHandler) Get(ctx *kaos.Context, payload []interface{}) (*CustomerPostRequest, error) {
	if len(payload) == 0 {
		return nil, errors.New("invalid request")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	c := new(tenantcoremodel.Customer)
	if e := h.GetByID(c, payload[0]); e != nil {
		return nil, fmt.Errorf("customer not found: %s", payload)
	}

	cDetail := new(bagongmodel.CustomerDetail)
	h.GetByFilter(cDetail, dbflex.Eq("CustomerID", c.ID))

	cConfig := new(bagongmodel.CustomerConfiguration)
	h.GetByFilter(cConfig, dbflex.Eq("CustomerID", c.ID))

	cContact := []tenantcoremodel.Contact{}
	if e := h.Gets(new(tenantcoremodel.Contact), dbflex.NewQueryParam().SetWhere(dbflex.Eq("CustomerID", c.ID)), &cContact); e != nil {
		return nil, e
	}

	res := CustomerPostRequest{
		Customer: *c,
		Detail:   *cDetail,
		Config:   *cConfig,
		Contacts: cContact,
	}

	return &res, nil
}

func (obj *CustomerHandler) Save(ctx *kaos.Context, payload *CustomerPostRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Customer.ID == "" {
		tenantcorelogic.MWPreAssignSequenceNo("Customer", false, "_id")(ctx, &payload.Customer)

		if e := h.GetByID(new(tenantcoremodel.Customer), payload.Customer.ID); e == nil {
			ctx.Log().Errorf("error duplicate key: %s", e.Error())
			return nil, errors.New("error duplicate key: " + e.Error())
		}
	}

	if e := h.GetByID(new(tenantcoremodel.Customer), payload.ID); e != nil {
		if e := h.Insert(&payload.Customer); e != nil {
			return nil, errors.New("error insert Customer: " + e.Error())
		}
	} else {
		if e := h.Save(&payload.Customer); e != nil {
			return nil, errors.New("error save Customer: " + e.Error())
		}
	}

	if e := h.GetByID(new(bagongmodel.CustomerDetail), payload.Detail.ID); e != nil {
		payload.Detail.CustomerID = payload.Customer.ID
		if e := h.Insert(&payload.Detail); e != nil {
			return nil, errors.New("error insert CustomerDetail: " + e.Error())
		}
	} else {
		if e := h.Save(&payload.Detail); e != nil {
			return nil, errors.New("error save CustomerDetail: " + e.Error())
		}
	}

	if e := h.GetByID(new(bagongmodel.CustomerConfiguration), payload.Config.ID); e != nil {
		payload.Config.CustomerID = payload.Customer.ID
		if e := h.Insert(&payload.Config); e != nil {
			return nil, errors.New("error insert CustomerConfiguration: " + e.Error())
		}
	} else {
		if e := h.Save(&payload.Config); e != nil {
			return nil, errors.New("error save CustomerConfiguration: " + e.Error())
		}
	}

	for _, l := range payload.Contacts {
		if e := h.GetByID(new(tenantcoremodel.Contact), l.ID); e != nil {
			l.CustomerID = payload.Customer.ID
			if e := h.Insert(&l); e != nil {
				return nil, errors.New("error insert Customer Contact : " + e.Error())
			}
		} else {
			if e := h.Save(&l); e != nil {
				return nil, errors.New("error save Customer Contact: " + e.Error())
			}
		}
	}

	return payload, nil
}
