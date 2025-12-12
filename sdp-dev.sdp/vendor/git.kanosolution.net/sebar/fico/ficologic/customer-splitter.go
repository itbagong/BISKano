package ficologic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/datahub"
)

type CustomerSplitOpt struct {
	CompanyID       string
	AccountID       string
	SourceType      string
	SourceJournalID string
	SourceLineNo    int
	SourceStatus    string
}

func NewCustomerTrxSplitter(db *datahub.Hub) *TrxSplitter[*ficomodel.CustomerTransaction, CustomerSplitOpt] {
	c := new(TrxSplitter[*ficomodel.CustomerTransaction, CustomerSplitOpt])
	c.db = db
	c.provider = new(customerSplitProvider)
	return c
}

type customerSplitProvider struct {
}

func (c *customerSplitProvider) GetTransactions(db *datahub.Hub, opt CustomerSplitOpt) ([]*ficomodel.CustomerTransaction, error) {
	wheres := []*dbflex.Filter{
		dbflex.Eq("CompanyID", opt.CompanyID),
		dbflex.Eq("SourceType", opt.SourceType),
		dbflex.Eq("SourceJournalID", opt.SourceJournalID),
		dbflex.Eq("SourceLineNo", opt.SourceLineNo),
		dbflex.Eq("Status", opt.SourceStatus),
	}
	txs, err := datahub.FindByFilter(db, new(ficomodel.CustomerTransaction), dbflex.And(wheres...))
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (c *customerSplitProvider) Split(
	db *datahub.Hub,
	sources []*ficomodel.CustomerTransaction,
	amt float64, newStatus string,
	opt CustomerSplitOpt) ([]*ficomodel.CustomerTransaction, []*ficomodel.CustomerTransaction, error) {

	sourceTrxs := []*ficomodel.CustomerTransaction{}
	newTrxs := []*ficomodel.CustomerTransaction{}
	settledAmountSum := float64(0)
	for _, source := range sources {
		settleAmt, err := validateQty(source.Amount, amt)
		if err != nil {
			return nil, nil, err
		}

		source.Amount -= settleAmt
		if source.Amount == 0 {
			if err := db.Delete(source); err != nil {
				return nil, nil, fmt.Errorf("delete fail: %s", err.Error())
			}
		} else {
			sourceTrxs = append(sourceTrxs, source)
		}

		newTrx := new(ficomodel.CustomerTransaction)
		*newTrx = *source
		newTrx.ID = ""
		newTrx.Amount = settleAmt
		newTrx.Status = ficomodel.AmountStatus(newStatus)
		if err := db.Save(newTrx); err != nil {
			return nil, nil, fmt.Errorf("save trx: %s", err.Error())
		}
		newTrxs = append(newTrxs, newTrx)

		settledAmountSum += settleAmt
		if settledAmountSum == amt {
			break
		}
	}

	for _, tx := range sourceTrxs {
		if err := db.Save(tx); err != nil {
			return nil, nil, fmt.Errorf("save source: %s", err.Error())
		}
	}

	return sourceTrxs, newTrxs, nil
}
