package tenantcoreconfig

import "git.kanosolution.net/kano/kaos"

type ModConfig struct {
	DefaultCompanyID         string `json:"default_company_id"`
	StorageEndPoint          string `json:"storage_end_point"`
	StorageRegion            string `json:"storage_region"`
	StorageBucket            string `json:"storage_bucket"`
	StorageKey               string `json:"storage_key"`
	StorageSecret            string `json:"storage_secret"`
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
	PdfExePath               string `json:"pdf_exe_path"`
	PdfTemplatePath          string `json:"pdf_template_path"`

	EventHub kaos.EventHub
}

var (
	Config = new(ModConfig)
)
