package ficologic

import (
	"errors"
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func (o *ApplyLogic) Save(ctx *kaos.Context, req []*ficomodel.CashApply) ([]*ficomodel.CashSchedule, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if coID == "" {
		return nil, errors.New("missing: company ID")
	}

	css := []*ficomodel.CashSchedule{}
	for _, apply := range req {
		apply.CompanyID = coID
		cs1, cs2, err := saveApply(db, apply)
		if err != nil {
			return nil, err
		}
		if cs1 != nil {
			css = append(css, cs1)
		}
		css = append(css, cs2)
	}
	recalcRes := recalcSched(db, css...)
	return recalcRes, nil
}

func saveApply(db *datahub.Hub, apply *ficomodel.CashApply) (*ficomodel.CashSchedule, *ficomodel.CashSchedule, error) {
	var (
		cs1, cs2 *ficomodel.CashSchedule
		err      error
	)

	filters := []*dbflex.Filter{}
	cs1, err = datahub.GetByID(db, new(ficomodel.CashSchedule), apply.Source.RecordID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: source: %s, %s", apply.Source.RecordID, err.Error())
	}
	filters = append(filters, dbflex.Eq("Source.RecordID", cs1.ID))

	cs2, err = datahub.GetByID(db, new(ficomodel.CashSchedule), apply.ApplyTo.RecordID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: apply to: %s, %s", apply.ApplyTo.RecordID, err.Error())
	}
	filters = append(filters, dbflex.Eq("ApplyTo.RecordID", cs2.ID))

	ca, err := datahub.GetByFilter(db, new(ficomodel.CashApply), dbflex.And(filters...))
	if err == io.EOF {
		ca = new(ficomodel.CashApply)
		ca.Source.RecordID = cs1.ID
		ca.Source.Module = cs1.SourceType
		ca.Source.JournalID = cs1.SourceJournalID
		ca.Source.LineNo = cs1.SourceLineNo
		ca.ApplyTo.RecordID = cs2.ID
		ca.ApplyTo.Module = cs2.SourceType
		ca.ApplyTo.JournalID = cs2.SourceJournalID
		ca.ApplyTo.LineNo = cs2.SourceLineNo
		ca.Draft = len(apply.Adjustment) > 0
	} else if err != nil {
		return nil, nil, fmt.Errorf("invalid: cash apply: %s", err.Error())
	}

	if len(apply.Adjustment) == 0 {
		ca.ApplyAmount = apply.Amount
	} else {
		for _, applyAdj := range apply.Adjustment {
			caAdj, has := lo.Find(ca.Adjustment, func(adj *ficomodel.CashAdjustment) bool {
				return adj.Account.AccountType == applyAdj.Account.AccountType &&
					adj.Account.AccountID == applyAdj.Account.AccountID
			})
			if !has {
				caAdj = applyAdj
				ca.Adjustment = append(ca.Adjustment, caAdj)
			}
			caAdj.Amount = applyAdj.Amount
		}
	}

	ca.Amount = ca.ApplyAmount + lo.SumBy(ca.Adjustment, func(adj *ficomodel.CashAdjustment) float64 {
		return adj.Amount
	})
	db.Save(ca)

	return cs1, cs2, nil
}

func recalcSched(db *datahub.Hub, sources ...*ficomodel.CashSchedule) []*ficomodel.CashSchedule {
	csIDs := lo.Keys(lo.GroupBy(sources, func(s *ficomodel.CashSchedule) string {
		return s.ID
	}))

	journals := map[string]orm.DataModel{}
	css := []*ficomodel.CashSchedule{}
iterateCsID:
	for _, csID := range csIDs {
		cs, err := datahub.GetByID(db, new(ficomodel.CashSchedule), csID)
		if err != nil {
			continue
		}

		// get the amount
		applies, _ := datahub.FindByFilter(db, new(ficomodel.CashApply), dbflex.Eqs("Source.RecordID", cs.ID))
		amount := lo.SumBy(applies, func(ca *ficomodel.CashApply) float64 {
			return ca.ApplyAmount
		})

		applies, _ = datahub.FindByFilter(db, new(ficomodel.CashApply), dbflex.Eqs("ApplyTo.RecordID", cs.ID))
		amount += lo.SumBy(applies, func(ca *ficomodel.CashApply) float64 {
			return ca.ApplyAmount
		})

		// if it is a DRAFT then update the amount of respective line, else assign the amount, settled, outstanding and status
		if cs.Status == ficomodel.CashDraft {
			cs.Amount = amount
			cs.Settled = amount
			cs.Outstanding = 0

			switch cs.SourceType {
			case ficomodel.SubledgerCashBank:
				var err error
				j, has := journals[cs.SourceJournalID].(*ficomodel.CashJournal)
				if !has {
					j, err = datahub.GetByID(db, new(ficomodel.CashJournal), cs.SourceJournalID)
					if err != nil {
						continue iterateCsID
					}
					journals[j.ID] = j
				}

				_, lineIndex, ok := lo.FindIndexOf(j.Lines, func(l ficomodel.JournalLine) bool {
					return l.LineNo == cs.SourceLineNo
				})
				if !ok {
					continue
				}
				j.Lines[lineIndex].Amount = lo.Ternary(amount < 0, -amount, amount)
			}
		} else {
			cs.Settled = amount
			cs.Outstanding = cs.Amount - cs.Settled

			if cs.Outstanding == 0 {
				cs.Status = ficomodel.CashSettled
			} else if cs.Outstanding == cs.Amount {
				cs.Status = ficomodel.CashScheduled
			} else {
				cs.Status = ficomodel.CashPartiallySettled
			}
		}
		db.Save(cs)
		css = append(css, cs)
	}

	for _, j := range journals {
		db.Save(j)
	}

	return css
}
