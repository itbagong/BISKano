package ficologic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

func NewCashSchedule(companyID string, sourceType tenantcoremodel.TrxModule, sourceID string, sourceLineNo int, amount float64) *ficomodel.CashSchedule {
	return &ficomodel.CashSchedule{CompanyID: companyID, SourceType: sourceType, SourceJournalID: sourceID, SourceLineNo: sourceLineNo,
		Status: ficomodel.CashScheduled, Amount: amount}
}

func CashScheduleSplit(h *datahub.Hub, companyID, sourceType, sourceID string, origStatus ficomodel.CashScheduleStatus,
	amt float64, destStatus ficomodel.CashScheduleStatus) ([]*ficomodel.CashSchedule, []*ficomodel.CashSchedule, error) {
	sources := []*ficomodel.CashSchedule{}
	targets := []*ficomodel.CashSchedule{}

	db, err := h.BeginTx()
	if err != nil {
		return sources, targets, fmt.Errorf("invalid: %s", "transaction is not supported")
	}
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()

	schedules, err := datahub.FindByFilter(db, new(ficomodel.CashSchedule),
		dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceID", sourceID, "Status", origStatus))
	if err != nil {
		db.Rollback()
		return []*ficomodel.CashSchedule{}, []*ficomodel.CashSchedule{}, err
	}

	remaining := amt
	totalSplitAmt := float64(0)
	for _, sc := range schedules {
		splitAmt := float64(0)
		if sc.Amount > remaining {
			splitAmt = remaining
		} else if sc.Amount < remaining {
			splitAmt = sc.Amount
		} else {
			splitAmt = remaining
		}
		remaining -= splitAmt
		totalSplitAmt += splitAmt

		sc.Amount -= splitAmt
		if sc.Amount == 0 {
			db.Delete(sc)
		} else {
			sources = append(sources, sc)
		}
		target := new(ficomodel.CashSchedule)
		*target = *sc
		target.ID = ""
		target.Amount = splitAmt
		target.Status = destStatus
		targets = append(targets, target)

		if totalSplitAmt == amt {
			break
		}
	}

	for totalSplitAmt != amt {
		return []*ficomodel.CashSchedule{}, []*ficomodel.CashSchedule{}, errors.New("not enough amount")
	}

	for _, src := range sources {
		if e := db.Update(src); e != nil {
			db.Rollback()
			return []*ficomodel.CashSchedule{}, []*ficomodel.CashSchedule{}, err
		}
	}

	for _, target := range targets {
		if e := db.Update(target); e != nil {
			db.Rollback()
			return []*ficomodel.CashSchedule{}, []*ficomodel.CashSchedule{}, err
		}
	}
	db.Commit()
	return sources, targets, nil
}

func SettleCashSchedule(db *datahub.Hub, from *ficomodel.CashSchedule, tos ...*ficomodel.CashSchedule) ([]*ficomodel.CashApply, error) {
	amt := from.Outstanding
	cas := []*ficomodel.CashApply{}

	for _, to := range tos {
		if from.CompanyID != to.CompanyID && from.CompanyID != "" {
			return nil, fmt.Errorf("apply should be from transaction in same company")
		}

		if from.Outstanding == 0 {
			return nil, fmt.Errorf("insufficient amount to apply: %s, %s, %s", from.SourceType, from.SourceJournalID, from.ID)
		}

		if from.Outstanding == 0 {
			return nil, fmt.Errorf("already fully settled %s, %s, %s", to.SourceType, to.SourceJournalID, to.ID)
		}

		if amt == 0 {
			amt = -to.Outstanding
		}

		if to.Outstanding == 0 {
			continue
		}

		if amt < from.Outstanding {
			amt = from.Outstanding
		}

		from.Settled += amt
		from.Outstanding -= amt
		from.Calc()

		to.Settled += amt
		to.Outstanding -= amt
		to.Calc()

		ca := new(ficomodel.CashApply)
		ca.CompanyID = from.CompanyID
		ca.Source = tenantcoremodel.TrxMetadata{
			Module:    from.SourceType,
			JournalID: from.SourceJournalID,
			LineNo:    from.SourceLineNo,
			RecordID:  from.ID,
		}
		ca.ApplyTo = tenantcoremodel.TrxMetadata{
			Module:    to.SourceType,
			JournalID: to.SourceJournalID,
			LineNo:    to.SourceLineNo,
			RecordID:  to.ID,
		}
		ca.Amount = amt

		db.Save(ca)
		db.Save(to)

		cas = append(cas, ca)

		if from.Outstanding == 0 {
			break
		}
	}
	db.Save(from)

	return cas, nil
}

func FindOrCreateCashSchedule(db *datahub.Hub, companyID string, sourceType tenantcoremodel.TrxModule, journalID string, lineNo int, status ficomodel.CashScheduleStatus) ([]*ficomodel.CashSchedule, error) {
	filter := dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceJournalID", journalID, "SourceLineNo", lineNo)
	if status != "" {
		filter = dbflex.And(filter, dbflex.Eq("Status", status))
	}

	schedules, err := datahub.FindByFilter(db, new(ficomodel.CashSchedule), filter)
	if err != nil {
		return nil, err
	}
	if len(schedules) == 0 {
		cs := new(ficomodel.CashSchedule)
		cs.CompanyID = companyID
		cs.SourceType = sourceType
		cs.SourceJournalID = journalID
		cs.SourceLineNo = lineNo
		cs.Status = status
		schedules = append(schedules, cs)
	}

	return schedules, nil
}

func SplitCashSchedule(db *datahub.Hub, companyID string, sourceType tenantcoremodel.TrxModule, journalID string, lineNo int, status ficomodel.CashScheduleStatus, newStatus ficomodel.CashScheduleStatus, amt float64) ([]*ficomodel.CashSchedule, []*ficomodel.CashSchedule, error) {
	filter := dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceJournalID", journalID, "SourceLineNo", lineNo)
	if status != "" {
		filter = dbflex.And(filter, dbflex.Eq("Status", status))
	}

	schedules, err := datahub.FindByFilter(db, new(ficomodel.CashSchedule), filter)
	if err != nil {
		return nil, nil, err
	}
	if newStatus == "" {
		return nil, nil, fmt.Errorf("status is mandatory")
	}

	sources := []*ficomodel.CashSchedule{}
	dests := []*ficomodel.CashSchedule{}

	splittedAmt := float64(0)
	applyAmt := amt
	for _, schedule := range schedules {
		if schedule.Outstanding < applyAmt {
			applyAmt = schedule.Outstanding
		}

		schedule.Outstanding -= applyAmt
		schedule.Settled += applyAmt
		schedule.Amount -= applyAmt
		schedule.Calc()
		sources = append(sources, schedule)

		dest := new(ficomodel.CashSchedule)
		*dest = *schedule
		dest.ID = ""
		dest.Amount = applyAmt
		dest.Status = newStatus
		dest.Calc()
		dests = append(dests, dest)

		splittedAmt += applyAmt
		if splittedAmt == amt {
			break
		}
		applyAmt = amt - splittedAmt
	}

	for _, source := range sources {
		if source.Amount == 0 {
			db.Delete(source)
		} else {
			db.Save(source)
		}
	}

	for _, dest := range dests {
		db.Save(dest)
	}

	return sources, dests, nil
}
