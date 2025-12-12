package ficologic

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/datahub"
)

func CalcMoneyCurrency(h *datahub.Hub, value float64, date *time.Time, groupID, fromCurrencyID string, toCurrencyID string) (float64, error) {
	if h == nil {
		return value, errors.New("missing: db")
	}

	if fromCurrencyID == "" || toCurrencyID == "" {
		return value, errors.New("missing: urrency")
	}

	if fromCurrencyID == toCurrencyID {
		return value, nil
	}

	var where *dbflex.Filter
	if date == nil {
		where = dbflex.And(
			dbflex.Eq("GroupID", groupID),
			dbflex.Or(dbflex.Eqs("FromCurrencyID", fromCurrencyID, "ToCurrencyID", toCurrencyID),
				dbflex.Eqs("ToCurrencyID", fromCurrencyID, "FromCurrencyID", toCurrencyID)))
	} else {
		where = dbflex.And(
			dbflex.Eq("GroupID", groupID),
			dbflex.Gte("ExchDate", *date),
			dbflex.Or(dbflex.Eqs("FromCurrencyID", fromCurrencyID, "ToCurrencyID", toCurrencyID),
				dbflex.Eqs("ToCurrencyID", fromCurrencyID, "FromCurrencyID", toCurrencyID)))
	}

	xr, err := datahub.GetByParm(h, new(ficomodel.ExchangeRate), dbflex.NewQueryParam().
		SetSort("-ExchDate").
		SetWhere(where))
	if err != nil {
		return value, err
	}

	if xr.Rate == 0 {
		xr.Rate = 1
	}

	if xr.ToCurrencyID == fromCurrencyID {
		return value / xr.Rate, nil
	}

	return value * xr.Rate, nil
}
