package ficologic

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type RecalcHandler struct {
}

type CalcRespond struct {
	Done map[string]string
	Fail map[string]string
}

type FixByTypeRequest struct {
	Worker      int
	TenantID    string
	CompanyID   string
	JournalType string
	JournalIDs  []string
}

func (obj *RecalcHandler) FixVendor(ctx *kaos.Context, payload *FixByTypeRequest) (*[]ficomodel.VendorJournal, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	vndJournal := []ficomodel.VendorJournal{}
	if e := db.GetsByFilter(new(ficomodel.VendorJournal), dbflex.Eq("JournalTypeID", "VJTOpeningBalance"), &vndJournal); e != nil {
		return nil, e
	}

	for _, c := range vndJournal {
		if c.Dimension.Get("CC") == "" {
			c.Dimension.Set("CC", "FAT")
			lines := []ficomodel.JournalLine{}
			for _, line := range c.Lines {
				tmp := line
				tmp.Dimension = c.Dimension
				lines = append(lines, tmp)
			}
			c.Lines = lines
			if e := db.Save(&c); e != nil {
				return nil, fmt.Errorf("error : save vendorJournal: %s", e)
			}
		}

	}

	return nil, nil
}

func (obj *RecalcHandler) FixCustomer(ctx *kaos.Context, payload *FixByTypeRequest) (*[]ficomodel.CustomerJournal, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	custJournal := []ficomodel.CustomerJournal{}
	if e := db.GetsByFilter(new(ficomodel.CustomerJournal), dbflex.Eq("JournalTypeID", "JTOpeningBalance"), &custJournal); e != nil {
		return nil, e
	}

	for _, c := range custJournal {
		if c.Dimension.Get("CC") == "" {
			c.Dimension.Set("CC", "FAT")
			lines := []ficomodel.JournalLine{}
			for _, line := range c.Lines {
				tmp := line
				tmp.Dimension = c.Dimension
				lines = append(lines, tmp)
			}
			c.Lines = lines
			if e := db.Save(&c); e != nil {
				return nil, fmt.Errorf("error : save vendorJournal: %s", e)
			}
		}

	}

	return nil, nil
}

func (obj *RecalcHandler) FixByType(ctx *kaos.Context, payload *FixByTypeRequest) (*CalcRespond, error) {
	db := sebar.GetTenantDB(ctx, payload.TenantID)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	resp := &CalcRespond{
		Done: make(map[string]string),
		Fail: make(map[string]string),
	}

	if payload.Worker == 0 {
		payload.Worker = 10
	}

	ids := make(chan string, payload.Worker)
	count := 0
	processed := 0

	mx := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	wg.Add(payload.Worker)

	for i := 0; i < payload.Worker; i++ {
		go func(wg *sync.WaitGroup, mx *sync.Mutex, resp *CalcRespond, count, processed *int) {
			defer wg.Done()

			for id := range ids {
				mx.Lock()
				*processed++
				ctx.Log().Infof("fix journal processing %d of %d", *processed, *count)

				logTest := ficomodel.LogTest{
					Text:    fmt.Sprintf("fix journal processing %d of %d", *processed, *count),
					Created: time.Now(),
				}
				db.Save(&logTest)

				mx.Unlock()

				req := PostRequest{
					JournalType: tenantcoremodel.TrxModule(payload.JournalType),
					JournalID:   id,
				}
				reqOpt := PostingHubCreateOpt{
					Db:        db,
					UserID:    "system",
					CompanyID: payload.CompanyID,
					ModuleID:  payload.CompanyID,
					JournalID: id,
				}

				ph, err := createPostingEngine(req, reqOpt)
				if err != nil {
					mx.Lock()
					resp.Fail[id] = err.Error()
					mx.Unlock()
					continue
				}

				vch, err := ph.Recalc(db, payload.CompanyID, &PostRequest{
					JournalType: tenantcoremodel.TrxModule(payload.JournalType),
					JournalID:   id,
				})
				if err != nil {
					mx.Lock()
					resp.Fail[id] = err.Error()
					mx.Unlock()
					continue
				}

				mx.Lock()
				resp.Done[id] = vch
				mx.Unlock()
			}
		}(wg, mx, resp, &count, &processed)
	}

	switch payload.JournalType {
	case string(ficomodel.SubledgerVendor):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.VendorJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerCustomer):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.CustomerJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerAccounting):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.LedgerJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerCashBank):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.CashJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)
	}

	wg.Wait()

	return resp, nil
}

func (obj *RecalcHandler) FixByTypeEv(ctx *kaos.Context, payload *FixByTypeRequest) (*CalcRespond, error) {
	// return nil, nil
	db := sebar.GetTenantDB(ctx, payload.TenantID)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	resp := &CalcRespond{
		Done: make(map[string]string),
		Fail: make(map[string]string),
	}

	if payload.Worker == 0 {
		payload.Worker = 10
	}

	ids := make(chan string, payload.Worker)
	count := 0
	processed := 0

	mx := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	wg.Add(payload.Worker)

	for i := 0; i < payload.Worker; i++ {
		go func(wg *sync.WaitGroup, mx *sync.Mutex, resp *CalcRespond, count, processed *int) {
			defer wg.Done()

			for id := range ids {
				mx.Lock()
				*processed++
				ctx.Log().Infof("fix journal processing %d of %d", *processed, *count)

				logTest := ficomodel.LogTest{
					Text:    fmt.Sprintf("fix journal processing %d of %d", *processed, *count),
					Created: time.Now(),
				}
				db.Save(&logTest)

				mx.Unlock()

				req := PostRequest{
					JournalType: tenantcoremodel.TrxModule(payload.JournalType),
					JournalID:   id,
				}
				reqOpt := PostingHubCreateOpt{
					Db:        db,
					UserID:    "system",
					CompanyID: payload.CompanyID,
					ModuleID:  payload.CompanyID,
					JournalID: id,
				}

				ph, err := createPostingEngine(req, reqOpt)
				if err != nil {
					mx.Lock()
					resp.Fail[id] = err.Error()
					mx.Unlock()
					continue
				}

				vch, err := ph.Recalc(db, payload.CompanyID, &PostRequest{
					JournalType: tenantcoremodel.TrxModule(payload.JournalType),
					JournalID:   id,
				})
				if err != nil {
					mx.Lock()
					resp.Fail[id] = err.Error()
					mx.Unlock()
					continue
				}

				mx.Lock()
				resp.Done[id] = vch
				mx.Unlock()
			}
		}(wg, mx, resp, &count, &processed)
	}

	switch payload.JournalType {
	case string(ficomodel.SubledgerVendor):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.VendorJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerCustomer):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.CustomerJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerAccounting):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.LedgerJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)

	case string(ficomodel.SubledgerCashBank):
		wheres := []*dbflex.Filter{dbflex.Eqs("CompanyID", payload.CompanyID, "Status", ficomodel.JournalStatusPosted)}
		if len(payload.JournalIDs) > 0 {
			wheres = append(wheres, dbflex.In("_id", payload.JournalIDs...))
		}
		js, err := datahub.Find(db, new(ficomodel.CashJournal), dbflex.NewQueryParam().
			SetSelect("_id").
			SetWheres(wheres...))
		if err != nil {
			return nil, err
		}

		count = len(js)
		for _, vj := range js {
			ids <- vj.ID
		}
		close(ids)
	}

	wg.Wait()
	logTest2 := ficomodel.LogTest{
		Text:     fmt.Sprintf("finished"),
		Response: resp,
		Created:  time.Now(),
	}
	db.Save(&logTest2)

	return resp, nil
}

func (obj *PostingHub[H, L]) Recalc(db *datahub.Hub, companyID string, request *PostRequest) (string, error) {
	res := ""

	_, err := obj.Header()
	if err != nil {
		return res, err
	}

	_, err = obj.Lines()
	if err != nil {
		return res, err
	}

	//res, _, _, err = obj.provider.Calculate(obj.opt, obj.header, obj.lines)
	err = obj.Calculate()
	if err != nil {
		return res, fmt.Errorf("calculate: %s", err.Error())
	}

	/* clear data */
	filter := dbflex.Eqs("CompanyID", companyID, "SourceType", request.JournalType, "SourceJournalID", request.JournalID)

	db.DeleteByFilter(new(ficomodel.TaxTransaction), dbflex.And(filter))
	db.DeleteByFilter(new(ficomodel.CashTransaction), dbflex.And(filter))
	db.DeleteByFilter(new(ficomodel.VendorTransaction), dbflex.And(filter))
	db.DeleteByFilter(new(ficomodel.CustomerTransaction), dbflex.And(filter))
	db.DeleteByFilter(new(ficomodel.LedgerTransaction), dbflex.And(filter))

	res, err = obj.PostJournal()
	if err != nil {
		return res, fmt.Errorf("posting: %s", err.Error())
	}

	return res, nil
}
