package karalogic

import (
	"git.kanosolution.net/sebar/kara/karaconfig"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

	"errors"
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
)

type UserTrx struct {
}

func (u *UserTrx) Create(ctx *kaos.Context, payload *karamodel.TrxRequest) (*karamodel.AttendanceTrx, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("missing: event hub")
	}

	userID, userIDIsOK := ctx.Data().Get("jwt_reference_id", "").(string)
	if !userIDIsOK {
		return nil, errors.New("missing: User")
	}

	var paramTrx karamodel.AttendanceTrx
	paramTrx.UserID = userID
	paramTrx.Op = karamodel.OpCode(payload.Op)
	paramTrx.WorkLocationID = payload.WorkLocationID
	paramTrx.TrxDate = payload.TrxDate
	paramTrx.Long = payload.Long
	paramTrx.Lat = payload.Lat
	if payload.TrxTime != "" {
		newDateTimeStr := fmt.Sprintf("%sT%s %s", payload.TrxDate.Format("2006-01-02"), payload.TrxTime, payload.TrxDate.Format("MST"))
		newDateTime, err := time.Parse("2006-01-02T15:04 MST", newDateTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid: time: %s", err.Error())
		}
		paramTrx.TrxDate = newDateTime
	}
	paramTrx.Ref1 = payload.Ref1

	// find employee
	employees := []tenantcoremodel.Employee{}
	query := dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", userID),
	)
	err := ev.Publish("/v1/tenant/employee/find", query, &employees, nil)
	if err != nil {
		return nil, err
	}

	// set dimension
	if len(employees) > 0 {
		paramTrx.Dimension = employees[0].Dimension
	}

	res, err := saveTrx(h, ev, &paramTrx, payload.ConfirmForReview)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func saveTrx(h *datahub.Hub, ev kaos.EventHub, trx *karamodel.AttendanceTrx, confirmForReview bool) (*karamodel.AttendanceTrx, error) {
	lastTrx, opErr := validateCheckOp(h, trx.UserID, trx.WorkLocationID, trx.TrxDate, trx.Op)
	if opErr != nil {
		return nil, fmt.Errorf("save trx: %s", opErr.Error())
	}

	worklocUsers := []*karamodel.WorkLocationUser{}
	err := h.GetsByFilter(new(karamodel.WorkLocationUser), dbflex.And(
		dbflex.Eq("UserID", trx.UserID), dbflex.Eq("WorkLocationID", trx.WorkLocationID),
		dbflex.Lte("From", trx.TrxDate), dbflex.Gte("To", trx.TrxDate),
	), &worklocUsers)
	if err != nil {
		return nil, fmt.Errorf("invalid_location: %s", err.Error())
	}

	lineFound := false
	ruleID := ""
	lineID := ""

findRuleLineID:
	for _, workLocUser := range worklocUsers {
		locationTimeLoc := getLocationTimeLoc(h, workLocUser.WorkLocationID, nil)
		ruleLines := []*karamodel.RuleLine{}
		h.GetsByFilter(new(karamodel.RuleLine), dbflex.Eq("RuleID", workLocUser.RuleID), &ruleLines)
		if len(ruleLines) == 0 {
			continue
		}
		for _, line := range ruleLines {
			var (
				d1, d2 time.Time
			)
			if trx.Op == karamodel.Checkin {
				d1, d2 = date2RangeDateTime(trx.TrxDate, line.CheckinStart, line.CheckinEnd, locationTimeLoc)
			} else if trx.Op == karamodel.Checkout {
				d1, d2 = date2RangeDateTime(trx.TrxDate, line.CheckoutStart, line.CheckoutEnd, locationTimeLoc)
			} else if trx.Op == karamodel.Transit {
				d1, d2 = date2RangeDateTime(trx.TrxDate, line.WorkStart, line.WorkEnd, locationTimeLoc)
			}
			if d1 == d2 || d1.After(d2) {
				return nil, fmt.Errorf("invalid_time")
			}
			if (d1 == trx.TrxDate || d1.Before(trx.TrxDate)) && (d2 == trx.TrxDate || d2.After(trx.TrxDate)) {
				ruleID = line.RuleID
				lineID = line.ID
				lineFound = true
				break findRuleLineID
			}
		}
	}

	if !lineFound && !confirmForReview {
		return nil, fmt.Errorf("need_review_confirm")
	}

	errLocation := validateTrxLocation(h, trx.WorkLocationID, trx.Long, trx.Lat)
	if errLocation != nil && !confirmForReview {
		return nil, err
	}

	// if a transit, create a checkout
	if trx.Op == karamodel.Transit {
		checkoutTrx := new(karamodel.AttendanceTrx)
		*checkoutTrx = *trx
		checkoutTrx.TrxDate = trx.TrxDate.Add(-5 * time.Second)
		checkoutTrx.Op = karamodel.Checkout
		_, err := saveTrx(h, ev, checkoutTrx, confirmForReview)
		if err != nil {
			return nil, fmt.Errorf("save trx: create auto checkout: %s", err.Error())
		}
	}

	// save intended trx
	trx.RuleID = ruleID
	trx.RuleLineID = lineID
	if lineFound && errLocation == nil {
		trx.Status = karamodel.TrxOK
	} else {
		trx.Status = karamodel.TrxNeedReview
	}

	// get user name
	getUserTopic := karaconfig.Config.GetUserTopic
	user := new(rbacmodel.User)
	getUserReq := &rbaclogic.GetUserByRequest{
		FindBy: "userid",
		FindID: trx.UserID,
	}
	err = ev.Publish(getUserTopic, getUserReq, user, nil)
	if err != nil {
		//trx.Name = user.DisplayName
		return nil, fmt.Errorf("missing: user")
	}
	trx.Name = user.DisplayName

	// calc hours if op is checkout
	// and if op is checkin and prev trx transit
	if trx.Op == karamodel.Checkout {
		trx.Hours = trx.TrxDate.Sub(lastTrx.TrxDate).Hours()
	} else if trx.Op == karamodel.Checkin && lastTrx != nil && lastTrx.Op == karamodel.Transit {
		trx.Hours = trx.TrxDate.Sub(lastTrx.TrxDate).Hours()
	}
	if err = h.Save(trx); err != nil {
		return nil, fmt.Errorf("save trx: %s", err.Error())
	}

	return trx, nil
}

func validateCheckOp(h *datahub.Hub, userid string, locationid string, date time.Time, op karamodel.OpCode) (*karamodel.AttendanceTrx, error) {
	lastTrx, err := getUserLastOP(h, userid)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("validate check op: %s", err.Error())
	}
	if err == io.EOF {
		if op != karamodel.Checkin {
			return nil, fmt.Errorf("checkin_only")
		}
		return nil, nil
	}

	if lastTrx.TrxDate.After(date) {
		return nil, errors.New("past_due_trx_date")
	}

	switch op {
	case karamodel.Checkout:
		// can checkout only after checkin
		if lastTrx.Op != karamodel.Checkin {
			return nil, fmt.Errorf("need_checkin")
		}
	case karamodel.Checkin:
		// check if still in same day
		if lastTrx.TrxDate.In(date.Location()).Day() == date.Day() {
			// can not checkin after checkin
			if lastTrx.Op == karamodel.Checkin {
				return nil, fmt.Errorf("checkin_already")
			}
		}
	case karamodel.Transit:
		// can not transit if not checkin
		if lastTrx.Op != karamodel.Checkin {
			return nil, fmt.Errorf("need_checkin")
		}
		// transit and checkin location shld be same
		if lastTrx.WorkLocationID != locationid {
			return nil, fmt.Errorf("invalid_location")
		}

	default:
		return nil, fmt.Errorf("invalid_op")
	}
	return lastTrx, nil
}

func getUserLastOP(h *datahub.Hub, userID string) (*karamodel.AttendanceTrx, error) {
	w := dbflex.Eq("UserID", userID)
	qp := dbflex.NewQueryParam().SetWhere(w).SetSort("-TrxDate", "-_id").SetTake(1)
	trx := new(karamodel.AttendanceTrx)
	err := h.GetByParm(trx, qp)
	if err != nil {
		return nil, err
	}
	return trx, nil
}

func validateTrxLocation(h *datahub.Hub, locationID string, long, lat float64) error {
	location := new(karamodel.WorkLocation)
	err := h.GetByID(location, locationID)
	if err != nil {
		return errors.New("invalid_location")
	}

	if location.DistanceTolerance == 0 {
		return nil
	}

	distance := CalcDistance(location.Location.Coordinates[1], location.Location.Coordinates[0], lat, long)
	if distance >= float64(location.DistanceTolerance) {
		return errors.New("out_of_distance")
	}

	return nil
}
