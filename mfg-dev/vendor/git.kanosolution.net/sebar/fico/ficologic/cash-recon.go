package ficologic

import (
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
)

type CashReconLogic struct {
}

func (obj *CashReconLogic) Save(ctx *kaos.Context, payload *ficomodel.CashRecon) (*ficomodel.CashRecon, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	payload.Status = ficomodel.CashReconStatusDraft

	err := h.Save(payload)
	if err != nil {
		return nil, fmt.Errorf("err when save: %s", err.Error())
	}

	return payload, nil
}

type GetPreviousReconRequest struct {
	CompanyID  string
	CashBankID string
}

func (obj *CashReconLogic) GetPreviousRecon(ctx *kaos.Context, payload *GetPreviousReconRequest) (*ficomodel.CashRecon, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	prevRecon, err := datahub.GetByParm(h, new(ficomodel.CashRecon), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eqs("CompanyID", payload.CompanyID, "CashBankID", payload.CashBankID, "Status", ficomodel.CashReconStatusCompleted),
		),
	).SetSort("-ReconDate").SetTake(1))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("err when get cash recon: %s", err.Error())
	}

	return prevRecon, nil
}

func (obj *CashReconLogic) GetCurrentRecon(ctx *kaos.Context, payload *GetPreviousReconRequest) (*ficomodel.CashRecon, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := now.AddDate(0, 0, 1)
	prevRecon, err := datahub.GetByParm(h, new(ficomodel.CashRecon), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eqs("CompanyID", payload.CompanyID, "CashBankID", payload.CashBankID, "Status", ficomodel.CashReconStatusDraft),
			dbflex.Gte("ReconDate", start), dbflex.Lt("ReconDate", end),
		),
	))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("err when get cash recon: %s", err.Error())
	}

	return prevRecon, nil
}

func (obj *CashReconLogic) StartRecon(ctx *kaos.Context, payload *ficomodel.CashRecon) (*ficomodel.CashRecon, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	res, err := datahub.GetByID(db, new(ficomodel.CashRecon), payload.ID)
	if err != nil {
		return nil, err
	}

	// get previous recon
	prevRecon, err := datahub.GetByFilter(db, new(ficomodel.CashRecon), dbflex.And(
		dbflex.Eqs("CompanyID", res.CompanyID, "CashBankID", res.CashBankID, "Status", ficomodel.CashReconStatusCompleted),
		dbflex.Lt("ReconDate", res.ReconDate),
	))

	// prev recon found
	if err == nil {
		if prevRecon.Diff != 0 {
			return nil, fmt.Errorf("invalid: cash recon: there is other open recon %s, %s", prevRecon.ID, prevRecon.ReconDate.Format("2016-Jan-02"))
		}
		// get prev open trxs, if any shld return fail
		_, err := datahub.GetByFilter(db, new(ficomodel.CashTransaction), dbflex.And(
			dbflex.Eqs("CompanyID", res.CompanyID, "CashBankD._id", res.CashBankID),
			dbflex.Lte("TrxDate", prevRecon.ReconDate),
			dbflex.Eq("CashReconID", ""),
		))
		if err == nil {
			return nil, fmt.Errorf("invalid: cash recon: there is open transaction for recon %s, %s", prevRecon.ID, prevRecon.ReconDate)
		}
		res.PreviousReconDate = &prevRecon.ReconDate
		res.PreviousBalance = prevRecon.ReconBalance
	}

	trxs, _ := obj.GetTransactions(ctx, res)
	sumRecon := float64(0)
	for _, trx := range trxs {
		if trx.CashReconID == res.ID {
			sumRecon += trx.Amount
		} else if trx.CashReconID != "" && trx.CashReconID != res.ID {
			trx.CashReconID = ""
			db.Save(trx)
		}
	}
	res.Diff = res.ReconBalance - res.PreviousBalance - sumRecon
	db.Save(res, "Diff", "PreviosReconDate", "PreviousBalance")

	return res, nil
}

func (obj *CashReconLogic) GetTransactions(ctx *kaos.Context, payload *ficomodel.CashRecon) ([]*ficomodel.CashTransaction, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	if payload.CashBankID == "" {
		return nil, fmt.Errorf("missing: cashbankid")
	}

	filters := []*dbflex.Filter{
		dbflex.Eqs("CompanyID", payload.CompanyID, "CashBank._id", payload.CashBankID, "CashReconID", ""),
		dbflex.Lte("TrxDate", payload.ReconDate),
	}
	if payload.PreviousReconDate != nil {
		filters = append(filters, dbflex.Gt("TrxDate", payload.PreviousReconDate))
	}
	filter := dbflex.And(filters...)

	res, err := datahub.FindByFilter(db, new(ficomodel.CashTransaction), filter)
	return res, err
}

type CashReconcileRequest struct {
	CashReconID        string
	CashTransactionIDs []string
}

func (obj *CashReconLogic) Reconcile(ctx *kaos.Context, payload *CashReconcileRequest) (*ficomodel.CashRecon, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	res, err := datahub.GetByID(db, new(ficomodel.CashRecon), payload.CashReconID)
	if err != nil {
		return nil, fmt.Errorf("missing: cash recon: %s", payload.CashReconID)
	}

	db.UpdateField(&ficomodel.CashTransaction{CashReconID: ""},
		dbflex.Eqs("CompanyID", res.CompanyID, "CashReconID", res.ID), "CashReconID")

	total := float64(0)
	for _, id := range payload.CashTransactionIDs {
		trx, err := datahub.GetByID(db, new(ficomodel.CashTransaction), id)
		if err != nil {
			return nil, fmt.Errorf("invalid: cash transaction: %s: %s", id, err.Error())
		}
		if trx.CashBank.ID != res.CashBankID {
			return nil, fmt.Errorf("invalid: cash transaction: %s: wrong cash bank", id)
		}
		if trx.TrxDate.After(res.ReconDate) || (res.PreviousReconDate != nil && trx.TrxDate.Before(*res.PreviousReconDate)) {
			return nil, fmt.Errorf("invalid: cash transaction: %s: wrong date", id)
		}
		trx.CashReconID = res.ID
		total += trx.Amount
		go db.Save(trx)
	}

	res.Diff = res.ReconBalance - res.PreviousBalance - total
	res.Status = ficomodel.CashReconStatusCompleted
	db.Save(res, "Diff", "Status")

	return res, nil
}
