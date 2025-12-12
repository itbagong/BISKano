package ficomodel

import (
	"errors"
	"fmt"

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
	fmt.Println("sendEmailTopic", sendEmailTopic)
	fmt.Println("sendMessageRequest", sendMessageRequest)
	msgID := ""
	err := ficoconfig.Config.EventHub.Publish(sendEmailTopic, sendMessageRequest, &msgID, &kaos.PublishOpts{Headers: codekit.M{}})
	fmt.Println("err", err, msgID)
	return msgID, err
}
