package karaconfig

import "git.kanosolution.net/kano/kaos"

type ModuleConfig struct {
	GetUserTopic string `json:"get_user_topic"`
	Ev           kaos.EventHub
}

var Config = new(ModuleConfig)
