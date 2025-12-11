package bagonglogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
)

type TaxEngine struct{}

type FPRequest struct {
	ID     string
	FPOld  string
	FPNew  string
	Status ficomodel.TaxStatus
}

func (engine *TaxEngine) SaveFp(ctx *kaos.Context, req FPRequest) (string, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	taxTrx := new(ficomodel.TaxTransaction)
	e := hub.GetByID(taxTrx, req.ID)
	if e != nil {
		return "", errors.New("error: failed to save tax transaction")
	}

	taxTrx.FPNo = req.FPNew
	taxTrx.Status = req.Status
	// save tax transaction
	if e = hub.Save(taxTrx); e != nil {
		return "", fmt.Errorf("error when save tax transaction: %s", e.Error())
	}

	if req.FPOld != "" {
		taxInvoice := new(bagongmodel.TaxInvoice)
		e = hub.GetByID(taxInvoice, req.FPOld)
		if e != nil {
			return "", errors.New("error: failed to save tax transaction")
		}
		taxInvoice.Status = string(ficomodel.TaxOpen)
		// save tax transaction
		if e = hub.Save(taxInvoice); e != nil {
			return "", fmt.Errorf("error when save tax transaction: %s", e.Error())
		}
	}

	return "Success", nil
}

func (engine *TaxEngine) SaveDownload(ctx *kaos.Context, req FPRequest) ([]ficomodel.TaxTransaction, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	res := []ficomodel.TaxTransaction{}
	if e := h.GetsByFilter(new(ficomodel.TaxTransaction), dbflex.Eq("Status", "Allocated"), &res); e != nil {
		return nil, e
	}

	for _, c := range res {
		c.Status = ficomodel.TaxSubmitted
		// save tax transaction
		if e := h.Save(&c); e != nil {
			return nil, fmt.Errorf("error when save tax transaction: %s", e.Error())
		}
	}

	return res, nil
}
