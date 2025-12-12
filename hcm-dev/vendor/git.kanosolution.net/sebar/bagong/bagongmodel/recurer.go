package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/datahub"
)

type Recurer interface {
	orm.DataModel
	GetDate() time.Time
	SetDate(time.Time)
}
type RecurerFrequency string

const (
	WEEKLY  RecurerFrequency = "WEEKLY"
	MONTHLY RecurerFrequency = "MONTHLY"
	YEARLY  RecurerFrequency = "YEARLY"
)

type RecuringParam struct {
	Freq      RecurerFrequency
	DateStart time.Time
	DateEnd   time.Time
}

func GetAddition(start time.Time, f RecurerFrequency) time.Time {
	switch f {
	case WEEKLY:
		return start.AddDate(0, 0, 7)
	case MONTHLY:
		return start.AddDate(0, 1, 0)
	case YEARLY:
	default:
		return start.AddDate(1, 0, 0)
	}
	return time.Time{}
}

// buat recurence dari r dan parameter param
// nilai id dari r akan dihapus dan dimulai dari r.Date+freq
func Recur(hub *datahub.Hub, r Recurer, param RecuringParam) {
	curDate := param.DateStart
	for {
		if curDate.After(param.DateEnd) {
			break
		}
		trxDate := r.GetDate()
		trxDate = GetAddition(trxDate, param.Freq)
		if trxDate.After(param.DateEnd) {
			break
		}
		r.SetID("")
		r.SetDate(trxDate)
		hub.Save(r)
		curDate = GetAddition(curDate, param.Freq)
	}
}
