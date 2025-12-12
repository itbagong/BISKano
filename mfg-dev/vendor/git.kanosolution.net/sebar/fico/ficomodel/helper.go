package ficomodel

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"github.com/ariefdarmawan/kmsg"
	"github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/sebarcode/codekit"
)

func SendEmailByTemplate(to, kind, langID string, data codekit.M) (string, error) {
	sendEmailTopic := ficoconfig.Config.TopicSendMessageTemplate
	if sendEmailTopic == "" {
		return "", errors.New("missing_config: topic_send_message_template")
	}

	msg := kmsg.Message{
		Kind:   kind,
		To:     to,
		Method: "SMTP",
	}

	sendMessageRequest := ksmsg.SendTemplateRequest{
		TemplateName: kind,
		Message:      &msg,
		LanguageID:   langID,
		Data:         data,
	}

	msgID := ""
	err := ficoconfig.Config.EventHub.Publish(sendEmailTopic, sendMessageRequest, &msgID, &kaos.PublishOpts{Headers: codekit.M{}})
	return msgID, err
}
