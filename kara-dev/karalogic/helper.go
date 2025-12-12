package karalogic

import (
	"math"
	"strings"
	"time"

	"git.kanosolution.net/sebar/kara/karamodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

func getLocationTimeLoc(h *datahub.Hub, worklocID string, location *karamodel.WorkLocation) *time.Location {
	if location == nil {
		location = new(karamodel.WorkLocation)
		h.GetByID(location, worklocID)
		if location == nil {
			return time.Local
		}
	}

	loc, err := time.LoadLocation(location.TimeLoc)
	if err != nil {
		return time.Local
	}

	return loc
}

func date2DateTime(base time.Time, timeStr string, dayType string, location *time.Location) time.Time {
	tm := codekit.String2Date(timeStr, "HH:mm")
	if location != nil && base.Location() != location {
		base = base.In(location)
	}
	if dayType == "Next" {
		base = base.AddDate(0, 0, 1)
	} else if dayType == "Prev" {
		base = base.AddDate(0, 0, -1)
	}

	res := time.Date(base.Year(), base.Month(), base.Day(), tm.Hour(), tm.Minute(), 0, 0, base.Location())
	return res
}

func date2RangeDateTime(base time.Time, from, to string, location *time.Location) (time.Time, time.Time) {
	if location != nil && location != base.Location() {
		base = base.In(location)
	}

	nextDay := false
	isLapse := isLapseDay(from, to)
	if isLapse {
		hhmm := codekit.Date2String(base, "HH:mm")
		if strings.Compare(hhmm, from) == -1 {
			nextDay = true
		}
	}

	var (
		d1, d2 time.Time
	)

	if isLapse {
		if nextDay {
			d1 = date2DateTime(base, from, "Prev", location)
			d2 = date2DateTime(base, to, "", location)
		} else {
			d1 = date2DateTime(base, from, "", location)
			d2 = date2DateTime(base, to, "Next", location)
		}
	} else {
		d1 = date2DateTime(base, from, "", location)
		d2 = date2DateTime(base, to, "", location)
	}
	return d1, d2
}

func isLapseDay(from, to string) bool {
	return strings.Compare(from, to) == 1
}

func tenary[T any](condition bool, ok T, notOK T) T {
	if condition {
		return ok
	}
	return notOK
}

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func CalcDistance(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	// Earth radius, in meters. If want to use another unit, please change this accordingly
	r = 6378100

	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func combineDateTime(dt time.Time, tm string, tz *time.Location) time.Time {
	tme := codekit.String2Date(tm, "HH:mm")
	return time.Date(dt.Year(), dt.Month(), dt.Day(), tme.Hour(), tme.Minute(), 0, 0, tz)
}
