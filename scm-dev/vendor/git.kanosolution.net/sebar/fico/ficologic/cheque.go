package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/suim"
)

func RegisterCGBook(s *kaos.Service) {
	modDb := sebar.NewDBModFromContext()
	modUI := suim.New()

	s.RegisterModel(new(CGLogic), "cg")
	s.RegisterModel(new(ficomodel.ChequeGiroBook), "cgbook").
		SetMod(modDb, modUI).DisableRoute("insert", "update", "save")
	s.RegisterModel(new(ficomodel.ChequeGiro), "cheque").
		SetMod(modDb, modUI)
}

type CGLogic struct {
}

type CGResponse struct {
	ChequeGiroID  []string
	CashJournalID string
}

func (o *CGLogic) CreateUpdateBook(ctx *kaos.Context, cgbook *ficomodel.ChequeGiroBook) (*ficomodel.ChequeGiroBook, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if coID == "" {
		return nil, fmt.Errorf("missing: company id")
	}

	now := time.Now()
	cgbook.CompanyID = coID
	cgbook.ReceiveDate = &now

	pattern, startIndex, err := ExtractPattern(cgbook.From)
	if err != nil {
		return nil, fmt.Errorf("get cheque/giro pattern: %s", err.Error())
	}

	if cgbook.Qty == 0 {
		toPattern, toIndex, err := ExtractPattern(cgbook.To)
		if err != nil {
			return nil, fmt.Errorf("get cheque/giro pattern: %s", err.Error())
		}

		if toPattern != pattern {
			return nil, fmt.Errorf("to pattern is not valid: %s", cgbook.To)
		}

		cgbook.Qty = toIndex - startIndex + 1
	} else if cgbook.To == "" {
		cgbook.To = fmt.Sprintf(pattern, startIndex+cgbook.Qty-1)
	}

	err = db.Save(cgbook)
	if err != nil {
		return nil, err
	}

	for index := 0; index < cgbook.Qty; index++ {
		cgID := fmt.Sprintf(pattern, startIndex+index)
		_, cgFound := datahub.GetByID(db, new(ficomodel.ChequeGiro), cgID)
		if cgFound == nil {
			continue
		}

		cg := &ficomodel.ChequeGiro{
			ID:          cgID,
			CashBookID:  cgbook.CashBookID,
			CheckBookID: cgbook.ID,
			Kind:        cgbook.Kind,
			Status:      ficomodel.CGOpen,
			CompanyID:   cgbook.CompanyID,
		}
		if err := db.Save(cg); err != nil {
			return nil, fmt.Errorf("save cheque giro: %s: %s", cg.ID, err.Error())
		}
	}

	return cgbook, nil
}

func (o *CGLogic) ValidateDeleteBook(ctx *kaos.Context, cgbook *ficomodel.ChequeGiroBook) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, fmt.Errorf("missing: db")
	}

	// check status
	chequeGiro := []ficomodel.ChequeGiro{}
	e := db.Gets(new(ficomodel.ChequeGiro), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("CheckBookID", cgbook.ID),
			dbflex.Eq("Status", ficomodel.CGCleared),
		)), &chequeGiro)
	if e != nil {
		return false, errors.New("Failed populate data checque/giro: " + e.Error())
	}

	if len(chequeGiro) > 0 {
		return false, errors.New("Cannot delete this data. Because some cheque/giro status is cleared.")
	}

	return true, nil
}

func (o *CGLogic) DeleteChequeGiro(ctx *kaos.Context, cgbook *ficomodel.ChequeGiroBook) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, fmt.Errorf("missing: db")
	}

	// delete by check book id
	db.DeleteQuery(new(ficomodel.ChequeGiro), dbflex.Eq("CheckBookID", cgbook.ID))

	return true, nil
}

func (o *CGLogic) UpdateCashJournalID(ctx *kaos.Context, cg *CGResponse) (bool, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return false, fmt.Errorf("missing: db")
	}

	if cg == nil {
		return false, fmt.Errorf("Payload is required.")
	} else {
		if cg.CashJournalID == "" {
			return false, fmt.Errorf("Cash journal ID is required.")
		}
	}

	// gets all cheque giro by id
	chequeGiros := []ficomodel.ChequeGiro{}
	mapCheque := map[string]ficomodel.ChequeGiro{}
	e := db.Gets(new(ficomodel.ChequeGiro), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", cg.ChequeGiroID),
		)), &chequeGiros)
	if e != nil {
		return false, errors.New("Failed populate data cheque/giro: " + e.Error())
	}

	if len(chequeGiros) > 0 {
		for _, val := range chequeGiros {
			mapCheque[val.ID] = val
		}
	}

	if len(cg.ChequeGiroID) > 0 {
		for _, val := range cg.ChequeGiroID {
			if v, ok := mapCheque[val]; ok {
				res := v
				res.CashJournalID = cg.CashJournalID

				db.Save(&res)
			}
		}
	}

	return true, nil
}

func (obj *CGLogic) Reserve(ctx *kaos.Context, payload *ficomodel.ChequeGiro) (*ficomodel.ChequeGiro, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	res, err := datahub.GetByID(db, new(ficomodel.ChequeGiro), payload.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid: cheque: %s", payload.ID)
	}

	if res.Status != ficomodel.CGOpen {
		return nil, fmt.Errorf("invalid: cheque: %s: %s", payload.ID, "has been used")
	}

	cj, err := datahub.GetByID(db, new(ficomodel.CashJournal), payload.CashJournalID)
	if err != nil {
		return nil, fmt.Errorf("invalid: cash journal: %s", err.Error())
	}
	if cj.Status != ficomodel.JournalStatusDraft {
		return nil, fmt.Errorf("invalid: cash journal: to reserve status must be draft")
	}
	cj.Lines, err = cj.Lines.
		Update(payload.LineNo, func(line *ficomodel.JournalLine) {
			line.ChequeGiroID = res.ID
		})
	if err != nil {
		return nil, fmt.Errorf("invalid: line no: %d", payload.LineNo)
	}
	db.Save(cj, "Lines")

	res.Status = ficomodel.CGReserved
	res.Amount = payload.Amount
	res.IssueDate = payload.IssueDate
	res.ClearDate = payload.ClearDate
	res.BfcAccount = payload.BfcAccount
	res.BfcBank = payload.BfcBank
	res.BfcName = payload.BfcName
	res.BfcSwift = payload.BfcSwift
	res.Memo = payload.Memo
	res.CashJournalID = payload.CashJournalID
	res.LineNo = payload.LineNo
	db.Save(res)

	return res, nil
}

func (obj *CGLogic) Unreserve(ctx *kaos.Context, payload *ficomodel.ChequeGiro) (*ficomodel.ChequeGiro, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	res, err := datahub.GetByID(db, new(ficomodel.ChequeGiro), payload.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid: cheque: %s", payload.ID)
	}

	if res.Status != ficomodel.CGReserved {
		return nil, fmt.Errorf("invalid: cheque: %s: %s", payload.ID, "status != reserved")
	}

	cj, err := datahub.GetByID(db, new(ficomodel.CashJournal), res.CashJournalID)
	if err != nil {
		return nil, fmt.Errorf("invalid: cash journal: %s", err.Error())
	}
	if cj.Status != ficomodel.JournalStatusDraft {
		return nil, fmt.Errorf("invalid: cash journal: to unreserve status must be draft")
	}
	cj.Lines, err = cj.Lines.
		Update(payload.LineNo, func(line *ficomodel.JournalLine) {
			line.ChequeGiroID = ""
		})
	if err != nil {
		return nil, fmt.Errorf("invalid: line no: %d", payload.LineNo)
	}
	db.Save(cj, "Lines")

	res.Status = ficomodel.CGOpen
	res.Amount = 0
	res.IssueDate = nil
	res.ClearDate = nil
	res.BfcAccount = ""
	res.BfcBank = ""
	res.BfcName = ""
	res.BfcSwift = ""
	res.Memo = ""
	res.CashJournalID = ""
	res.LineNo = 0
	db.Save(res)

	return res, nil
}

func (obj *CGLogic) ReleaseFund(ctx *kaos.Context, payload *ficomodel.ChequeGiro) (*ficomodel.ChequeGiro, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	res, err := datahub.GetByID(db, new(ficomodel.ChequeGiro), payload.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid: cheque: %s", payload.ID)
	}

	if res.Status != ficomodel.CGReserved {
		return nil, fmt.Errorf("invalid: cheque: %s: %s", payload.ID, "status != reserved")
	}

	if payload.ReleaseDate == nil {
		return nil, fmt.Errorf("missing: release date")
	}

	if res.ClearDate.Before(*payload.ReleaseDate) {
		return nil, fmt.Errorf("release date is before clear date")
	}

	res.Status = ficomodel.CGCleared
	res.ReleaseDate = payload.ReleaseDate
	db.Save(res)

	return res, nil
}
