package ficologic

import (
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

type BalanceHubProvider[B, T orm.DataModel, O any] interface {
	BalanceFilter(obj B, opt O) *dbflex.Filter
	TransactionFilter(obj T, opt O) *dbflex.Filter
	BalanceGrouping(obj B, opt O) string
	AggregateBalance(origModel B, balModels []B, add bool, opt O) error
	GetTransactions(db *datahub.Hub,
		dateFrom, dateTo *time.Time,
		initialWhere *dbflex.Filter,
		includeDateTo bool,
		opt O) ([]T, error)
	Update(bals []B, trxs []T, balDate *time.Time, deduct bool) ([]B, error)
}

type BalanceHub[B, T orm.DataModel, O any] struct {
	db       *datahub.Hub
	ev       kaos.EventHub
	provider BalanceHubProvider[B, T, O]

	refModel    B // IMPORTANT: this ref should be imutable and not use to store data in any function
	refTrxModel T
}

func NewBalanceHub[B, T orm.DataModel, O any](h *datahub.Hub, pvd BalanceHubProvider[B, T, O], ev kaos.EventHub, grOpt O) *BalanceHub[B, T, O] {
	lb := &BalanceHub[B, T, O]{}
	lb.db = h
	lb.ev = ev
	lb.provider = pvd
	return lb
}

func (o *BalanceHub[B, T, O]) Get(balDate *time.Time, getOpt O) ([]B, error) {
	var (
		ssDate *time.Time
	)

	refModel, err := reflector.CreateFromPtr(o.refModel, false)
	if err != nil {
		return nil, fmt.Errorf("invalid refmodel: %s", err.Error())
	}

	// get snapshot date
	//origWhere := o.BalanceFilterFn(o.refModel, getOpt)
	origWhere := o.provider.BalanceFilter(o.refModel, getOpt)
	deduct := balDate != nil

	if balDate != nil {
		ssDate, err = GetSnapshotRecord(o.db, refModel, origWhere, "BalanceDate", balDate, false)
		if err != nil {
			return nil, fmt.Errorf("get balance snapshot date: %s", err.Error())
		}
	}

	// get initial balances
	wheres := []*dbflex.Filter{origWhere}
	if balDate == nil {
		wheres = append(wheres, dbflex.Eq("BalanceDate", nil))
	} else {
		wheres = append(wheres, dbflex.Eq("BalanceDate", ssDate))
	}
	rawBalances, err := datahub.FindByFilter(o.db, refModel, dbflex.And(wheres...))
	if err != nil {
		return nil, nil
	}

	// group the balances
	groupRawBalances := lo.MapToSlice(lo.GroupBy(rawBalances, func(bal B) string {
		return o.provider.BalanceGrouping(bal, getOpt)
	}), func(k string, v []B) []B {
		return v
	})

	// factorizing trx
	resBals := []B{}
	lo.ForEach(groupRawBalances, func(bals []B, idx int) {
		gbItem, _ := reflector.CreateFromPtr(o.refModel, false)
		o.provider.AggregateBalance(gbItem, bals, true, getOpt)
		itemDim, _ := reflector.From(gbItem).Get("Dimension")

		gbTrxItem, _ := reflector.CreateFromPtr(o.refTrxModel, false)
		reflector.From(gbTrxItem).Set("Dimension", itemDim).Flush()
		initialWhere := o.provider.TransactionFilter(gbTrxItem, getOpt)
		if deduct {
			transactions, _ := o.provider.GetTransactions(o.db, balDate, ssDate, initialWhere, true, getOpt)
			for _, tr := range transactions {
				reflector.From(tr).Set("Dimension", itemDim).Flush()
			}
			bals, err := o.provider.Update([]B{gbItem}, transactions, balDate, true)
			if err != nil {
				resBals = append(resBals, gbItem)
			} else {
				resBals = append(resBals, bals...)
			}
		} else {
			resBals = append(resBals, gbItem)
		}
	})

	// final balancing
	// group the balances
	finalRawBals := lo.MapToSlice(lo.GroupBy(resBals, func(bal B) string {
		return o.provider.BalanceGrouping(bal, getOpt)
	}), func(k string, v []B) []B {
		return v
	})

	// final aggregate
	finalBals := []B{}
	for _, frBals := range finalRawBals {
		f, e := reflector.CreateFromPtr(frBals[0], true)
		if e != nil {
			return nil, e
		}
		o.provider.AggregateBalance(f, frBals, true, getOpt)
		finalBals = append(finalBals, f)
	}

	return finalBals, nil
}

func (o *BalanceHub[B, T, O]) Sync(balDate *time.Time, getOpt O) ([]B, error) {
	refModel, err := reflector.CreateFromPtr(o.refModel, false)
	if err != nil {
		return nil, fmt.Errorf("invalid refmodel: %s", err.Error())
	}

	// get snapshot date
	origBalanceFilter := o.provider.BalanceFilter(refModel, getOpt)
	ssDate, err := GetSnapshotRecord(o.db, refModel, origBalanceFilter, "BalanceDate", balDate, true)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if ssDate != nil && balDate != nil && ssDate.After(*balDate) {
		return nil, fmt.Errorf("sync date is less than last snapshot date")
	}

	// get snapshot balance, before expected date
	var bals []B
	if ssDate != nil {
		BalanceFilters := []*dbflex.Filter{origBalanceFilter, dbflex.Eq("BalanceDate", ssDate)}
		bals, err = datahub.FindByFilter(o.db, refModel, dbflex.And(BalanceFilters...))
		if err != nil {
			return nil, fmt.Errorf("get snapshot balance: %s", err.Error())
		}
	}

	// get trx, trx date > ssdate
	refTrxModel, _ := reflector.CreateFromPtr(o.refTrxModel, false)
	origTrxWhere := o.provider.TransactionFilter(refTrxModel, getOpt)
	trxs, err := o.provider.GetTransactions(o.db, ssDate, balDate, origTrxWhere, true, getOpt)
	if err != nil {
		return nil, err
	}
	bals, err = o.provider.Update(bals, trxs, balDate, false)
	if err != nil {
		return nil, err
	}

	// delete all balances per expected dates
	deleteWhere := dbflex.And(origBalanceFilter, dbflex.Eq("BalanceDate", balDate))
	o.db.DeleteByFilter(refModel, deleteWhere)

	for _, bal := range bals {
		err := reflector.From(bal).Set("ID", "").Set("BalanceDate", balDate).Flush()
		if err != nil {
			return nil, fmt.Errorf("assign value: %s", err.Error())
		}
		if err = o.db.Insert(bal); err != nil {
			return nil, err
		}
	}

	if balDate != nil {
		o.Sync(nil, getOpt)
	}

	return bals, nil
}
