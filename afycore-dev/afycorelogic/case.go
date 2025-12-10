package afycorelogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/afycore/afycoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type CaseLogic struct {
}

type CreateCaseRequest struct {
	Case   afycoremodel.MedicalCase
	Action afycoremodel.CaseAction
}

func (obj *CaseLogic) Create(ctx *kaos.Context, payload *CreateCaseRequest) (*afycoremodel.MedicalCase, error) {
	var err error

	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missing: db")
	}
	h, _ = h.BeginTx()
	defer func() {
		if r := recover(); r != nil {
			h.Rollback()
			return
		}

		if err != nil {
			h.Rollback()
			return
		}

		h.Commit()
	}()

	_, err = datahub.GetByID(h, new(afycoremodel.Patient), payload.Case.PatientID)
	if err != nil {
		return nil, fmt.Errorf("missing: patient: %s", payload.Case.PatientID)
	}

	act, err := datahub.GetByID(h, new(afycoremodel.MedicalAction), payload.Action.ActionID)
	if err != nil {
		return nil, fmt.Errorf("missing: action: %s: %s", payload.Action.ActionID, err.Error())
	}

	_, err = datahub.GetByID(h, new(afycoremodel.LocationPoli), payload.Action.LocationPoliID)
	if err != nil {
		return nil, fmt.Errorf("missing: location poli: %s: %s", payload.Action.LocationPoliID, err.Error())
	}

	_, err = datahub.GetByID(h, new(afycoremodel.MedicalStaff), payload.Action.MainDoctorID)
	if err != nil {
		return nil, fmt.Errorf("missing: medical staff (doctor): %s: %s", payload.Action.MainDoctorID, err.Error())
	}

	resp := &payload.Case
	resp.ID = ""
	resp.TrxDate = codekit.DateOnly(time.Now())
	if err = h.Insert(resp); err != nil {
		return nil, fmt.Errorf("create case: %s", err.Error())
	}

	caseAct := &payload.Action
	caseAct.ID = ""
	caseAct.ActionDate = codekit.DateOnly(time.Now())
	caseAct.QueNo, _ = obj.GetLastQueNo(h, caseAct)
	caseAct.QueNo++
	if caseAct.ActionStatus == "" {
		caseAct.ActionStatus = afycoremodel.ActionScheduled
	}
	if caseAct.BillingStatus == "" {
		caseAct.BillingStatus = lo.Ternary(act.Billable, afycoremodel.BillingDraft, afycoremodel.BillingNone)
	}
	caseAct.BillingStatus = afycoremodel.BillingDraft
	if err = h.Insert(caseAct); err != nil {
		return nil, fmt.Errorf("create case action: %s", err.Error())
	}

	return resp, nil
}

type CaseActionRequest struct {
	CaseID   string
	ActionID string
}

func (obj *CaseLogic) AddAction(ctx *kaos.Context, payload *CaseActionRequest) (*afycoremodel.CaseAction, error) {
	db, _ := ctx.DefaultHub()
	if db == nil {
		return nil, errors.New("missing:db")
	}

	_, err := datahub.GetByID(db, new(afycoremodel.MedicalCase), payload.CaseID)
	if err != nil {
		return nil, fmt.Errorf("missing: case: %s: %s", payload.CaseID, err.Error())
	}
	act, err := datahub.GetByID(db, new(afycoremodel.MedicalAction), payload.ActionID)
	if err != nil {
		return nil, fmt.Errorf("missing: case action: %s: %s", payload.ActionID, err.Error())
	}

	caseAct := new(afycoremodel.CaseAction)
	caseAct.ID = ""
	caseAct.ActionDate = codekit.DateOnly(time.Now())
	caseAct.QueNo, _ = obj.GetLastQueNo(db, caseAct)
	caseAct.QueNo++
	if caseAct.ActionStatus == "" {
		caseAct.ActionStatus = afycoremodel.ActionScheduled
	}
	if caseAct.BillingStatus == "" {
		caseAct.BillingStatus = lo.Ternary(act.Billable, afycoremodel.BillingDraft, afycoremodel.BillingNone)
	}
	caseAct.BillingStatus = afycoremodel.BillingDraft
	if err = db.Insert(caseAct); err != nil {
		return nil, fmt.Errorf("create case action: %s", err.Error())
	}
	return caseAct, nil
}

type CreateBillingFromCaseRequest struct {
	CaseID        string
	CaseActionIDs []string
}

func (obj *CaseLogic) CreateBilling(ctx *kaos.Context, payload *CreateBillingFromCaseRequest) (*afycoremodel.Billing, error) {
	h, _ := ctx.DefaultHub()
	if h == nil {
		return nil, errors.New("missing: db")
	}

	resp := &afycoremodel.Billing{}
	return resp, nil
}

func (obj *CaseLogic) GetLastQueNo(db *datahub.Hub, ca *afycoremodel.CaseAction) (int, error) {
	wheres := dbflex.Eqs("LocationPoliID", ca.LocationPoliID, "ShiftID", ca.ShiftID, "ActionDate", ca.ActionDate)
	parm := dbflex.NewQueryParam().SetWhere(wheres).SetSort("-QueNo").SetTake(1)
	cas, err := datahub.GetByParm(db, new(afycoremodel.CaseAction), parm)
	if err != nil {
		return 0, fmt.Errorf("process: get case action last que no: %s", err.Error())
	}
	return cas.QueNo, nil
}
