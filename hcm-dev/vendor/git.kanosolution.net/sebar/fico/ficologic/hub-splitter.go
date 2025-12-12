package ficologic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/datahub"
)

type TrxSplitProvider[M orm.DataModel, O any] interface {
	GetTransactions(db *datahub.Hub, opt O) ([]M, error)
	Split(db *datahub.Hub, sources []M, qty float64, newStatus string, opt O) ([]M, []M, error)
}

type TrxSplitter[model orm.DataModel, splitOpt any] struct {
	db       *datahub.Hub
	provider TrxSplitProvider[model, splitOpt]
}

func (o *TrxSplitter[model, splitOpt]) Split(qty float64, newStatus string, opt splitOpt) ([]model, []model, error) {
	if o.db == nil {
		return nil, nil, fmt.Errorf("db is nil")
	}

	// get source trxs
	sources, err := o.provider.GetTransactions(o.db, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("get transaction sources: %s", err.Error())
	}

	// split it
	srcs, splits, err := o.provider.Split(o.db, sources, qty, newStatus, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("split transaction: %s", err.Error())
	}

	return srcs, splits, nil
}
