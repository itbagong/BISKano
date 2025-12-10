package xibarconfig

import "git.kanosolution.net/kano/kaos"

type ModuleConfig struct {
	DefaultTenantID string `json:"default_tenant_id"`
	Ev              kaos.EventHub
}

var Config = new(ModuleConfig)
