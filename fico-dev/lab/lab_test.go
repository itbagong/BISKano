package lab_test

import (
	"testing"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
)

func TestGeneric(t *testing.T) {
	var h iHandler
	h = newHandler()
	tableName := h.Say()
	if tableName != new(ficomodel.LedgerJournal).TableName() {
		t.Errorf("not same: %s %s", tableName, new(ficomodel.LedgerJournal).TableName())
	}

	h = newCustomerHandler()
	tableName = h.Say()
	if tableName != new(ficomodel.CustomerJournal).TableName() {
		t.Errorf("not same: %s %s", tableName, new(ficomodel.CustomerJournal).TableName())
	}
}

type iHandler interface {
	Say() string
	SubSay() string
}

type handler[T orm.DataModel, O any] struct {
	model T
	this  iHandler
}

func (h *handler[T, O]) Say() string {
	return h.SubSay()
}

func (h *handler[T, O]) SubSay() string {
	return ""
}

type customOpt struct {
	Txt string
}

type handlerLedger struct {
	handler[*ficomodel.LedgerJournal, customOpt]
}

type handlerCustomer struct {
	handler[*ficomodel.CustomerJournal, customOpt]
}

func (h *handlerLedger) Say() string {
	return h.model.TableName()
}

func (h *handlerCustomer) Say() string {
	return h.model.TableName()
}

func newHandler() *handlerLedger {
	h := new(handlerLedger)
	h.model = new(ficomodel.LedgerJournal)
	return h
}

func newCustomerHandler() *handlerCustomer {
	h := new(handlerCustomer)
	h.model = new(ficomodel.CustomerJournal)
	return h
}
