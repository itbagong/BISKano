package scmconfig

import "git.kanosolution.net/kano/kaos"

type ModConfig struct {
	PostingTopic  string `json:"posting_topic"`
	AddrWebTenant string `json:"addr_web_tenant"`

	ev kaos.EventHub
}

func (m *ModConfig) SetEventHub(ev kaos.EventHub) *ModConfig {
	m.ev = ev
	return m
}

func (m *ModConfig) EventHub() kaos.EventHub {
	return m.ev
}

var Config = new(ModConfig)
