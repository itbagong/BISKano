package ficologic

import "time"

func TrimDate(dt time.Time) time.Time {
	return time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, dt.Location())
}

func EoD(dt time.Time) time.Time {
	return TrimDate(dt).AddDate(0, 0, 1).Add(-1 * time.Microsecond)
}

func NextDay(dt time.Time) time.Time {
	return TrimDate(dt).AddDate(0, 0, 1)
}

func FoMo(dt time.Time) time.Time {
	return time.Date(dt.Year(), dt.Month(), 1, 0, 0, 0, 0, dt.Location())
}

func EoMo(dt time.Time) time.Time {
	return FoMo(dt).AddDate(0, 1, 0).Add(-1 * time.Microsecond)
}

func EoMoDateOnly(dt time.Time) time.Time {
	return TrimDate(EoMo(dt))
}

func FoYr(dt time.Time) time.Time {
	return time.Date(dt.Year(), 1, 1, 0, 0, 0, 0, dt.Location())
}

func EoYr(dt time.Time) time.Time {
	return FoYr(dt).AddDate(1, 0, 0).Add(-1 * time.Microsecond)
}

func EoYrDateOnly(dt time.Time) time.Time {
	return EoYr(dt).Add(-1 * time.Microsecond)
}

func ChangeLocation(dt time.Time, loc time.Location) time.Time {
	return time.Date(dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond(), &loc)
}
