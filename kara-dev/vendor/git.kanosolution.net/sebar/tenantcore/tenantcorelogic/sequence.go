package tenantcorelogic

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

type SequenceHandler struct {
	Mutexes MutexMap
}
type NumSeqMapping struct {
	Mapping   string
	CompanyID string
	Date      time.Time
}
type NumSeqClaimPayload struct {
	NumberSequenceID string
	Date             time.Time
}
type MutexMap map[string]*sync.Mutex

func NewSequence() *SequenceHandler {
	res := &SequenceHandler{Mutexes: MutexMap{}}

	return res
}

var defaultSequence *SequenceHandler

func DefaultSequence() *SequenceHandler {
	if defaultSequence == nil {
		defaultSequence = NewSequence()
	}
	return defaultSequence
}

type NumSeqClaimRespond struct {
	Number string
}

func (u *SequenceHandler) ClaimBySetup(ctx *kaos.Context, payload *NumSeqMapping) (*NumSeqClaimRespond, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: db")
	}

	setup := GetSequenceSetup(db, "LedgerVoucher", payload.CompanyID)
	if setup == nil {
		return nil, fmt.Errorf("missing: LedgerVoucherNo")
	}

	resp, err := u.Claim(ctx, &NumSeqClaimPayload{
		NumberSequenceID: setup.NumSeqID,
		Date:             payload.Date,
	})
	return resp, err
}

func (u *SequenceHandler) Claim(ctx *kaos.Context, payload *NumSeqClaimPayload) (*NumSeqClaimRespond, error) {
	mutex, ok := u.Mutexes[payload.NumberSequenceID]
	if !ok {
		u.Mutexes[payload.NumberSequenceID] = &sync.Mutex{}
		mutex = u.Mutexes[payload.NumberSequenceID]
	}
	mutex.Lock()
	defer mutex.Unlock()

	result := new(NumSeqClaimRespond)

	hub := sebar.GetTenantDBFromContext(ctx)

	seq := &tenantcoremodel.NumberSequence{}
	err := hub.GetByID(seq, payload.NumberSequenceID)
	if err != nil {
		return nil, err
	}
	result.Number = seq.Format(seq.LastNo+1, &payload.Date)
	seq.LastNo += 1
	hub.Save(seq)
	return result, nil
}

func GetSequenceSetup(h *datahub.Hub, kind, companyid string) *tenantcoremodel.NumberSequenceSetup {
	setup := new(tenantcoremodel.NumberSequenceSetup)
	if e := h.GetByFilter(setup, dbflex.Eqs("Kind", kind, "CompanyID", companyid)); e != nil {
		if e = h.GetByFilter(setup, dbflex.Eqs("Kind", kind, "CompanyID", "")); e != nil {
			return nil
		}
	}

	return setup
}

// MWPreAssignSequenceNo middleware untuk menggunakan numseq, bisa juga dipanggil langsung kalau ndak mau ribet masang mw-nya
func MWPreAssignSequenceNo(sequenceKind string, useCompanyID bool, field string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if field == "" {
			field = "_id"
		}

		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		id, _ := m[field].(string)
		if id != "" {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		if h == nil {
			return false, errors.New("missing: db connection")
		}

		coID := GetCompanyIDFromContext(ctx)
		setup := GetSequenceSetup(h, sequenceKind, coID)
		if setup == nil || setup.NumSeqID == "" {
			return true, nil
		}

		ev, _ := ctx.DefaultEvent()
		if ev == nil {
			return true, nil
		}

		resp := new(NumSeqClaimRespond)
		if e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, nil); e != nil {
			return true, e
		}
		m.Set(field, resp.Number)
		serde.Serde(m, payload)

		return true, nil
	}
}
