package karalogic

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/datahub"
)

type adminEngine struct {
	hub *datahub.Hub
	ev  kaos.EventHub
}

func NewAdminEngine(h *datahub.Hub, ev kaos.EventHub) *adminEngine {
	eng := new(adminEngine)
	eng.hub = h
	eng.ev = ev
	return eng
}
