package ficologic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
)

type ApplyHandler struct {
}

type GetCashBankRequest struct {
	Account string
	ID      string
}

func (m *ApplyHandler) GetCashBank(ctx *kaos.Context, p *GetCashBankRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	match := bson.M{
		"SourceType": ficomodel.SubledgerCashBank,
		"$or": bson.A{
			bson.M{"Oustanding": bson.M{"$lt": 0}},
			bson.M{"Oustanding": bson.M{"$gt": 0}},
		},
	}
	if p.Account != "" {
		match["Account.AccountType"] = p.Account
	}

	if p.ID != "" {
		match["Account.AccountID"] = p.ID
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	cashes := []ficomodel.CashSchedule{}
	cmd := dbflex.From(new(ficomodel.CashSchedule).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &cashes); err != nil {
		return nil, fmt.Errorf("err when get cash schedule: %s", err.Error())
	}

	return cashes, nil
}

type GetInvoiceResponse struct {
	SourceRecID string
	ficomodel.CashSchedule
	Adjustments []*ficomodel.CashAdjustment
}

func (m *ApplyHandler) GetInvoice(ctx *kaos.Context, p *GetCashBankRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	match := bson.M{
		"SourceType": bson.M{"$ne": ficomodel.SubledgerCashBank},
		"$or": bson.A{
			bson.M{"Oustanding": bson.M{"$lt": 0}},
			bson.M{"Oustanding": bson.M{"$gt": 0}},
		},
	}
	if p.Account != "" {
		if strings.ToLower(p.Account) == string(ficomodel.SubledgerAccounting) {
			p.Account = "EXPENSE"
		}
		match["Account.AccountType"] = p.Account
	}

	if p.ID != "" {
		match["Account.AccountID"] = p.ID
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
	}
	cashes := []ficomodel.CashSchedule{}
	cmd := dbflex.From(new(ficomodel.CashSchedule).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &cashes); err != nil {
		return nil, fmt.Errorf("err when get cash schedule: %s", err.Error())
	}

	ids := lo.Map(cashes, func(m ficomodel.CashSchedule, index int) string {
		return m.ID
	})

	// get adjustment from cash apply
	cashApplies := []ficomodel.CashApply{}
	err := h.Gets(new(ficomodel.CashApply), dbflex.NewQueryParam().SetWhere(
		dbflex.In("ApplyTo.RecordID", ids...),
	), &cashApplies)
	if err != nil {
		return nil, fmt.Errorf("err when get cash apply: %s", err.Error())
	}

	mapCashApply := lo.Associate(cashApplies, func(cash ficomodel.CashApply) (string, ficomodel.CashApply) {
		return cash.ApplyTo.RecordID, cash
	})

	result := make([]GetInvoiceResponse, len(cashes))
	for i, c := range cashes {
		res := GetInvoiceResponse{
			CashSchedule: c,
		}
		if v, ok := mapCashApply[c.ID]; ok {
			res.SourceRecID = v.Source.RecordID
			res.Adjustments = v.Adjustment
		}

		result[i] = res
	}

	return result, nil
}

func (m *ApplyHandler) Apply(ctx *kaos.Context, p []CashApplyRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	for _, ca := range p {
		ca.Db = h

		err := ApplyCashShedule(ca)
		if err != nil {
			return "", fmt.Errorf("error when processing apply: %s", err.Error())
		}
	}

	return "success", nil
}

type GetInvoiceCashJournalRequest struct {
	Type           string // in or out
	CashScheduleID string
}

// GetInvoiceCashJournal get invoice for cash in and cash out
func (m *ApplyHandler) GetInvoiceCashJournal(ctx *kaos.Context, p GetInvoiceCashJournalRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	match := bson.M{
		"SourceType": bson.M{"$ne": ficomodel.SubledgerCashBank},
		"$or": bson.A{
			bson.M{"Oustanding": bson.M{"$lt": 0}},
			bson.M{"Oustanding": bson.M{"$gt": 0}},
		},
		"Account.AccountID": p.CashScheduleID,
	}

	if p.Type == "in" {
		match["Account.AccountType"] = "CUSTOMER"
	} else {
		match["Account.AccountType"] = bson.M{"$in": []string{"VENDOR", "EXPENSE"}}
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
	}

	cashes := []ficomodel.CashSchedule{}
	cmd := dbflex.From(new(ficomodel.CashSchedule).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &cashes); err != nil {
		return nil, fmt.Errorf("err when get cash schedule: %s", err.Error())
	}

	return cashes, nil
}

type CashJournalApplyRequest struct {
	CashJournalID string
	Applies       []ficomodel.CashApply
}

// CashApply apply for cash in and cash
func (m *ApplyHandler) ApplyCashJournal(ctx *kaos.Context, payloads []CashJournalApplyRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	models := []orm.DataModel{}
	for _, p := range payloads {
		journal, err := datahub.GetByID(h, new(ficomodel.CashJournal), p.CashJournalID)
		if err != nil {
			return nil, fmt.Errorf("error when get journal: %s", err.Error())
		}

		for _, app := range p.Applies {
			cashSchedule, err := datahub.GetByID(h, new(ficomodel.CashSchedule), app.ApplyTo.RecordID)
			if err != nil {
				return nil, fmt.Errorf("error when get cash schedule: %s", err.Error())
			}
			cashSchedule.Outstanding -= app.Amount
			cashSchedule.Settled += app.Amount

			models = append(models, cashSchedule)

			// update journal line
			for i, l := range journal.Lines {
				if l.LineNo == app.Source.LineNo {
					journal.Lines[i].Amount -= app.Amount
				}
			}

			/*
				for i := range app.Adjustment {
					//app.Adjustment[i].TrxType = "ApplyAdjustment"
				}
			*/

			models = append(models, &app)
		}

		models = append(models, journal)
	}

	for _, m := range models {
		if err := h.Save(m); err != nil {
			return nil, fmt.Errorf("save %s error: %s", m.TableName(), err.Error())
		}
	}

	return "success", nil
}
