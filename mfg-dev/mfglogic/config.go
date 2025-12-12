package mfglogic

import "git.kanosolution.net/kano/kaos"

var Config = new(struct {
	AddrAuthValidation       string `json:"addr_auth_validation"`
	AddrAccessValidation     string `json:"addr_access_validation"`
	TopicAssetwrite          string `json:"topic_asset_write"`
	TopicSendMessageTemplate string `json:"topic_send_message_template"`
	SmtpHost                 string `json:"msg_host"`
	SmtpPort                 int    `json:"smtp_port"`
	SmtpUserID               string `json:"msg_user_id"`
	SmtpPassword             string `json:"msg_password"`
	SmtpUseTLS               string `json:"msg_use_tls"`
	AssetEndPoint            string `json:"asset_end_point"`
	AssetRegion              string `json:"asset_region"`
	AssetBucket              string `json:"asset_bucket"`
	AssetKey                 string `json:"asset_key"`
	AssetSecret              string `json:"asset_secret"`
	AddrWebTenant            string `json:"addr_web_tenant"`

	EventHub kaos.EventHub
})
