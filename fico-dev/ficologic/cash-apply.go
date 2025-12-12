package ficologic

import (
	"fmt"
	"math"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type CashApplyAdjustment struct {
	Dimension tenantcoremodel.Dimension
	From      ficomodel.SubledgerAccount
	To        ficomodel.SubledgerAccount
	Text      string
	Amount    float64
}

type CashApplyTo struct {
	ApplyToRecID string
	Amount       float64
	IsUnchecked  bool
	Adjustment   []CashApplyAdjustment
}

type CashApplyRequest struct {
	Db          *datahub.Hub
	CompanyID   string
	TrxDate     time.Time
	SourceRecID string
	Applies     []CashApplyTo
}

func ApplyCashShedule(req CashApplyRequest) error {
	var err error

	// get source
	source := new(ficomodel.CashSchedule)
	err = req.Db.GetByID(source, req.SourceRecID)
	if err != nil {
		return fmt.Errorf("apply source: %s: %s", req.SourceRecID, err.Error())
	}

	adjustments := []CashApplyAdjustment{}

	for _, applySetup := range req.Applies {
		apply, err := datahub.GetByID(req.Db, new(ficomodel.CashSchedule), applySetup.ApplyToRecID)
		if err != nil {
			return fmt.Errorf("cash schedule apply: %s", err.Error())
		}

		if applySetup.Amount == 0 {
			continue
		}

		// check if apply is removed, reset outstanding and settled
		if applySetup.IsUnchecked {
			source.Outstanding += applySetup.Amount
			source.Settled -= applySetup.Amount
			apply.Outstanding += applySetup.Amount
			apply.Settled -= applySetup.Amount

			records := []orm.DataModel{source, apply}
			for _, rec := range records {
				if err := req.Db.Save(rec); err != nil {
					return fmt.Errorf("save %s error: %s", rec.TableName(), err.Error())
				}
			}

			// delete cash apply
			err = req.Db.DeleteByFilter(new(ficomodel.CashApply), dbflex.Eq("ApplyTo.RecordID", apply.ID))
			if err != nil {
				return fmt.Errorf("delete %s error: %s", apply.ID, err.Error())
			}
		} else {
			// check if apply has already been done
			err = req.Db.GetByFilter(new(ficomodel.CashApply), dbflex.Eq("ApplyTo.RecordID", apply.ID))

			// do apply if it has never been done
			if err != nil {
				if math.Abs(applySetup.Amount) > math.Abs(source.Outstanding) {
					return fmt.Errorf("apply amount is not valid")
				}

				if math.Abs(applySetup.Amount) > math.Abs(apply.Outstanding) {
					return fmt.Errorf("apply amount is not valid")
				}

				source.Outstanding -= applySetup.Amount
				source.Settled += applySetup.Amount
				apply.Outstanding -= applySetup.Amount
				apply.Settled += applySetup.Amount

				schedApply := ficomodel.CashApply{
					CompanyID: req.CompanyID,
					Source: tenantcoremodel.TrxMetadata{
						Module:    source.SourceType,
						JournalID: source.SourceJournalID,
						LineNo:    source.SourceLineNo,
						RecordID:  source.ID,
					},
					ApplyTo: tenantcoremodel.TrxMetadata{
						Module:    apply.SourceType,
						JournalID: apply.SourceJournalID,
						LineNo:    apply.SourceLineNo,
						RecordID:  apply.ID,
					},
					Amount: apply.Amount,
				}

				// build deduction
				if len(applySetup.Adjustment) > 0 {
					for _, adj := range applySetup.Adjustment {
						adj.From = apply.Account
						adjustments = append(adjustments, adj)

						// add adjustment to cash apply
						schedApply.Adjustment = append(schedApply.Adjustment, &ficomodel.CashAdjustment{
							Account: adj.From,
							//OffsetAccount: adj.To,
							//TrxType:       "ApplyAdjustment",
							Amount: adj.Amount,
						})
					}
				}

				records := []orm.DataModel{source, apply, &schedApply}
				for _, rec := range records {
					if err := req.Db.Save(rec); err != nil {
						return fmt.Errorf("save %s error: %s", rec.TableName(), err.Error())
					}
				}
			}
		}
	}

	// if len of deductions > 0, combine all of them into single ledger journal with TrxType=ApplyAdjustment
	if len(adjustments) > 0 {
		adjustmentLines := []ficomodel.JournalLine{}
		for _, adj := range adjustments {
			line := ficomodel.JournalLine{
				Account:       adj.From,
				OffsetAccount: adj.To,
				TrxType:       "ApplyAdjustment",
				Amount:        adj.Amount,
			}
			adjustmentLines = append(adjustmentLines, line)
		}

		journal := new(ficomodel.LedgerJournal)
		journal.CompanyID = req.CompanyID
		journal.Text = "Apply Adjustment"
		journal.TrxDate = req.TrxDate
		journal.Lines = adjustmentLines
		journal.References = tenantcoremodel.References{}.Set("SourceRecID", source.ID)

		journal.JournalTypeID = "DEF" // TODO: ini harus dibuat paramaterized di config

		req.Db.Save(journal)
	}

	return nil
}

func FindCashSchedule(db *datahub.Hub, companyID string, source tenantcoremodel.TrxModule, sourceID string, voucherNo string, lineNo int) []*ficomodel.CashSchedule {
	wheres := []*dbflex.Filter{dbflex.Eqs(
		"CompanyID", companyID, "SourceType", source, "SourceJournalID", sourceID)}
	if voucherNo != "" {
		wheres = append(wheres, dbflex.Eq("VoucherNo", lineNo))
	}
	if lineNo > 0 {
		wheres = append(wheres, dbflex.Eq("SourceLineNo", lineNo))
	}
	res, _ := datahub.FindByFilter(db, new(ficomodel.CashSchedule), dbflex.And(wheres...))
	return res
}
