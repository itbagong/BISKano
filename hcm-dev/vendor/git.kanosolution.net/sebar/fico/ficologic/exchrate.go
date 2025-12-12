package ficologic

import (
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/datahub"
)

func ConvertCurrency(h *datahub.Hub, dt time.Time, amt float64, group, from, to string) (float64, error) {
	w := dbflex.And(dbflex.Eq("GroupID", group), dbflex.Eq("FromCurrencyID", from), dbflex.Eq("ToCurrencyID", to), dbflex.Lte("ExchDate", dt))
	p := dbflex.NewQueryParam().SetWhere(w).SetSort("-ExchDate")
	r, e := datahub.GetByParm(h, new(ficomodel.ExchangeRate), p)
	if e == io.EOF {
		w = dbflex.And(dbflex.Eq("GroupID", group), dbflex.Eq("ToCurrencyID", from), dbflex.Eq("FromCurrencyID", to), dbflex.Lte("ExchDate", dt))
		p = dbflex.NewQueryParam().SetWhere(w).SetSort("-ExchDate")
		r, e = datahub.GetByParm(h, new(ficomodel.ExchangeRate), p)
		if e != nil {
			return amt, fmt.Errorf("exchange rate: %s", e.Error())
		}
		return amt / r.Rate, nil
	} else if e != nil {
		return amt, fmt.Errorf("exchange rate: %s", e.Error())
	}
	return r.Rate * amt, nil
}
