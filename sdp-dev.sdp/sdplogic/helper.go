package sdplogic

import (
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO"
	}
	return coID
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}

// GetStartOfDayInJakarta converts the given date to Jakarta time and returns the start of the day (midnight).
func GetStartOfDay(date time.Time) time.Time {
	// Load Jakarta timezone
	tzJakarta, _ := time.LoadLocation("Asia/Jakarta")

	// Convert to Jakarta timezone and set time to the start of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, tzJakarta)

	return startOfDay
}
