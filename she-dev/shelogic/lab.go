package shelogic

import (
	"errors"
	"sync"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
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

	if Config.EventHub == nil {
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
			if e := Config.EventHub.Publish(Config.SeqNumTopic,
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
