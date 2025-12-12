package karalogic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sort"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type MeLogic struct {
}

type StatusRespond struct {
	ID               string `json:"_id" bson:"_id"`
	CurrentOp        karamodel.OpCode
	WorkLocationID   string
	WorkLocationName string
	CheckinTime      *time.Time
	CheckoutTime     *time.Time
}

func (obj *MeLogic) Status(ctx *kaos.Context, payload string) (*StatusRespond, error) {
	res := new(StatusRespond)
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	meID := sebar.GetUserIDFromCtx(ctx)
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}

	// get company
	company, err := datahub.GetByParm(h, new(tenantcoremodel.Company), dbflex.NewQueryParam().
		SetWhere(dbflex.Eqs("_id", coID)))
	if err != nil {
		return nil, fmt.Errorf("error when get company: %s", err)
	}

	loc, err := time.LoadLocation(company.LocationCode)
	if err != nil {
		return nil, fmt.Errorf("error convert company location: %s", err)
	}

	now := time.Now().In(loc)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)

	attendances := []karamodel.AttendanceTrx{}
	err = h.Gets(new(karamodel.AttendanceTrx), dbflex.NewQueryParam().
		SetWhere(dbflex.And(
			dbflex.Eq("UserID", meID),
			dbflex.Gte("TrxDate", start),
			dbflex.Lt("TrxDate", end),
		)).SetSort("TrxDate"), &attendances)
	if err != nil {
		return nil, fmt.Errorf("error  when get attendance: %s", err)
	}

	for _, a := range attendances {
		res.WorkLocationID = a.WorkLocationID

		t := a.TrxDate
		if a.Op == karamodel.Checkin {
			res.CheckinTime = &t
			res.CurrentOp = karamodel.Checkin
		} else if a.Op == karamodel.Checkout {
			res.CheckoutTime = &t
			res.CurrentOp = karamodel.Checkout
		}
	}

	if len(attendances) == 0 {
		res.CurrentOp = karamodel.None
	}

	wl, _ := datahub.GetByID(h, new(karamodel.WorkLocation), res.WorkLocationID)
	res.WorkLocationName = wl.Name

	return res, nil
}

type HistoryRespond struct {
	CheckInID        string
	CheckIn          *time.Time
	CheckOutID       string
	CheckOut         *time.Time
	TotalHour        float64
	WorkLocationID   string
	WorkLocationName string
}

// type HistoryRespond []StatusRespond

func (obj *MeLogic) History(ctx *kaos.Context, payload *dbflex.QueryParam) ([]*HistoryRespond, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	meID := sebar.GetUserIDFromCtx(ctx)
	payload = payload.MergeWhere(false, dbflex.Eq("UserID", meID)).SetSort("TrxDate")
	trxs, _ := datahub.Find(h, new(karamodel.AttendanceTrx), payload)

	i := 0
	res := make([]*HistoryRespond, 0)
	for _, trx := range trxs {
		wl, _ := datahub.GetByID(h, new(karamodel.WorkLocation), trx.WorkLocationID)

		if trx.Op == karamodel.Checkin {
			res = append(res, &HistoryRespond{
				CheckInID:        trx.ID,
				WorkLocationID:   trx.WorkLocationID,
				WorkLocationName: wl.Name,
				CheckIn:          &trx.TrxDate,
			})

			i++
		} else {
			// only to handle first record is checkout
			if len(res) > 0 {
				res[i-1].CheckOutID = trx.ID
				res[i-1].CheckOut = &trx.TrxDate
				res[i-1].TotalHour = trx.Hours
			}
		}
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].CheckIn.After(*res[j].CheckIn)
	})

	return res, nil
}

type PostRequest struct {
	ConfirmForReview bool
	Op               string
	WorkLocationID   string
	Time             *time.Time
	Ref1             string
	Photo            string
	Long             float64
	Lat              float64
}

func (obj *MeLogic) Post(ctx *kaos.Context, payload *PostRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", errors.New("missing: db")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		return "", fmt.Errorf("missing: user id")
	}

	dt := time.Now()
	if payload.Time != nil {
		dt = *payload.Time
	}
	trx, err := new(UserTrx).Create(ctx, &karamodel.TrxRequest{
		UserID:           userID,
		Op:               payload.Op,
		WorkLocationID:   payload.WorkLocationID,
		TrxDate:          dt,
		TrxTime:          "",
		Ref1:             payload.Ref1,
		ConfirmForReview: true,
		Long:             payload.Long,
		Lat:              payload.Lat,
	})
	if err != nil {
		return "", err
	}

	if len(payload.Photo) != 0 {
		createUpdatePhoto(h, trx.ID, payload.Photo)
	}

	return trx.ID, nil
}

func (obj *MeLogic) UploadPhoto(ctx *kaos.Context, payload struct {
	ID    string
	Photo string
}) (string, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return "", errors.New("missing: db")
	}

	if e := createUpdatePhoto(db, payload.ID, payload.Photo); e != nil {
		return "", e
	}
	return "OK", nil
}

func createUpdatePhoto(db *datahub.Hub, id string, photoBase64 string) error {
	if id == "" {
		return fmt.Errorf("id is mandatory")
	}

	if e := db.Save(&karamodel.AttendancePhoto{
		ID:    id,
		Photo: photoBase64,
	}); e != nil {
		return e
	}

	return nil
}

func decodeBase64(content string) ([]byte, error) {
	bs, e := base64.StdEncoding.DecodeString(content)
	if e != nil {
		return nil, fmt.Errorf("fail to decode content. %s", e.Error())
	}
	return bs, nil
}

type AttendaceSummaryRequest struct {
	Month string
}

func (obj *MeLogic) AttendaceSummary(ctx *kaos.Context, payload AttendaceSummaryRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	start, err := time.Parse("200601", payload.Month)
	if err != nil {
		return nil, fmt.Errorf("error when convert month payload: %s", err.Error())
	}

	var end time.Time
	// check if filter is current month
	if start.Month() == time.Now().Month() {
		end = time.Now()
	} else {
		end = start.AddDate(0, 1, 0).Add(-1 * time.Second)
	}

	meID := sebar.GetUserIDFromCtx(ctx)
	param := dbflex.NewQueryParam().SetWhere(dbflex.And(
		dbflex.Eq("UserID", meID),
		dbflex.Gte("TrxDate", start),
		dbflex.Lte("TrxDate", end),
		dbflex.Eq("Op", karamodel.Checkin),
	))
	attendace, err := datahub.Count(h, new(karamodel.AttendanceTrx), param)
	if err != nil {
		return nil, fmt.Errorf("error when get attendace: %s", err.Error())
	}

	absence := end.Day() - attendace
	return codekit.M{"Absence": absence, "TotalAttendance": attendace}, nil
}
