package ficologic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type ApplyLogic struct {
}

type GetApplyDataRequest struct {
	Direction          string
	Draft              bool
	SourceTypeExcept   string
	SourceType         string
	SourceJournalID    string
	ApplyFromType      string
	ApplyFromJournalID string
	ApplyFromLineNo    int
	ApplyFromRecordID  string
	SubledgerType      string
	SubledgerID        string
}

type GetScheduleRespond struct {
	Schedules []*ficomodel.CashSchedule
	Applies   []*ficomodel.CashApply
}

func (o *ApplyLogic) GetSchedules(ctx *kaos.Context, req *GetApplyDataRequest) (*GetScheduleRespond, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}
	if req == nil {
		req = new(GetApplyDataRequest)
	}
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)

	var (
		scheds []*ficomodel.CashSchedule
		err    error
	)

	// get top 25 schedules
	if req.Draft {
		scheds, err = o.getDraftSchedules(db, req)
	} else {
		scheds, err = o.getSchedules(db, coID, req)
	}

	if err != nil {
		return nil, err
	}

	// combine with already applied
	schedsA, applies, err := o.getApplied(db, req)
	if err != nil {
		return nil, err
	}

	schedsAF := lo.Filter(schedsA, func(sa *ficomodel.CashSchedule, _ int) bool {
		_, _, ok := lo.FindIndexOf(scheds, func(s *ficomodel.CashSchedule) bool {
			return s.ID == sa.ID
		})
		return !ok
	})

	if len(schedsAF) > 0 {
		scheds = append(scheds, schedsAF...)
	}

	return &GetScheduleRespond{
		Schedules: scheds,
		Applies:   applies,
	}, nil
}

func (o *ApplyLogic) getSchedules(db *datahub.Hub, coID string, req *GetApplyDataRequest) ([]*ficomodel.CashSchedule, error) {
	filters := []*dbflex.Filter{dbflex.Eq("CompanyID", coID), dbflex.Ne("Outstanding", 0)}

	switch req.Direction {
	case "CashOut":
		filters = append(filters, dbflex.Lt("Amount", 0))

	case "CashIn":
		filters = append(filters, dbflex.Gt("Amount", 0))
	}

	if req.SourceTypeExcept != "" {
		filters = append(filters, dbflex.Ne("SourceType", req.SourceTypeExcept))
	}

	if req.SourceType != "" {
		filters = append(filters, dbflex.Eq("SourceType", req.SourceType))
	}

	if req.SourceJournalID != "" {
		filters = append(filters, dbflex.Eq("SourceJournalID", req.SourceJournalID))
	}

	if req.SubledgerType != "" {
		filters = append(filters, dbflex.Eq("Account.AccountType", req.SubledgerType))
	}

	if req.SubledgerID != "" {
		filters = append(filters, dbflex.Eq("Account.AccountID", req.SubledgerID))
	}

	parm := dbflex.NewQueryParam().
		SetWhere(dbflex.And(filters...))

	scheds, err := datahub.Find(db, new(ficomodel.CashSchedule), parm)

	return scheds, err
}

func (o *ApplyLogic) getDraftSchedules(db *datahub.Hub, req *GetApplyDataRequest) ([]*ficomodel.CashSchedule, error) {
	switch strings.ToLower(req.SourceType) {
	case "cashbank":
		cb, err := datahub.GetByID(db, new(ficomodel.CashJournal), req.SourceJournalID)
		if err != nil {
			return nil, fmt.Errorf("invalid: source journal: %s, %s", req.SourceJournalID, err.Error())
		}
		if cb.CashJournalType == "CASH OUT" {
			req.Direction = "CashOut"
		} else {
			req.Direction = "CashIn"
		}

		db.DeleteByFilter(new(ficomodel.CashSchedule), dbflex.Eqs("SourceType", ficomodel.SubledgerCashBank, "SourceJournalID", cb.ID))

		res := []*ficomodel.CashSchedule{}
		for _, line := range cb.Lines {
			if line.Account.AccountID == "" {
				continue
			}

			cs := new(ficomodel.CashSchedule)
			cs.ID = fmt.Sprintf("CB_%s_%04d", cb.ID, line.LineNo)
			cs.CompanyID = cb.CompanyID
			cs.Account = line.Account
			cs.SourceType = ficomodel.SubledgerCashBank
			cs.SourceJournalID = cb.ID
			cs.SourceLineNo = line.LineNo
			cs.Text = fmt.Sprintf("%s %s", cb.Text, line.Text)
			cs.Expected = cb.TrxDate
			cs.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, cb.Dimension)
			cs.Status = ficomodel.CashDraft

			cs.Amount = lo.Ternary(req.Direction == "CashOut", float64(-1), float64(1)) * line.Amount
			cs.Outstanding = cs.Amount
			cs.Settled = 0
			db.Save(cs)

			res = append(res, cs)
		}

		return res, nil

	default:
		return nil, fmt.Errorf("invalid: source type: %s", req.SourceType)
	}
}

func (o *ApplyLogic) getApplied(db *datahub.Hub, req *GetApplyDataRequest) ([]*ficomodel.CashSchedule, []*ficomodel.CashApply, error) {
	if req.ApplyFromJournalID == "" && req.ApplyFromRecordID == "" {
		return []*ficomodel.CashSchedule{}, []*ficomodel.CashApply{}, nil
	}

	var wheres *dbflex.Filter
	if req.ApplyFromRecordID != "" {
		wheres = dbflex.Eq("Source.RecordID", req.ApplyFromRecordID)
	} else {
		wheres = dbflex.Eqs(
			"Source.Module", req.ApplyFromType,
			"Source.JournalID", req.ApplyFromJournalID,
			"Source.LineNo", req.ApplyFromLineNo,
		)
	}

	applies, _ := datahub.FindByFilter(db, new(ficomodel.CashApply), wheres)

	toIDs := lo.FindUniques(lo.Map(applies, func(a *ficomodel.CashApply, _ int) string {
		return a.ApplyTo.RecordID
	}))

	scheds, _ := datahub.FindByFilter(db, new(ficomodel.CashSchedule), dbflex.In("_id", toIDs...))

	return scheds, applies, nil
}

type GetApplyToResponse struct {
	Module    string
	JournalID string
}

func (o *ApplyLogic) GetApplyTo(ctx *kaos.Context, req *GetApplyDataRequest) ([]GetApplyToResponse, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}
	if req == nil {
		req = new(GetApplyDataRequest)
	}

	schedules := []ficomodel.CashSchedule{}
	err := db.Gets(new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.StartWith("SourceJournalID", req.SourceJournalID),
		),
	), &schedules)
	if err != nil {
		return nil, fmt.Errorf("error when get cash schedule: %s", err.Error())
	}

	if len(schedules) == 0 {
		return []GetApplyToResponse{}, nil
	}

	ids := lo.Map(schedules, func(m ficomodel.CashSchedule, index int) string {
		return m.ID
	})

	applies := []ficomodel.CashApply{}
	err = db.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
		dbflex.In("Source.RecordID", ids...),
	), &applies)
	if err != nil {
		return nil, fmt.Errorf("error when get cash apply: %s", err.Error())
	}

	applyIds := lo.Map(applies, func(m ficomodel.CashApply, index int) string {
		return m.ApplyTo.RecordID
	})

	err = db.Gets(new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", applyIds...),
			dbflex.Eq("Status", ficomodel.CashSettled),
		),
	), &schedules)
	if err != nil {
		return nil, fmt.Errorf("error when get cash schedule: %s", err.Error())
	}

	mapSchedule := lo.Associate(schedules, func(detail ficomodel.CashSchedule) (string, string) {
		return detail.ID, detail.SourceJournalID
	})

	result := make([]GetApplyToResponse, len(mapSchedule))
	i := 0
	for _, app := range applies {
		if _, ok := mapSchedule[app.ApplyTo.RecordID]; ok {
			result[i] = GetApplyToResponse{
				Module:    app.ApplyTo.Module.String(),
				JournalID: app.ApplyTo.JournalID,
			}

			i++
		}
	}

	return result, nil
}
