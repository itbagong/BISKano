package afycoreconfig

import "git.kanosolution.net/kano/kaos"

type ModConfig struct {
	Ev kaos.EventHub
}

var (
	Config = &ModConfig{}
)
