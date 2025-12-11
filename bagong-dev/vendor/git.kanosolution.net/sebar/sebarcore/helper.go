package sebarcore

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/kmsg"
	"github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/sebarcode/codekit"
)

func SendEmail(ctx *kaos.Context, to, kind, langID string, data codekit.M) (string, error) {
	sendEmailTopic := ctx.Data().Get("service_topic_send_message_template", "").(string)
	if sendEmailTopic == "" {
		return "", errors.New("missing_config: service_topic_send_message_template")
	}
	msg := kmsg.Message{
		Kind:   kind,
		To:     to,
		Method: "SMTP",
	}
	ev, err := ctx.DefaultEvent()
	if err != nil {
		return "", errors.New("nil: EventHub")
	}
	sendMessageRequest := ksmsg.SendTemplateRequest{
		TemplateName: kind,
		Message:      &msg,
		LanguageID:   langID,
		Data:         data,
	}
	msgID := ""
	err = ev.Publish(sendEmailTopic, sendMessageRequest, &msgID, &kaos.PublishOpts{
		Headers: codekit.M{},
	})
	return msgID, err
}
