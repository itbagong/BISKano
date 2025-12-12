package ficologic

import (
	"errors"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

func RegisterLab(s *kaos.Service) error {
	s.RegisterModel(new(LabLogic), "lab")
	return nil
}

type LabLogic struct {
}

type ClaimNosRequest struct {
	Num int
}

func (obj *LabLogic) ClaimNos(ctx *kaos.Context, payload *ClaimNosRequest) (codekit.M, error) {
	res := codekit.M{}

	if ficoconfig.Config.EventHub == nil {
		return nil, errors.New("missing: event hub")
	}

	if payload.Num <= 0 {
		return nil, errors.New("num <= 0")
	}

	wg := new(sync.WaitGroup)
	wg.Add(payload.Num)
	codes := []string{}
	fail := 0
	mtx := new(sync.Mutex)
	for i := 0; i < payload.Num; i++ {
		go func(wg *sync.WaitGroup, codes *[]string, mtx *sync.Mutex) {
			defer wg.Done()

			resp := new(tenantcorelogic.NumSeqClaimRespond)
			number := ""
			if e := ficoconfig.Config.EventHub.Publish(ficoconfig.Config.SeqNumTopic,
				&tenantcorelogic.NumSeqMapping{Mapping: "Vendor", CompanyID: "", Date: time.Now()}, resp, &kaos.PublishOpts{Timeout: 60 * time.Second}); e == nil {
				number = resp.Number
			} else {
				fail++
				ctx.Log().Warningf("get code error: %s", e.Error())
			}
			if number != "" {
				mtx.Lock()
				newCodes := *codes
				newCodes = append(newCodes, number)
				*codes = newCodes
				mtx.Unlock()
			}
		}(wg, &codes, mtx)
	}

	wg.Wait()
	res.Set("Fail", fail).Set("Success", payload.Num-fail).Set("Codes", codes)
	return res, nil
}

type LedgerBalanceRequest struct {
	Opt LedgerBalanceGetOpts
}

func (obj *LabLogic) GetLedgerBalance(ctx *kaos.Context, payload string) ([]*ficomodel.LedgerBalance, error) {
	dt01Des := time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	//dt17Des := time.Date(2023, 12, 17, 0, 0, 0, 0, time.UTC)
	db := sebar.GetTenantDBFromContext(ctx)
	bal := NewLedgerBalanceHub(db)

	/*
		_, err := bal.Sync(&dt17Des, LedgerBalanceOpt{
			CompanyID:        "DEMO00",
			AccountIDs:       []string{"626005"},
			GroupByDimension: []string{"Site"},
		})
		if err != nil {
			return nil, err
		} */

	bals, err := bal.Get(&dt01Des, LedgerBalanceOpt{
		CompanyID:        "DEMO00",
		AccountIDs:       []string{"626005"},
		GroupByDimension: []string{"Site"},
	})
	return bals, err
}

type GetCustomerBalanceReq struct {
	BalanceDate *time.Time
	Opts        CustomerBalanceOpt
}

func (obj *LabLogic) GetCustomerBalance(ctx *kaos.Context, payload *GetCustomerBalanceReq) ([]*ficomodel.CustomerBalance, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	bal := NewCustomerBalanceHub(db)

	bals, err := bal.Get(payload.BalanceDate, payload.Opts)
	return bals, err
}

type CashBalanceRequest struct {
	BalanceDate *time.Time
	Opts        CashBalanceOpt
}

func (obj *LabLogic) GetCashBalance(ctx *kaos.Context, payload *CashBalanceRequest) ([]*ficomodel.CashBalance, error) {
	tenantID := getTenantID(ctx)
	db := sebar.GetTenantDB(ctx, tenantID)
	bal := NewCashBalanceHub(db)

	bals, err := bal.Get(payload.BalanceDate, payload.Opts)
	return bals, err
}

type GetVendorBalanceRequest struct {
	BalanceDate *time.Time
	Opts        VendorBalanceOpt
}

func (obj *LabLogic) GetVendorBalance(ctx *kaos.Context, payload *GetVendorBalanceRequest) ([]*ficomodel.VendorBalance, error) {
	tenantID := getTenantID(ctx)
	db := sebar.GetTenantDB(ctx, tenantID)
	bal := NewVendorBalanceHub(db)

	bals, err := bal.Get(payload.BalanceDate, payload.Opts)
	return bals, err
}

func (l *LabLogic) SyncCustomerBalance(ctx *kaos.Context, payload *CustomerBalanceOpt) ([]*ficomodel.CustomerBalance, error) {
	db := sebar.GetTenantDB(ctx, "Demo")
	bal := NewCustomerBalanceHub(db)
	if payload == nil {
		payload = &CustomerBalanceOpt{
			CompanyID:  "DEMO00",
			AccountIDs: []string{},
		}
	}
	return bal.Sync(nil, *payload)
}

func (l *LabLogic) SyncCashBalance(ctx *kaos.Context, payload *CashBalanceOpt) ([]*ficomodel.CashBalance, error) {
	db := sebar.GetTenantDB(ctx, "Demo")
	bal := NewCashBalanceHub(db)
	if payload == nil {
		payload = &CashBalanceOpt{
			CompanyID:  "DEMO00",
			AccountIDs: []string{},
		}
	}
	return bal.Sync(nil, *payload)
}

func getTenantID(ctx *kaos.Context) string {
	jwtdata := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	tenantID := jwtdata.GetString("TenantID")
	if tenantID == "" {
		tenantID = "Demo"
	}
	return tenantID
}

func (obj *LabLogic) RecalcCashSchedule(ctx *kaos.Context, payload string) ([]*ficomodel.CashSchedule, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: database")
	}

	css, _ := datahub.Find(db, new(ficomodel.CashSchedule), dbflex.NewQueryParam().SetWhere(dbflex.Ne("Settled", 0)))
	css = recalcSched(db, css...)

	return css, nil
}
