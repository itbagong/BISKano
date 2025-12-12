package ficoconfig

import (
	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/logger"
)

type ModuleConfig struct {
	PostingTopic             string `json:"posting_topic"`
	SeqNumTopic              string `json:"seqnum_topic"`
	AssetEndPoint            string
	AssetRegion              string
	AssetBucket              string
	AssetKey                 string
	AssetSecret              string
	AddrUserFind             string
	AddrUserAccess           string
	FinancialPeriodModules   []string
	AddrAuthValidation       string `json:"addr_auth_validation"`
	AddrAccessValidation     string `json:"addr_access_validation"`
	TopicSendMessageTemplate string `json:"topic_send_message_template"`
	AddrWebTenant            string `json:"addr_web_tenant"`
	EventHub                 kaos.EventHub
	Log                      *logger.LogEngine
}

var (
	Config = NewModuleConfig()
)

func NewModuleConfig() *ModuleConfig {
	c := new(ModuleConfig)
	c.Log = logger.NewLogEngine(true, false, "", "", "")
	return c
}
