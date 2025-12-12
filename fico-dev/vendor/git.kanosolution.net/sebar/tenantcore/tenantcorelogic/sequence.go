package tenantcorelogic

import (
	"errors"
	"fmt"
	"strings"
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

	setup := GetSequenceSetup(db, payload.Mapping, payload.CompanyID)
	if setup == nil {
		return nil, fmt.Errorf("missing: %s %s", payload.Mapping, payload.CompanyID)
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
		return nil, fmt.Errorf("missing: number sequence: %s", payload.NumberSequenceID)
	}
	result.Number = seq.Format(seq.LastNo+1, &payload.Date)
	seq.LastNo += 1
	go hub.Save(seq)
	return result, nil
	// err = hub.Save(seq)
	// fmt.Println("Claim error:", err.Error())
	// return result, err
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
	mapIDPrefix := map[string]string{
		"PurchaseRequest": "PR",
		"ItemRequest":     "IR",
	}

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
		userID := sebar.GetUserIDFromCtx(ctx)

		getIDFromSequence := func() (string, error) {
			setup := GetSequenceSetup(h, sequenceKind, coID)
			if setup == nil || setup.NumSeqID == "" {
				return "", nil
			}

			ev, _ := ctx.DefaultEvent()
			if ev == nil {
				return "", nil
			}

			resp := new(NumSeqClaimRespond)
			if e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, &kaos.PublishOpts{Headers: codekit.M{
				"CompanyID": coID, sebar.CtxJWTReferenceID: userID}}); e != nil {
				return "", e
			}

			return resp.Number, nil
		}

		sequenceID, err := getIDFromSequence()
		if err != nil {
			return true, err
		}

		if prefix, ok := mapIDPrefix[sequenceKind]; ok {
			if sequenceID == "" || strings.Contains(sequenceID, prefix) == false {
				retry := 5
				for i := 0; i < retry; i++ {
					sequenceID, err = getIDFromSequence()
					if err != nil {
						return true, err
					}

					if sequenceID != "" && strings.Contains(sequenceID, prefix) {
						break
					}
				}
			}
		}

		m.Set(field, sequenceID)
		serde.Serde(m, payload)

		return true, nil
	}
}

type NumSeqClaimPayloadV2 struct {
	CompanyID        string
	UserID           string
	Text             string
	NumberSequenceID string
	Date             time.Time
	CustomSequence   []string
	Dimension        tenantcoremodel.Dimension
}

func mapSequence(ctx *kaos.Context, payload NumSeqClaimPayloadV2) (result string, err error) {
	mapFormat := map[string]string{}
	totalCustom := 0

	// split by any delimiter
	words := splitAny(payload.Text, "${}")

	// get total custom format
	for _, c := range words {
		fSeq, _, ok := hasFormat(c)
		if ok {
			if fSeq != "counter" && fSeq != "dt" && fSeq != "dim" {
				totalCustom++
			}
		}
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return result, errors.New("nil: EventHub")
	}

	// map sequence and compare total custom sequence
	custSeq := 0
	for _, c := range words {
		fSeq, vSeq, ok := hasFormat(c)
		if ok {
			key := "${" + c + "}"
			switch fSeq {
			case "counter":
				resp := new(NumSeqClaimRespond)
				if e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: payload.NumberSequenceID, Date: time.Now()}, resp, &kaos.PublishOpts{Headers: codekit.M{
					"CompanyID": payload.CompanyID, sebar.CtxJWTReferenceID: payload.UserID}}); e != nil {
					return result, e
				}
				numbers := splitAny(resp.Number, "${}")
				for _, n := range numbers {
					arrSplit := strings.Split(n, ":")
					if arrSplit[0] == "counter" {
						vSeq = arrSplit[1]
						break
					}
				}
				mapFormat[key] = vSeq
			case "dt":
				mapFormat[key] = time.Now().Format(vSeq)
			case "dim":
				if vSeq == "site" {
					param := struct {
						ID string
					}{
						ID: payload.Dimension.Get("Site"),
					}
					res := tenantcoremodel.SiteTenantBagong{}
					e := ev.Publish("/v1/bagong/sitesetup/get-site-by-id", &param, &res, nil)
					if e != nil {
						return result, e
					}
					if res.Alias != "" {
						mapFormat[key] = res.Alias
					}
				} else {
					mapFormat[key] = ""
				}
			default:
				if totalCustom == len(payload.CustomSequence) {
					mapFormat[key] = fmt.Sprintf(vSeq, payload.CustomSequence[custSeq])
					custSeq++
				} else {
					return result, fmt.Errorf("custom Sequence not match %s", key)
				}
			}
		} else {
			mapFormat[c] = c
		}
	}

	// replace sequence from map key
	for i, c := range mapFormat {
		payload.Text = strings.Replace(payload.Text, i, c, -1)
	}

	result = payload.Text
	return
}

func splitAny(s string, seps string) []string {
	splitter := func(r rune) bool {
		return strings.ContainsRune(seps, r)
	}
	return strings.FieldsFunc(s, splitter)
}

func hasFormat(v string) (sformat, vformat string, ok bool) {
	if strings.Contains(v, ":") {
		arrSplit := strings.Split(v, ":")
		return arrSplit[0], arrSplit[1], true
	}
	return v, v, false
}

// MWPreAssignCustomSequenceNo middleware untuk menggunakan numseq, bisa juga dipanggil langsung kalau ndak mau ribet masang mw-nya
func MWPreAssignCustomSequenceNo(sequenceKind string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		customSeq := []string{}
		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		id, _ := m["_id"].(string)
		if id != "" {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		if h == nil {
			return false, errors.New("missing: db connection")
		}

		var trxType string
		if _, ok := m.Get("TransactionType").(string); ok {
			trxType = m.Get("TransactionType").(string)
		}
		if _, ok := m.Get("CashJournalType").(string); ok {
			trxType = m.Get("CashJournalType").(string)
		}
		if trxType == "Mining Invoice - Rent" {
			sequenceKind = "CustomerJournalInvoiceMiningRent"

			var actProject, actTimesheet string

			if _, ok := m.Get("References").(tenantcoremodel.References); ok {
				for _, c := range m.Get("References").(tenantcoremodel.References) {
					if c.Key == "Action - Bulan Timesheet End" {
						timeParse, _ := time.Parse("2006-01-02", c.Value.(string))
						actTimesheet = timeParse.Format("01")

					} else if c.Key == "Action - Project" {
						actProject = strings.Split(c.Value.(string), "|")[0]
					}
				}
			}

			customer := new(tenantcoremodel.Customer)
			if e := h.GetByID(customer, m.Get("CustomerID").(string)); e != nil {
				ctx.Log().Errorf("Failed populate data customer: %s", e.Error())
				return false, errors.New("error: get customer")
			}
			customSeq = append(customSeq, customer.CustomerAlias)
			customSeq = append(customSeq, actProject)
			customSeq = append(customSeq, actTimesheet)
		} else if trxType == "Employee Expense" {
			sequenceKind = "EmployeeExpense"
		} else if trxType == "General Invoice - Tourism" {
			sequenceKind = "CustomerJournalInvoiceTourism"
		} else if trxType == "BTS Invoice" {
			sequenceKind = "CustomerJournalInvoiceBTS"
		} else if trxType == "Trayek Invoice" {
			sequenceKind = "CustomerJournalInvoiceTrayek"
		} else if trxType == "General Invoice - Sparepart" {
			sequenceKind = "CustomerJournalInvoiceSparepart"
		} else if trxType == "General Invoice" {
			sequenceKind = "CustomerJournal"
		} else if trxType == "CASH IN" {
			sequenceKind = "CashIn"
			var actType string
			if _, ok := m.Get("References").(tenantcoremodel.References); ok {
				for _, c := range m.Get("References").(tenantcoremodel.References) {
					if c.Key == "Submission Type" {
						actType = c.Value.(string)
					}
				}
			}
			if actType == "Petty Cash" {
				sequenceKind = "PettyCashSubmission"
			}
		} else if trxType == "CASH OUT" {
			sequenceKind = "CashOut"
		} else if trxType == "SUBMISSION CASH IN" {
			sequenceKind = "SubmissionCashIn"
		} else if trxType == "SUBMISSION CASH OUT" {
			sequenceKind = "SubmissionCashOut"
		}

		coID := GetCompanyIDFromContext(ctx)
		if coID == "" {
			return false, errors.New("error: missing company, please relogin")
		}

		userID := sebar.GetUserIDFromCtx(ctx)
		if userID == "" {
			return false, errors.New("error: missing user, please relogin")
		}

		setup := GetSequenceSetup(h, sequenceKind, coID)
		if setup == nil || setup.NumSeqID == "" {
			return true, nil
		}

		seq := &tenantcoremodel.NumberSequence{}
		err := h.GetByID(seq, setup.ID)
		if err != nil {
			return false, fmt.Errorf("missing: number sequence: %s", setup.ID)
		}

		mapSeq, err := mapSequence(ctx, NumSeqClaimPayloadV2{
			Dimension:        m.Get("Dimension").(tenantcoremodel.Dimension),
			CompanyID:        coID,
			UserID:           userID,
			Text:             seq.OutFormat,
			NumberSequenceID: seq.ID,
			Date:             time.Now(),
			CustomSequence:   customSeq,
		})
		if err != nil {
			return false, fmt.Errorf("error: get map seq %s", err)
		}

		m.Set("_id", mapSeq)
		serde.Serde(m, payload)

		return true, nil
	}
}

func GenerateIDFromNumSeq(ctx *kaos.Context, sequenceKind string) (string, error) {
	coID := GetCompanyIDFromContext(ctx)
	userID := sebar.GetUserIDFromCtx(ctx)
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", errors.New("missing: db connection")
	}

	setup := GetSequenceSetup(h, sequenceKind, coID)
	if setup == nil || setup.NumSeqID == "" {
		return "", nil
	}

	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", nil
	}

	resp := new(NumSeqClaimRespond)
	if e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, &kaos.PublishOpts{Headers: codekit.M{
		"CompanyID": coID, sebar.CtxJWTReferenceID: userID}}); e != nil {
		return "", e
	}

	return resp.Number, nil
}
